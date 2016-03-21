package cli

import "sync"

var (
	cleanupTasks = make(chan func() error)
	wg           sync.WaitGroup
)

// AddCleanupTask registers a cleanup task that will be run before the
// program terminates.
func AddCleanupTask(f func() error) {
	// Each time this is called, we increment the waitgroup and spin up
	// a goroutine waiting to write a cleanupTask decorated with decrementing
	// the waitgroup to the channel. When Cleanup is called, it consumes this
	// channel, running each task one at a time, and waits on wg.
	wg.Add(1)
	go func() {
		cleanupTasks <- func() error {
			defer wg.Done()
			return f()
		}
	}()
}

// Cleanup starts consuming the cleanupTasks channel, as each
// of the waiting tasks collapse into it randomly. Note that
// cleanup tasks are able to add more cleanup tasks if necessary.
// Beware of infinitely populating the cleanup queue ;)
func Cleanup() {
	go func() {
		for t := range cleanupTasks {
			if err := t(); err != nil {
				Logf("Cleanup error: %s", err)
			}
		}
	}()
	wg.Wait()
}
