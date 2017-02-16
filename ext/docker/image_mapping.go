package docker

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	// triggers the loading of sqlite3 as a database driver
	sqlite "github.com/mattn/go-sqlite3"
	"github.com/opentable/sous/lib"
	"github.com/opentable/sous/util/docker_registry"
	"github.com/samsalisbury/semv"
)

type (
	// NameCache is a database for looking up SourceIDs based on
	// Docker image names and vice versa.
	NameCache struct {
		RegistryClient     docker_registry.Client
		DB                 *sql.DB
		DockerRegistryHost string
	}

	imageName string

	// NotModifiedErr is returned when an HTTP server returns Not Modified in
	// response to a conditional request
	NotModifiedErr struct{}

	// NoImageNameFound is returned when we cannot find an image name for a
	// given SourceID.
	NoImageNameFound struct {
		sous.SourceID
	}

	// NoSourceIDFound is returned when we cannot find a SourceID for a
	// given image name
	NoSourceIDFound struct {
		imageName
	}
)

// NewBuildArtifact creates a new sous.BuildArtifact representing a Docker
// image.
// XXX this should be removed in favor of sous.NewBuildArtifact
func NewBuildArtifact(imageName string, qstrs strpairs) *sous.BuildArtifact {
	var qs []sous.Quality
	for _, qstr := range qstrs {
		qs = append(qs, sous.Quality{Name: qstr[0], Kind: qstr[1]})
	}

	return &sous.BuildArtifact{Name: imageName, Type: "docker", Qualities: qs}
}

// InMemory configures SQLite to use an in-memory database
// The dummy file allows multiple goroutines see the same in-memory DB
const InMemory = "file:dummy.db?mode=memory&cache=shared"

// InMemoryConnection builds a connection string based on a base name
// This is mostly useful for testing, so that we can have separate cache DBs per test
func InMemoryConnection(base string) string {
	return "file:" + base + "?mode=memory&cache=shared"
}

func (e NoImageNameFound) Error() string {
	return fmt.Sprintf("No image name for %v", e.SourceID)
}

func (e NoSourceIDFound) Error() string {
	return fmt.Sprintf("No source ID for %v", e.imageName)
}

func (e NotModifiedErr) Error() string {
	return "Not modified"
}

// NewNameCache builds a new name cache.
func NewNameCache(drh string, cl docker_registry.Client, db *sql.DB) *NameCache {
	nc := &NameCache{
		RegistryClient:     cl,
		DB:                 db,
		DockerRegistryHost: drh,
	}
	nc.GroomDatabase()
	return nc
}

// ListSourceIDs lists all the known SourceIDs.
func (nc *NameCache) ListSourceIDs() ([]sous.SourceID, error) {
	return nc.dbQueryAllSourceIds()
}

// Warmup warms up the cache.
func (nc *NameCache) Warmup(r string) error {
	ref, err := reference.ParseNamed(r)
	if err != nil {
		return errors.Errorf("%v for %v", err, r)
	}
	ts, err := nc.RegistryClient.AllTags(r)
	if err != nil {
		return errors.Wrap(err, "warming up")
	}
	for _, t := range ts {
		Log.Debug.Printf("Harvested tag: %v for repo: %v", t, r)
		in, err := reference.WithTag(ref, t)
		Log.Vomit.Print(in, err)
		if err == nil {
			a := NewBuildArtifact(in.String(), strpairs{})
			nc.GetSourceID(a) //pull it into the cache...
		}
	}
	return nil
}

// ImageLabels gets the labels for an image name.
func (nc *NameCache) ImageLabels(in string) (map[string]string, error) {
	a := NewBuildArtifact(in, nil)
	sv, err := nc.GetSourceID(a)
	if err != nil {
		return map[string]string{}, errors.Wrapf(err, "Image name: %s", in)
	}

	return Labels(sv), nil
}

// GetArtifact implements sous.Registry.GetArtifact.
func (nc *NameCache) GetArtifact(sid sous.SourceID) (*sous.BuildArtifact, error) {
	name, qls, err := nc.getImageName(sid)
	if err != nil {
		return nil, err
	}
	return NewBuildArtifact(name, qls), nil
}

