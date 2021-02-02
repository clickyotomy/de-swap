package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// Compile the expressions here (it needs to be done only once).
	mmapRe = regexp.MustCompile(mmapPat)
	swapRe = regexp.MustCompile(swapPat)

	// For no-op, disabled split-reads, and debug logging.
	dryRun = false
	rSplit = true
	debug  = 0

	// pageSz is the amount of memory to be read (in chunks)
	// from "/proc/{pid}/mem" per read operation. This value
	// is set to the memory page-size of the machine.
	pageSz = uint64(os.Getpagesize()) // In bytes.

	// overflow is the threshold for swapped-out memory in a
	// region that triggers a split operation in that region (break
	// it into smaller chunks) for better performance. By default,
	// it is set to 64 kB.
	overflow = uint64(64 * 1024)

	// Total number of kilobytes read and swapped-in,
	mRead = uint64(0)
	mSwap = uint64(0)
)

// mmappedSwap holds the range of the "mmap()"-ed region
// which contains the swapped memory and its size.
type mmapSwapRegion struct {
	sz    uint64 // Number of kB in swap-space.
	off   uint64 // Begin offset (for seeking).
	end   uint64 // End of region.
	split bool   // Has been split into smaller regions.
}

// deSwap brings back a process's memory from the swap-space.
func deswap(pid, routines uint) error {
	var (
		err error
		wg  sync.WaitGroup
		ent mmapSwapRegion
		nr  int
		thr uint

		swp = make([]mmapSwapRegion, 0)
		mch = make(chan mmapSwapRegion)
		mem = fmt.Sprintf(pidMem, pid)
		smp = fmt.Sprintf(pidSmaps, pid)
		rd  = uint64(0)
		sw  = uint64(0)
	)

	// Read all the entries at once because we don't
	// want to keep the file open for a long time.
	if err = smaps(smp, &swp); err != nil {
		return err
	}

	if len(swp) <= 0 {
		return nil
	}

	// Get the stats for swapped-in memory.
	for _, ent = range swp {
		if !ent.split {
			sw += ent.sz
			nr++
		}
	}
	fmt.Fprintf(os.Stdout, fmt.Sprintf(smapsSwapMsg, smp, (sw/1024), nr))

	// Split the regions into smaller chunks; loop again
	// for total readable memory (after the splitting).
	if rSplit {
		split(&swp)
		nr = 0
		for _, ent = range swp {
			if !ent.split {
				rd += ent.sz
				nr++
			}
		}
		fmt.Fprintf(os.Stdout, fmt.Sprintf(
			smapsReadMsg, smp, (rd/1024), nr,
		))
	}

	// Start sending data to the channel.
	wg.Add(1)
	go stream(mch, &swp, &wg)

	// Spawn workers to "yank" memory out of the swap-space.
	for thr = 0; thr < routines; thr++ {
		wg.Add(1)
		go yank(mch, mem, &wg)
	}

	// Wait for all go-routines to finish.
	wg.Wait()

	if atomic.LoadUint64(&mRead) != rd && !dryRun {
		return fmt.Errorf(fmt.Sprintf(
			swapErrMsg,
			((rd - atomic.LoadUint64(&mRead)) / 1024),
		))
	}

	if !dryRun {
		mSwap = sw
	}
	return nil
}

func main() {
	flag.Usage = help

	var (
		pid = flag.Uint("p", 0, "PID")
		dry = flag.Bool("n", false, "no-op")
		thr = flag.Uint("j", 1, "number of threads")
		ssz = flag.Uint64("r", overflow, "region split size (bytes)")
		dbg = flag.Bool("v", false, "debug")
		vrb = flag.Bool("vv", false, "extra debug")

		// For tracking the time elapsed.
		now = time.Now()
	)

	// Parse and validate the specified arguments.
	flag.Parse()

	if *pid <= 0 {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(
			argsErrMsg, "invalid or no PID specified",
		))
		os.Exit(2)
	}

	// Doesn't matter, just use 1.
	if *thr <= 0 {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(
			argsErrMsg, "invalid thread value",
		))
		os.Exit(2)
	}

	if *ssz == 0 {
		rSplit = false
	} else if (*ssz < 0) || ((*ssz & (*ssz - 1)) != 0) {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(
			argsErrMsg, "bad read split size value",
		))
		os.Exit(2)
	}

	// For no-op mode.
	dryRun = *dry

	// Set read split size value.
	overflow = *ssz

	// For debug and verbose output.
	if *dbg {
		debug = 1
	}

	if *vrb {
		debug = 2
	}

	fmt.Fprintf(os.Stdout, fmt.Sprintf(
		mainRunMsg, *pid, *thr, fmtBytes(overflow), dryRun,
	))

	if err := deswap(*pid, *thr); err != nil {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(mainErrMsg, err))
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, fmt.Sprintf(
			mainEndMsg, (mSwap/1024), time.Since(now),
		))
	}
}
