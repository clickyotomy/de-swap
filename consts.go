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

	// pageSz is the amount of memory to be read from "/proc/{pid}/mem"
	// per read operation (assuming 4 kB pages).
	pageSz = int64(4096)

	// For logging.
	yankLogPfx  = "de-swap: * %s: region[0x%012x-0x%012x]: %8d kB"
	yankErrMsg  = "ERR  (%s: %v)\n"
	yankOKMsg   = "OK\n"
	yankWarnMsg = "WARN (read-fail: %d/%d)\n"
	yankNoopMsg = "NO-OP\n"
	yankDbgMsg  = "%s, read(from=0x%012x, to=0x%012x): " +
		"got/exp: %8d/%-8d B (err: %v)\n"
)