// GetSourceID looks up the source ID for a given image name.
func (nc *NameCache) GetSourceID(a *sous.BuildArtifact) (sous.SourceID, error) {
	in := a.Name
	var sid sous.SourceID

	Log.Vomit.Printf("Getting source ID for %s", in)

	etag, repo, offset, version, _, err := nc.dbQueryOnName(in)
	if nif, ok := err.(NoSourceIDFound); ok {
		Log.Vomit.Print(nif)
	} else if err != nil {
		Log.Vomit.Print("Err: ", err)
		return sous.SourceID{}, err
	} else {
		Log.Vomit.Printf("Found: %v %v %v %v", repo, offset, version, etag)

		sid, err = sous.NewSourceID(repo, offset, version)
		if err != nil {
			return sid, err
		}
	}

	md, err := nc.RegistryClient.GetImageMetadata(in, etag)
	Log.Vomit.Printf("%+ v %v %T %#v", md, err, err, err)
	if _, ok := err.(NotModifiedErr); ok {
		Log.Debug.Printf("Image name: %s -> Source ID: %v", in, sid)
		return sid, nil
	}
	if err == distribution.ErrManifestNotModified {
		Log.Debug.Printf("Image name: %s -> Source ID: %v", in, sid)
		return sid, nil
	}
	if err != nil {
		return sid, err
	}

	newSID, err := SourceIDFromLabels(md.Labels)
	if err != nil {
		return sid, err
	}

	qualities := qualitiesFromLabels(md.Labels)

	err = nc.dbInsert(newSID, md.Registry+"/"+md.CanonicalName, md.Etag, qualities)
	if err != nil {
		return sid, err
	}

	Log.Vomit.Printf("cn: %v all: %v", md.CanonicalName, md.AllNames)
	names := []string{}
	for _, n := range md.AllNames {
		names = append(names, md.Registry+"/"+n)
	}
	err = nc.dbAddNames(md.Registry+"/"+md.CanonicalName, names)

	Log.Debug.Printf("Image name: %s -> (updated) Source ID: %v", in, newSID)
	return newSID, err
}

// GetImageName returns the docker image name for a given source ID
func (nc *NameCache) getImageName(sid sous.SourceID) (string, strpairs, error) {
	Log.Vomit.Printf("Getting image name for %+v", sid)
	name, qualities, err := nc.getImageNameFromCache(sid)
	defer func() { Log.Debug.Printf("SourceID: %q -> image name %s", sid, name) }()
	if err == nil {
		// We got it from the cache first time.
		return name, qualities, nil
	}
	if _, ok := errors.Cause(err).(NoImageNameFound); !ok {
		// We got a probable database error, give up.
		Log.Info.Printf("Cache error: %s", err)
		return "", nil, errors.Wrap(err, "getting name from cache")
	}
	// The error was a NoImageNameFound.
	if name, qualities, err = nc.getImageNameAfterHarvest(sid); err != nil {
		// Failed even after a harvest, give up.
		return "", nil, errors.Wrapf(err, "getting image from cache after harvest")
	}
	return name, qualities, nil
}

func (nc *NameCache) getImageNameFromCache(sid sous.SourceID) (string, strpairs, error) {
	cn, _, qls, err := nc.dbQueryOnSourceID(sid)
	return cn, qls, err
}

func (nc *NameCache) getImageNameAfterHarvest(sid sous.SourceID) (string, strpairs, error) {
	if err := nc.harvest(sid.Location); err != nil {
		Log.Debug.Printf("getImageName: harvest error: %v", err)
		return "", nil, err
	}
	return nc.getImageNameFromCache(sid)
}

func qualitiesFromLabels(lm map[string]string) []sous.Quality {
	advs, ok := lm[`com.opentable.sous.advisories`]
	if !ok {
		return []sous.Quality{}
	}
	var qs []sous.Quality
	for _, adv := range strings.Split(advs, `,`) {
		qs = append(qs, sous.Quality{Name: adv, Kind: "advisory"})
	}
	return qs
}

// GetCanonicalName returns the canonical name for an image given any known name
func (nc *NameCache) GetCanonicalName(in string) (string, error) {
	_, _, _, _, cn, err := nc.dbQueryOnName(in)
	Log.Debug.Printf("Canonicalizing %s - got %s / %v", in, cn, err)
	return cn, err
}

