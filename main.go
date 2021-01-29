package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Compile the expressions here (it needs to be done only once).
	mmapRe = regexp.MustCompile(mmapPat)
	swapRe = regexp.MustCompile(swapPat)

	// For no-op and debug outputs.
	dryRun = false
	debug  = 0

	// Total number of kilobytes swapped out (approximately).
	swappedOut = uint64(0)
)

// mmappedSwap holds the range of the "mmap()"-ed region
// which contains the swapped memory and its size.
type mmapSwapRegion struct {
	sz  uint64 // Size.
	off int64  // Begin offset (for seeking).
	end int64  // End of region.
}

// deSwap brings back a process's memory from the swap-space.
func deSwap(pid, routines int) error {
	var (
		err error
		wg  sync.WaitGroup
		ent mmapSwapRegion

		swp = make([]mmapSwapRegion, 0)
		mch = make(chan mmapSwapRegion)
		mem = fmt.Sprintf(pidMem, pid)
		tot = uint64(0)
	)

	// Read all the entries at once because we don't
	// want to keep the file open for a long time.
	if err = loadSmaps(fmt.Sprintf(pidSmaps, pid), &swp); err != nil {
		return err
	}

	if len(swp) <= 0 {
		return nil
	}

	// Start sending data to the channel.
	wg.Add(1)
	go stream(mch, &swp, &wg)

	// Spawn workers to "yank" memory out of the swap-space.
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go yank(mch, mem, &wg)
	}

	// Wait for all go-routines to finish.
	wg.Wait()

	for _, ent = range swp {
		tot += ent.sz
	}

	if atomic.LoadUint64(&swappedOut) != tot && !dryRun {
		return fmt.Errorf(
			"swap: failed to swap-in some "+
				"memory regions; diff: %d kB",
			(tot - atomic.LoadUint64(&swappedOut)),
		)
	}

	return nil
}

func main() {
	flag.Usage = help

	var (
		pid = flag.Int("p", -1, "PID")
		thr = flag.Int("j", 1, "number of threads")
		dry = flag.Bool("n", false, "no-op")
		dbg = flag.Bool("v", false, "debug")
		vrb = flag.Bool("vv", false, "debug (verbose)")

		// For tracking the time elapsed.
		now = time.Now()
	)

	// Parse and validate the specified arguments.
	flag.Parse()

	if *pid <= 0 {
		fmt.Fprintf(
			os.Stderr,
			"de-swap: ERR: invalid or no PID specified\n",
		)
		os.Exit(2)
	}

	// Doesn't matter, just use 1.
	if *thr <= 0 {
		fmt.Fprintf(
			os.Stderr,
			"de-swap: ERR: bad thread value; using defaults\n",
		)

		*thr = 1
	}

	// For no-op mode.
	dryRun = *dry

	// For debug and verbose output.
	if *dbg {
		debug = 1
	}

	if *vrb {
		debug = 2
	}

	fmt.Fprintf(os.Stdout,
		"de-swap: running for PID %d (no-op: %v, threads: %d)\n",
		*pid, dryRun, *thr,
	)

	if err := deSwap(*pid, *thr); err != nil {
		fmt.Fprintf(os.Stderr, "de-swap: ERR: %s\n", err)
		os.Exit(1)
	} else {

		fmt.Fprintf(os.Stdout,
			"de-swap: OK; moved %d kB in %s\n",
			atomic.LoadUint64(&swappedOut), time.Since(now),
		)
	}
}
