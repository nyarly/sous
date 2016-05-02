// test_with_docker provides utilities for using docker-compose for writing
// integration tests.
package test_with_docker

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type (
	serviceMap map[string]uint

	// An Agent manages operations directed at Docker
	// This is an interface that abstracts the differece between local
	// docker-daemons (useful, for instance, for Linux based CI (e.g. Travis)) and
	// VM hosted docker-machine managed daemons (e.g. for OS X development.
	Agent interface {
		//	ComposeServices uses docker-compose to set up one or more services, using
		//	serviceMap to check availability.
		//
		//	Importantly, the serviceMap is used both to determine if the services are
		//	already available - since docker-compose can take some time to execute, it
		//	can be handy to run the compose in a different console and let
		//	ComposeServices discover the services.
		//
		//	Finally, if ComposeServices determined that a service was missing and
		//	needed to be run, it will return a value that represents the
		//	docker-compose command that it executed. You can pass this value to
		//	Shutdown to shut down the docker-compose after tests have run.
		ComposeServices(string, serviceMap) (*command, error)

		// InstallFile puts a path found on the local machine to a path on the docker host.
		InstallFile(string, string) error

		// DifferingFile takes a list of pairs of [local, remote] paths, and filters them
		// for pairs whose contents differ.
		DifferingFiles(...[]string) ([][]string, error)

		// IP returns the IP address where the daemon is located.
		// In order to access the services provided by a docker-compose on a
		// docker-machine, we need to know the ip address. Some client test code
		// needs to know the IP address prior to starting up the services, which is
		// why this function is exposed
		IP() (net.IP, error)

		// MD5s computes digests of a list of paths
		// This can be used to compare to local digests and avoid copying files or
		// restarting the daemon
		MD5s(...string) (map[string]string, error)

		// RebuildService forces the rebuild of a docker-compose service.
		RebuildService(string, string) error

		// Shutdown terminates the set of services started by ComposeServices
		// If passed a nil (as ComposeServices returns in the event that all services
		// were available, Shutdown is a no-op
		Shutdown(*command)

		// RestartDaemon reboots the docker daemon
		RestartDaemon() error

		// Exec executes commands as root on the daemon host
		// It uses sudo
		Exec(...string) error
	}
)

var (
	rnums = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

	md5RE        = regexp.MustCompile(`(?m)^([0-9a-fA-F]+)\s+(\S+)$`)
	md5missingRE = regexp.MustCompile(`(?m)^md5sum: (?:can't open '(.*)'|(.*)): No such file or directory$`)
	ip           string
)

func NewAgent(timeout float32, maybeMachineName string) Agent {
	if shouldMachine() {
		log.Println("Using docker-machine")
		return &Machine{name: maybeMachineName, serviceTimeout: timeout}
	} else {
		log.Println("Using local docker daemon")
		return &LocalDaemon{serviceTimeout: timeout}
	}
}

func shouldMachine() bool {
	info := runCommand("docker", "info")
	ip := runCommand("docker-machine", "ip")
	if ip.err == nil {
		return true
	}
	if info.stderr == "Cannot connect to the Docker daemon. Is the docker daemon running on this host?\n" {
		return true
	}
	return false
}

func fileDiffs(pathPairs [][]string, localMD5, remoteMD5 map[string]string) [][]string {
	differentPairs := make([][]string, 0, len(pathPairs))
	for _, pair := range pathPairs {
		localPath, remotePath := pair[0], pair[1]

		localHash, localPresent := localMD5[localPath]
		remoteHash, remotePresent := remoteMD5[remotePath]

		if localPresent != remotePresent || strings.Compare(remoteHash, localHash) != 0 {
			differentPairs = append(differentPairs, []string{localPath, remotePath})
		}
	}

	return differentPairs
}

func composeService(dir string, ip net.IP, env []string, servicePorts serviceMap, timeout float32) (shutdown *command, err error) {
	if !servicesRunning(3.0, ip, servicePorts) {
		log.Printf("Services need to be started - tip: running `docker-compose up` in %s will speed tests up.", dir)

		shutdownCmd := dockerComposeUp(dir, ip, env, servicePorts, timeout)
		shutdown = &shutdownCmd
	} else {
		log.Printf("All services already up and running")
	}
	return
}

func dockerComposeUp(dir string, ip net.IP, env []string, services serviceMap, timeout float32) (upCmd command) {
	log.Println("Pulling compose config in ", dir)
	pullCmd := buildCommand("docker-compose", "pull")
	pullCmd.itself.Env = env
	pullCmd.itself.Dir = dir
	pullCmd.run()
	log.Println("\n", pullCmd.err)
	log.Println("\n", pullCmd.stdout)
	log.Println("\n", pullCmd.stderr)

	upCmd = buildCommand("docker-compose", "up")

	upCmd.itself.Env = env
	upCmd.itself.Dir = dir
	upCmd.start()

	if upCmd.err != nil {
		log.Panic(upCmd.err)
	}

	if servicesRunning(timeout, ip, services) {
		return
	}
	log.Println("\n", upCmd.stdout)
	log.Println("\n", upCmd.stderr)
	panic(fmt.Sprintf("Services were not available!"))
}

func (c *command) interrupt() {
	c.itself.Process.Signal(syscall.SIGTERM)
	c.wait()
}

func dockerComposeDown(cmd *command) error {
	log.Print("Downing compose started by: ", cmd)
	cmd.interrupt()
	if cmd.err != nil {
		return cmd.err
	}

	down := buildCommand("docker-compose", "down")
	down.itself.Env = cmd.itself.Env
	down.itself.Dir = cmd.itself.Dir
	down.run()

	return down.err
}

func rebuildService(dir, name string, env []string) error {
	cmd := buildCommand("docker-compose", "build", "--no-cache", name)
	cmd.itself.Env = env
	cmd.itself.Dir = dir
	cmd.run()
	return cmd.err
}

func servicesRunning(limit float32, ip net.IP, services map[string]uint) bool {
	goodCh := make(chan string)
	badCh := make(chan string)
	done := make(chan bool)
	defer close(done)

	for name, port := range services {
		go func(name string, ip net.IP, port uint) {
			if serviceRunning(done, ip, port) {
				goodCh <- name
			} else {
				badCh <- name
			}
		}(name, ip, port)
	}

	for len(services) > 0 {
		select {
		case good := <-goodCh:
			log.Printf("  %s up and running", good)
			delete(services, good)
		case bad := <-badCh:
			log.Printf("  Error trying to connect to %s", bad)
			return false
		case <-time.After(time.Duration(limit * float32(time.Second))):
			log.Printf("Attempt to contact remaining service expired after %f seconds", limit)
			for service, port := range services {
				log.Printf("  Still unavailable: %s at %s:%d", service, ip, port)
			}

			return false
		}
	}
	return true
}

func serviceRunning(done chan bool, ip net.IP, port uint) bool {
	addr := fmt.Sprintf("%s:%d", ip, port)
	log.Print("Attempting connection: ", addr)

	for {
		select {
		case <-done:
			return false
		default:
			conn, err := net.Dial("tcp", addr)
			defer func() {
				if conn != nil {
					conn.Close()
				}
			}()

			if err != nil {
				if _, ok := err.(*net.OpError); ok {
					time.Sleep(time.Duration(0.5 * float32(time.Second)))
					continue
				}
				return false
			}

			return true
		}
	}
}

func localMD5s(paths ...string) (md5s map[string]string) {
	md5s = make(map[string]string)

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		hash := md5.New()
		io.Copy(hash, file)
		md5s[path] = fmt.Sprintf("%x", hash.Sum(nil))
	}
	return
}