// Insert puts a given SourceID/image name pair into the name cache
// used by Builder at the moment to register after a build
func (nc *NameCache) Insert(sid sous.SourceID, in, etag string, qs []sous.Quality) error {
	return nc.dbInsert(sid, in, etag, qs)
}

func (nc *NameCache) harvest(sl sous.SourceLocation) error {
	Log.Vomit.Printf("Harvesting source location %#v", sl)
	repos, err := nc.dbQueryOnSL(sl)
	if err != nil {
		Log.Vomit.Printf("Err looking up repos for location %#v: %v - proceeding with guessed repo", sl, err)
		repos = []string{}
	}
	guessed := fullRepoName(nc.DockerRegistryHost, sl)
	knowGuess := false

	Log.Vomit.Printf("Attempting to harvest %d repos", len(repos))
	for _, r := range repos {
		if r == guessed {
			knowGuess = true
		}
		err := nc.Warmup(r)
		if err != nil {
			return err
		}
	}
	if !knowGuess {
		err := nc.Warmup(guessed)
		if err != nil {
			return err
		}
	}
	return nil
}

func union(left, right []string) []string {
	set := make(map[string]struct{})
	for _, s := range left {
		set[s] = struct{}{}
	}

	for _, s := range right {
		set[s] = struct{}{}
	}

	res := make([]string, 0, len(set))

	for k := range set {
		res = append(res, k)
	}

	return res
}

// DBConfig is a database configuration for a NameCache.
type DBConfig struct {
	Driver, Connection string
}

var schema = []string{
	"pragma foreign_keys = ON;",

	"create table _database_metadata_(" +
		"name text not null unique on conflict replace" +
		", value text" +
		");",

	"create table docker_repo_name(" +
		"repo_name_id integer primary key autoincrement" +
		", name text not null" +
		", constraint upsertable unique (name)" +
		");",

	"create table docker_search_location(" +
		"location_id integer primary key autoincrement" +
		", repo text not null" +
		", offset text not null" +
		", constraint upsertable unique (repo, offset)" +
		");",

	"create table repo_through_location(" +
		"repo_name_id references docker_repo_name" +
		"    not null" +
		", location_id references docker_search_location" +
		"    not null" +
		",  primary key (repo_name_id, location_id)" +
		");",

	"create table docker_search_metadata(" +
		"metadata_id integer primary key autoincrement" +
		", location_id references docker_search_location" +
		"    not null" +
		", etag text not null" +
		", canonicalName text not null" +
		", version text not null" +
		", constraint upsertable unique (location_id, version)" +
		", constraint canonical unique (canonicalName)" +
		");",

	"create table docker_search_name(" +
		"name_id integer primary key autoincrement" +
		", metadata_id references docker_search_metadata" +
		"    on delete cascade not null" +
		", name text not null unique" +
		");",

	// "qualities" includes advisories. assuming that assertions will also
	// be represented here
	"create table docker_image_qualities(" +
		"assertion_id integer primary key autoincrement" +
		", metadata_id references docker_search_metadata" +
		"    not null" +
		", quality text not null" +
		", kind text not null" +
		", constraint upsertable unique (metadata_id, quality, kind) on conflict ignore" +
		");",
}

var schemaFingerprint = fingerPrintSchema(schema)

var registerSQLOnce = &sync.Once{}

// GetDatabase initialises a new database for a NameCache.
func GetDatabase(cfg *DBConfig) (*sql.DB, error) {
	driver := "sqlite3_sous"
	conn := InMemory
	if cfg != nil {
		if cfg.Driver != "" {
			driver = cfg.Driver
		}
		if cfg.Connection != "" {
			conn = cfg.Connection
		}
	}

	if driver == "sqlite3" {
		driver = "sqlite3_sous"
	}

	registerSQLOnce.Do(func() {
		sql.Register(driver, &sqlite.SQLiteDriver{
			ConnectHook: func(conn *sqlite.SQLiteConn) error {
				if err := conn.RegisterFunc("semverEqual", semverEqual, true); err != nil {
					return err
				}
				return nil
			},
		})
	})

	db, err := sql.Open(driver, conn) //only call once
	return db, errors.Wrapf(err, "get DB/open: %v", cfg)
}

