// Command de-swap brings back swapped out pages into memory.
// Usage
//   de-swap -p <pid> [-n -j <threads> -r <bytes> -v[v]]
// Arguments
//   -p     PID of the process to swap-in
//   -n     no-op mode; turned off by default
//   -j     number of parallel operations; runs as a
//          single thread if unspecified
//   -r     split memory regions with large swap areas
//          into smaller regions for better throughput
//          during reads
//   -v[v]  output verbosity; off by defaultpackage main
package main

const usage = "de-swap: Bring back swapped out pages into memory.\n\n" +
	"USAGE\n" +
	"  de-swap -p <pid> [-n -j <threads> -r <bytes> -v[v]]\n\n" +
	"ARGUMENTS\n" +
	"  -p     PID of the process to swap-in\n" +
	"  -n     no-op mode; turned off by default\n" +
	"  -j     number of parallel operations; runs as a\n" +
	"         single thread if unspecified\n" +
	"  -r     split memory regions with large swap areas\n" +
	"         into smaller regions for better throughput\n" +
	"         during reads\n" +
	"  -v[v]  output verbosity; off by default\n"
