// Command de-swap brings back swapped out pages into memory.
//
// Usage
//   de-swap -p <pid> [-n -j <threads> -r <bytes> -v[v]]
//
// Arguments
//   -p     PID of the process to swap-in
//   -n     no-op mode; turned off by default
//   -j     number of parallel operations; runs as a
//          single thread if unspecified
//   -r     split memory regions with large swap areas
//          into smaller regions for better throughput
//          during reads; disabled if set to 0, and if
//          enabled it must be > 0 and in exponents of
//          2; defaults to 64 kB if unspecified
//   -v[v]  output verbosity; off by default
package main

import "fmt"

var usage = fmt.Sprintf(
	"de-swap: Bring back swapped out pages into memory.\n\n"+
		"USAGE\n"+
		"  de-swap -p <pid> [-n -j <threads> -r <bytes> -v[v]]\n\n"+
		"ARGUMENTS\n"+
		"  -p     PID of the process to swap-in\n"+
		"  -n     no-op mode; turned off by default\n"+
		"  -j     number of parallel operations; runs as a\n"+
		"         single thread if unspecified\n"+
		"  -r     split memory regions with large swap areas\n"+
		"         into smaller regions for better throughput\n"+
		"         during reads; disabled if set to 0, and if\n"+
		"         enabled it must be > 0 and in exponents of\n"+
		"         2; defaults to %s if unspecified\n"+
		"  -v[v]  output verbosity; off by default\n",
	fmtBytes(overflow),
)