func semverEqual(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}
	aVer, err := semv.Parse(a)
	if err != nil {
		return false, err
	}
	bVer, err := semv.Parse(b)
	if err != nil {
		return false, err
	}
	return aVer.Equals(bVer), nil
}

// GroomDatabase ensures that the database to back the cache is the correct schema
func (nc *NameCache) GroomDatabase() error {
	db := nc.DB
	var tgp string
	err := db.QueryRow("select value from _database_metadata_ where name = 'fingerprint';").Scan(&tgp)
	if err != nil || tgp != schemaFingerprint {
		//log.Println(err, tgp, schemaFingerprint)
		err = nil
		repos := captureRepos(db)

		clobber(db)

		for _, cmd := range schema {
			if err := sqlExec(db, cmd); err != nil {
				return errors.Wrapf(err, "groom DB/create: %v", db)
			}
		}
		if _, err := db.Exec("insert into _database_metadata_ (name, value) values"+
			" ('fingerprint', ?),"+
			" ('created', ?);",
			schemaFingerprint, time.Now().UTC().Format(time.UnixDate)); err != nil {
			return errors.Wrapf(err, "groom DB/fp: %v", db)
		}
		for _, r := range repos {
			if err := nc.Warmup(r); err != nil {
				return errors.Wrap(err, "groom DB")
			}
		}
	}

	return errors.Wrap(err, "groom DB")
}

func clobber(db *sql.DB) {
	Log.Debug.Print("DB Clobbering time!")
	sqlExec(db, "PRAGMA writable_schema = 1;")
	sqlExec(db, "delete from sqlite_master where type in ('table', 'index', 'trigger');")
	sqlExec(db, "PRAGMA writable_schema = 0;")
	sqlExec(db, "vacuum;")
}

func captureRepos(db *sql.DB) (repos []string) {
	res, err := db.Query("select name from docker_repo_name;")
	if err != nil {
		Log.Debug.Print(err)
		return
	}
	for res.Next() {
		var repo string
		res.Scan(&repo)
		repos = append(repos, repo)
	}
	return
}

func fingerPrintSchema(schema []string) string {
	h := sha256.New()
	for i, s := range schema {
		fmt.Fprintf(h, "%d:%s\n", i, s)
	}
	buf := &bytes.Buffer{}
	b6 := base64.NewEncoder(base64.StdEncoding, buf)
	b6.Write(h.Sum([]byte(``)))
	b6.Close()
	return buf.String()
}

func (nc *NameCache) dumpRows(io io.Writer, sql string) {
	fmt.Fprintln(io, sql)
	rows, err := nc.DB.Query(sql)
	if err != nil {
		panic(err)
	}

	w := &tabwriter.Writer{}
	w.Init(io, 2, 4, 2, ' ', 0)
	heads, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, strings.Join(heads, "\t"))

	vals := make([]interface{}, len(heads))
	for i := range vals {
		vals[i] = new(string)
	}

	for rows.Next() {
		rows.Scan(vals...)
		for i, v := range vals {
			if i != len(vals)-1 {
				fmt.Fprintf(w, "%s\t", *(v.(*string)))
			} else {
				fmt.Fprintf(w, "%s\n", *(v.(*string)))
			}
		}
	}
	w.Flush()
	fmt.Fprintln(io, "")
}

func (nc *NameCache) dump(io io.Writer) {
	nc.dumpRows(io, "select * from _database_metadata_")
	nc.dumpRows(io, "select * from docker_repo_name")
	nc.dumpRows(io, "select * from docker_search_location")
	nc.dumpRows(io, "select * from repo_through_location")
	nc.dumpRows(io, "select * from docker_search_metadata")
	nc.dumpRows(io, "select * from docker_search_name")
	nc.dumpRows(io, "select * from docker_image_qualities")
}

func sqlExec(db *sql.DB, sql string) error {
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("Error: %s in SQL: %s", err, sql)
	}
	return nil
}

var sqlBindingRE = regexp.MustCompile(`[$]\d+`)

