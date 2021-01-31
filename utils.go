package main

import (
	"fmt"
	"os"
	"regexp"
)

// help prints the usage of the command.
func help() {
	fmt.Fprintf(os.Stderr, usage)
}

// fmtBytes format bytes to a human-readable string (IEC).
func fmtBytes(nb uint64) string {
	var (
		base = uint64(1024)
		exp  = base
		idx  = uint64(0)

		div uint64
	)

	if nb < base {
		return fmt.Sprintf("%d B", nb)
	}

	for div = nb / base; div >= base; div /= base {
		exp *= base
		idx++
	}

	return fmt.Sprintf("%d %cB", (nb / exp), "kMGTPE"[idx])
}

// capture returns the captured groups for a string
// matched with the given regular expression as a map.
func capture(re *regexp.Regexp, str string) map[string]string {
	var match = re.FindStringSubmatch(str)
	if match != nil {
		var group = make(map[string]string)
		for idx, name := range re.SubexpNames() {
			// Because the first one is the fully
			// matched line; it can be discarded.
			if (idx > 0) && idx < len(match) {
				group[name] = match[idx]
			}
		}

		return group
	}

	return nil
}

// split divides memory regions with large swapped-out areas
// into regions of that are at most "overflow" kilobytes each.
func split(mr *[]mmapSwapRegion) {
	var (
		idx int
		off uint64
		end uint64
		sz  uint64
		brk []mmapSwapRegion
		big []int
		nr  int
	)

	// Find regions containing with swapped-out
	// memory greater than the split threshold.
	for idx = 0; idx < len(*mr); idx++ {
		if (*mr)[idx].sz > overflow {
			big = append(big, idx)

			// Mark the region as split.
			(*mr)[idx].split = true
		}
	}

	// Split up those regions into snaller chunks.
	for _, idx = range big {
		nr = 0
		for off = (*mr)[idx].off; off < (*mr)[idx].end; off += overflow {
			if (off + overflow) >= (*mr)[idx].end {
				// For leftover regions (towards the end).
				end = (*mr)[idx].end
				sz = (end - off)
			} else {
				end = (off + overflow)
				sz = overflow
			}

			brk = append(brk, mmapSwapRegion{sz, off, end, false})

			nr++
		}
	}

	fmt.Fprintf(os.Stdout, fmt.Sprintf(
		splitLogMsg, len(*mr), len(brk), len(big),
		len(*mr)-len(big), fmtBytes(overflow),
	))

	// Add the new regions.
	*mr = append(*mr, brk...)
}
