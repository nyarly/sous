package storage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/opentable/sous/lib"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

// GitStateManager wraps a DiskStateManager and implements transactional writes
// to a Git remote. It also polls the Git remote for changes
//
// Methods of GitStateManager are serialised, and thus safe for concurrent
// access. No two GitStateManagers should have DiskStateManagers using the same
// BaseDir.
type GitStateManager struct {
	sync.Mutex
	*DiskStateManager //can't just be a StateReader/Writer: needs dir
	remote            string
}

// NewGitStateManager creates a new GitStateManager wrapping the provided
// DiskStateManager.
func NewGitStateManager(dsm *DiskStateManager) *GitStateManager {
	return &GitStateManager{DiskStateManager: dsm}
}

func (gsm *GitStateManager) git(cmd ...string) error {
	if !gsm.isRepo() {
		return nil
	}
	git := exec.Command(`git`, cmd...)
	git.Dir = gsm.DiskStateManager.BaseDir
	//git.Env = []string{"GIT_CONFIG_NOSYSTEM=true", "HOME=none", "XDG_CONFIG_HOME=none"}
	out, err := git.CombinedOutput()
	if err == nil {
		sous.Log.Debug.Printf("%+v: success", git.Args)
	} else {
		sous.Log.Debug.Printf("%+v: error: %v", git.Args, err)
	}
	sous.Log.Vomit.Print("git: " + string(out))
	return errors.Wrapf(err, strings.Join(git.Args, " ")+": "+string(out))
}

func (gsm *GitStateManager) revert(tn string) {
	gsm.git("reset", "--hard", tn)
	gsm.git("clean", "-f")
}

func (gsm *GitStateManager) isRepo() bool {
	s, err := os.Stat(filepath.Join(gsm.DiskStateManager.BaseDir, ".git"))
	return err == nil && s.IsDir()
}

// ReadState reads sous state from the local disk.
func (gsm *GitStateManager) ReadState() (*sous.State, error) {
	// git pull
	gsm.git("pull")

	return gsm.DiskStateManager.ReadState()
}

func (gsm *GitStateManager) needCommit() bool {
	err := gsm.git("diff-index", "--exit-code", "HEAD")
	if ee, is := errors.Cause(err).(*exec.ExitError); is {
		return !ee.Success()
	}
	return false
}

// WriteState writes sous state to disk, then attempts to push it to Remote.
// If the push fails, the state is reset and an error is returned.
func (gsm *GitStateManager) WriteState(s *sous.State) (err error) {
	// git pull
	tn := "sous-fallback-" + uuid.New()
	if err = gsm.git("tag", tn); err != nil {
		return
	}
	defer gsm.git("tag", "-d", tn)

	if err = gsm.DiskStateManager.WriteState(s); err != nil {
		return
	}
	if err = gsm.git(`add`, `.`); err != nil {
		gsm.revert(tn)
		return
	}
	if gsm.needCommit() {
		if err = gsm.git("commit", "-m", "sous commit: Update State"); err != nil {
			gsm.revert(tn)
			return
		}
		err = gsm.git("push", "-u", "origin", "master")
		if err != nil {
			gsm.revert(tn)
		}
	}
	return
}