func (nc *NameCache) ensureInDB(sel, ins string, args ...interface{}) (id int64, err error) {
	selN := len(sqlBindingRE.FindAllString(sel, -1))
	insN := len(sqlBindingRE.FindAllString(ins, -1))
	if selN > len(args) {
		return 0, errors.Errorf("only %d args when %d needed for %q", len(args), selN, sel)
	}
	if insN > len(args) {
		return 0, errors.Errorf("only %d args when %d needed for %q", len(args), insN, ins)
	}

	row := nc.DB.QueryRow(sel, args[0:selN]...)
	err = row.Scan(&id)
	if err == nil {
		Log.Vomit.Printf("Found id: %d with %q %v", id, sel, args)
		return
	}

	if errors.Cause(err) != sql.ErrNoRows {
		return 0, errors.Wrapf(err, "getting id with %q %v", sel, args[0:selN])
	}

	nr, err := nc.DB.Exec(ins, args[0:insN]...)
	if err != nil {
		return 0, errors.Wrapf(err, "inserting new value: %q %v", ins, args[0:insN])
	}
	id, err = nr.LastInsertId()
	Log.Vomit.Printf("Made (?err: %v) id: %d with %q", err, id, ins)
	return id, errors.Wrapf(err, "getting id of new value: %q %v", ins, args[0:insN])
}

func (nc *NameCache) dbInsert(sid sous.SourceID, in, etag string, quals []sous.Quality) error {
	ref, err := reference.ParseNamed(in)
	Log.Debug.Printf("Parsed image name: %v from %q", ref, in)
	if err != nil {
		return errors.Errorf("%v for %v", err, in)
	}

	Log.Vomit.Printf("Inserting name %s for %#v", ref.Name(), sid)

	var nid, id int64
	nid, err = nc.ensureInDB(
		"select repo_name_id from docker_repo_name where name = $1",
		"insert into docker_repo_name (name) values ($1);",
		ref.Name())

	if err != nil {
		return err
	}

	Log.Vomit.Printf("name: %s -> id: %d", ref.Name(), nid)

	id, err = nc.ensureInDB(
		"select location_id from docker_search_location where repo = $1 and offset = $2",
		"insert into docker_search_location (repo, offset) values ($1, $2);",
		sid.Location.Repo, sid.Location.Dir)

	if err != nil {
		return err
	}

	Log.Vomit.Printf("Source Loc: %s -> id: %d", sid.Location, id)

	_, err = nc.DB.Exec("insert or ignore into repo_through_location "+
		"(repo_name_id, location_id) values ($1, $2)", nid, id)
	if err != nil {
		return errors.Wrapf(err, "inserting (%d, %d) into repo_through_location", nid, id)
	}

	versionString := sid.Version.Format(semv.Complete)
	Log.Vomit.Printf("Inserting metadata id:%v etag:%v name:%v version:%v", id, etag, in, versionString)

	id, err = nc.ensureInDB(
		"select metadata_id from docker_search_metadata  where canonicalName = $1",
		"insert or replace into docker_search_metadata (canonicalName, location_id, etag, version) values ($1, $2, $3, $4);",
		in, id, etag, versionString)

	if err != nil {
		return err
	}

	for _, q := range quals {
		if q.Kind == "advisory" && q.Name == "" {
			continue
		}
		nc.DB.Exec("insert into docker_image_qualities"+
			"  (metadata_id, quality, kind)"+
			"  values"+
			"  ($1,$2,$3)",
			id, q.Name, q.Kind)
	}

	Log.Vomit.Printf("Inserting search name %v %v", id, in)

	return nc.dbAddNamesForID(id, []string{in})
}

func (nc *NameCache) dbAddNamesForID(id int64, ins []string) error {
	add, err := nc.DB.Prepare("insert or replace into docker_search_name " +
		"(metadata_id, name) values ($1, $2)")
	if err != nil {
		return errors.Wrap(err, "adding names")
	}

	for _, n := range ins {
		_, err := add.Exec(id, n)
		if err != nil {
			Log.Vomit.Printf("%v %T", errors.Cause(err), errors.Cause(err))
			return errors.Wrapf(err, "adding name: %s", n)
		}
	}
	return nil
}

