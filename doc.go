// Command de-swap brings back swapped out pages into memory.
//
// Usage
//    de-swap -p <pid> [-n -j <threads> -v[v]]
//
// Arguments
//   -p       PID of the process to swap-in
//   -n       no-op mode; off by default
//   -j       number of parallel operations; defaults
// 	   to a single thread if unspecified
//   -v[v]    output verbosity; off by default
package main

const usage = " de-swap: Bring back swapped out pages into memory.\n\n" +
	" USAGE\n" +
	"    de-swap -p <pid> [-n -j <threads> -v[v]]\n\n" +
	" ARGUMENTS\n" +
	"   -p       PID of the process to swap-in\n" +
	"   -n       no-op mode; off by default\n" +
	"   -j       number of parallel operations; defaults\n" +
	"            to a single thread if unspecified\n" +
	"   -v[v]    output verbosity; off by default\n"
