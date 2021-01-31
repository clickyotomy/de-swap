package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

// smaps parses "/proc/{pid}/smaps" and loads
// the given slice with memory regions in swap.
func smaps(path string, regions *[]mmapSwapRegion) error {
	var (
		file *os.File
		scnr *bufio.Scanner
		err  error

		rOff uint64
		rEnd uint64
		size uint64
		swap bool
	)

	if file, err = os.Open(path); err != nil {
		return err
	}

	defer file.Close()

	// Read the whole file (line-by-line).
	scnr = bufio.NewScanner(file)
	for scnr.Scan() {
		swap = false

		// Match lines.
		if mr := capture(mmapRe, scnr.Text()); mr != nil {
			// Get the regions of memory. We can trust
			// the kernel to give us correct hex values.
			rOff, _ = strconv.ParseUint(mr["off"], 16, 64)
			rEnd, _ = strconv.ParseUint(mr["end"], 16, 64)
		} else if sw := capture(swapRe, scnr.Text()); sw != nil {
			size, _ = strconv.ParseUint(sw["sz"], 10, 64)
			// We only care about regions that are swapped.
			// Also, convert to bytes.
			if (size * 1024) > 0 {
				swap = true
			}
		}

		// Append to the slice.
		if swap && rOff > 0 && rEnd > 0 {
			*regions = append(*regions, mmapSwapRegion{
				(size * 1024), rOff, rEnd, false,
			})
		}
	}

	if err := scnr.Err(); err != nil {
		return err
	}

	return nil
}

// stream iterates oveer the slice generated from reading
// "/proc/{pid}/smaps" and sends it a channel (subscribed
// by "yank()").
func stream(ch chan mmapSwapRegion, r *[]mmapSwapRegion, wg *sync.WaitGroup) {
	for i := 0; i < len(*r); i++ {
		if !(*r)[i].split {
			ch <- (*r)[i]
		}
	}

	defer func() {
		close(ch)
		wg.Done()
	}()
}

// yank reads "/proc/{pid}/mem" for a process starting at an offset
// until the specified boundary, effectively "swapping-in" the memory
// that was "swapped-out."
func yank(ch chan mmapSwapRegion, path string, wg *sync.WaitGroup) {
	var (
		ok  bool
		err error

		file *os.File
		mr   mmapSwapRegion

		rBad uint
		rTot uint
		nBuf int

		pfx string
		log string

		buf = make([]byte, pageSz)
	)

loop:
	for {
		select {
		case mr, ok = <-ch:
			// The channel's closed - there
			// is nothing else left to do.
			if !ok {
				break loop
			}

			pfx = fmt.Sprintf(
				yankLogPfx, path, mr.off,
				mr.end, (mr.sz / 1024),
			)

			// Don't do anything, just log and continue.
			if dryRun {
				if debug > 0 {
					log = fmt.Sprintf(
						"%s %s", pfx, yankNoopMsg,
					)
					fmt.Fprintf(os.Stderr, log)
				}

				continue loop
			}

			// Read the file.
			if file, err = os.Open(path); err != nil {
				// It is (usually) a bad sign if
				// this fails; but we'll try again.
				if debug > 0 {
					log = fmt.Sprintf(
						yankErrMsg, "open", err,
					)
					fmt.Fprintf(
						os.Stderr,
						fmt.Sprintf("%s %s", pfx, log),
					)
				}

				continue loop
			}

			// Move to the offset in memory.
			_, err = file.Seek(int64(mr.off), os.SEEK_SET)
			if err != nil {
				// Maybe something else might going on
				// here (weird edge case); close the file
				// handle and move on.
				if debug > 0 {
					log = fmt.Sprintf(
						yankErrMsg, "seek", err,
					)
					fmt.Fprintf(
						os.Stderr,
						fmt.Sprintf("%s %s", pfx, log),
					)
				}

				file.Close()
				continue loop
			}

			// Bring it back to memory from swap.
			// The "yanking" happens here.
			rBad, rTot = 0, 0
			for mr.off < mr.end {
				// We don't care about the contents;
				// just note that something went wrong.
				if nBuf, err = file.Read(buf); err != nil {
					if err != io.EOF {
						rBad++
					}

					continue
				}

				if debug > 1 {
					log = fmt.Sprintf(
						yankDbgMsg, pfx, mr.off,
						(mr.off + pageSz), nBuf,
						pageSz, err,
					)
					fmt.Fprintf(os.Stderr, log)
				}

				mr.off += pageSz
				rTot++
			}

			if rBad > 0 {
				log = fmt.Sprintf(yankWarnMsg, rBad, rTot)
			} else {
				atomic.AddUint64(&mRead, mr.sz)
				log = fmt.Sprintf(yankOKMsg)
			}

			if debug >= 1 {
				fmt.Fprintf(
					os.Stderr,
					fmt.Sprintf("%s %s", pfx, log),
				)
			}

			file.Close()
		}
	}

	// We're out!
	defer func() {
		// Just in case.
		if file != nil {
			file.Close()
		}
		wg.Done()
	}()
}