func (nc *NameCache) dbAddNames(cn string, ins []string) error {
	var id int64
	Log.Debug.Printf("Adding names for %s: %+v", cn, ins)
	row := nc.DB.QueryRow("select metadata_id from docker_search_metadata "+
		"where canonicalName = $1", cn)
	err := row.Scan(&id)
	if err != nil {
		return err
	}
	return nc.dbAddNamesForID(id, ins)
}

func (nc *NameCache) dbQueryOnName(in string) (etag, repo, offset, version, cname string, err error) {
	row := nc.DB.QueryRow("select "+
		"docker_search_metadata.etag, "+
		"docker_search_location.repo, "+
		"docker_search_location.offset, "+
		"docker_search_metadata.version, "+
		"docker_search_metadata.canonicalName "+
		"from "+
		"docker_search_name natural join docker_search_metadata "+
		"natural join docker_search_location "+
		"where docker_search_name.name = $1", in)
	err = row.Scan(&etag, &repo, &offset, &version, &cname)
	if err == sql.ErrNoRows {
		err = NoSourceIDFound{imageName(in)}
	}
	return
}

func (nc *NameCache) dbQueryOnSL(sl sous.SourceLocation) (rs []string, err error) {
	rows, err := nc.DB.Query("select docker_repo_name.name "+
		"from "+
		"docker_repo_name natural join repo_through_location "+
		"  natural join docker_search_location "+
		"where "+
		"docker_search_location.repo = $1 and "+
		"docker_search_location.offset = $2",
		string(sl.Repo), string(sl.Dir))

	if err == sql.ErrNoRows {
		return []string{}, err
	}
	if err != nil {
		return []string{}, err
	}

	for rows.Next() {
		var r string
		rows.Scan(&r)
		rs = append(rs, r)
	}
	err = rows.Err()
	if len(rs) == 0 {
		err = fmt.Errorf("no repos found for %+v", sl)
	}
	return
}

func (nc *NameCache) dbQueryAllSourceIds() (ids []sous.SourceID, err error) {
	rows, err := nc.DB.Query("select docker_search_location.repo, " +
		"docker_search_location.offset, " +
		"docker_search_metadata.version " +
		"from " +
		"docker_search_location natural join docker_search_metadata")
	if err != nil {
		return
	}
	for rows.Next() {
		var r, o, v string
		rows.Scan(&r, &o, &v)
		ids = append(ids, sous.SourceID{
			Location: sous.SourceLocation{
				Repo: r, Dir: o,
			},
			Version: semv.MustParse(v),
		})
	}
	err = rows.Err()
	return
}

type strpairs []strpair
type strpair [2]string

func (nc *NameCache) dbQueryOnSourceID(sid sous.SourceID) (cn string, ins []string, quals strpairs, err error) {
	rows, err := nc.DB.Query("select docker_search_metadata.canonicalName, "+
		"docker_search_name.name "+
		"from "+
		"docker_search_name natural join docker_search_metadata "+
		"natural join docker_search_location "+
		"where "+
		"docker_search_location.repo = $1 and "+
		"docker_search_location.offset = $2 and "+
		"semverEqual(docker_search_metadata.version, $3)",
		sid.Location.Repo, sid.Location.Dir, sid.Version.String())

	Log.Vomit.Printf("Selecting on %q %q %q", sid.Location.Repo, sid.Location.Dir, sid.Version.String())

	if err == sql.ErrNoRows {
		err = errors.Wrap(NoImageNameFound{sid}, "")
		return
	}
	if err != nil {
		return
	}

	for rows.Next() {
		var in string
		rows.Scan(&cn, &in)
		ins = append(ins, in)
	}
	err = rows.Err()
	if len(ins) == 0 {
		err = errors.Wrap(NoImageNameFound{sid}, "")
	}
	if err != nil {
		return
	}

	rows, err = nc.DB.Query("select"+
		" docker_image_qualities.quality,"+
		" docker_image_qualities.kind"+
		"   from"+
		" docker_image_qualities natural join docker_search_metadata"+
		" where"+
		" docker_search_metadata.canonicalName = $1", cn)

	if err != nil {
		return
	}

	for rows.Next() {
		var pr strpair
		rows.Scan(&pr[0], &pr[1])
		quals = append(quals, pr)
	}
	err = rows.Err()

	return
}
