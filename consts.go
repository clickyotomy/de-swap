package main

const (
	// mmapPat is a regular expression for mapped memory ranges
	// in "/proc/{pid}/smaps" of a process. For example, it would
	// be "deadbeef-badb002 r-xp 00000000 03:14 0 /foo/bar" and we
	// should capture "deadbeef" and "badb002" in the match group.
	mmapPat = `(?is)^(?P<off>[0-9a-f]+)-(?P<end>[0-9a-f]+)\s+[rwxp-]{4}`

	// swapPat is a regular expression for the amount of swap memory
	// used by a region described above. Typically, a line with that
	// information would look lik "Swap: 42 kB", and the captureed
	// value will be "42".
	swapPat = `(?is)^Swap:\s*(?P<sz>\d+)\s+kB`

	// pidSmaps and pidMem are paths to the process's
	// memory mapping and memory on `procfs' respectively.
	pidSmaps = "/proc/%d/smaps"
	pidMem   = "/proc/%d/mem"

	// For logging.
	mainRunMsg = "de-swap: main:\tPID %d (theads: %d, split: %s, " +
		"no-op: %v)\n"
	mainEndMsg = "de-swap: main:\tOK; swapped-in %d kB in %s\n"
	mainErrMsg = "de-swap: main:\tERR: %s\n"
	argsErrMsg = "de-swap: args:\tERR: %s\n"

	smapsSwapMsg = "de-swap: smaps:\t%s: in-swap: %d kB, regions: %d\n"
	smapsReadMsg = "de-swap: smaps:\t%s: to-read: %d kB, regions: %d\n"
	splitLogMsg  = "de-swap: split:\tinit: %d, brk: %d, big: %d, " +
		"skip: %d, overflow: %s\n"

	yankLogPfx  = "de-swap: yank:\t%s: region[0x%012x-0x%012x]: %8d kB"
	yankOKMsg   = "OK\n"
	yankWarnMsg = "WARN:\tread-fail: %-8d/%8d)\n"
	yankErrMsg  = "ERR:\t%s: %v)\n"
	yankNoopMsg = "NO-OP\n"
	yankDbgMsg  = "%s, read(from=0x%012x, to=0x%012x): " +
		"got/exp: %8d/%-8d B (err: %v)\n"

	swapErrMsg = "swap: failed to swap-in some memory (read-diff: %d kB)"
)
