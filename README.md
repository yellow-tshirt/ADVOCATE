# AdvocateGo: Automated Detection and Verification Of Concurrency bugs through Analysis of Trace recordings of program Executions in Go."

> [!WARNING]
> This program is still under development and may return no or wrong results.

## What
We want to analyze concurrent Go programs to automatically find potential concurrency bug. The different analysis scenarios can be found in `doc/Analysis.md`.

We also implement a trace replay mechanism, to replay a trace as recorded.

## Recording
To analyze the program, we first need
to record it. To do this, we modify the go runtime
to automatically record a program while it runs. The modified runtime can 
be found in the `advocate-go-patch` directory. Running a program with this modified 
go runtime will create a trace of the program including 

- spawning of new routines
- atomic operations
- mutex operations
- channel operations
- select operations
- wait group operations
- once operations

The following is a short explanation about how to build and run the 
new runtime and create the trace. A full explanation of the created trace can be found in the 
`doc` directory. 

> [!WARNING]
> The recording of atomic operations is only tested with `amd64`. For `arm64` an untested implementation exists. 

### How
The go-patch folder contains a modified version of the go runtime.
With this modified version it is possible to save a trace of the program.

To build the new runtime, run the `make.bash` or `make.bat` file in the `src`
directory. This will create a `bin` directory containing a `go` executable.
This executable can be used as your new go environment e.g. with
`./go run main.go` or `./go build`.

WARNING: It can currently happen, that `make.bash` command result in a `fatal error: runtime: releaseSudog with non-nil gp.param`. It can normally be fixed by just running `make.bash` again. I'm working on fixing it.

It is necessary to set the GOROOT environment variable to the path of `advocate-go-patch`, e.g. with 
```
export GOROOT=$HOME/ADVOCATE/advocate-go-patch/
```

To create a trace, add

```go
runtime.InitAdvocate(0)
defer advocate.CreateTrace("trace_name.log")
```

at the beginning of the main function.
Also include the following imports 
```go
runtime
advocate
```

Autocompletion often includes "std/runtime" instead of "runtime". Make sure to include the correct one.

For some reason, `fmt.Print` and similar can lead to `fatal error: schedule: holding lock`. In this case increase the argument in `runtime.InitAtomics(0)` until the problem disappears.

After that run the program with `./go run main.go` or `./go build && ./main`,
using the new runtime.


### Example
Let's create the trace for the following program:

```go
package main

import (
	"time"
)

func main() {
	c := make(chan int, 0)

	go func() {
		c <- 1
	}()

	go func() {
		<-c
	}()

	time.Sleep(10 * time.Millisecond)
	close(c)
}
```

After adding the preamble, we get 

```go
package main

import (
	"runtime"
	"advocate"
	"time"
)

func main() {
	// ======= Preamble Start =======
	runtime.InitAdvocate(0)
	defer advocate.CreateTrace("trace_name.log")
	// ======= Preamble End =======

	c := make(chan int, 0)

	go func() {
		c <- 1  // line 48
	}()

	go func() {
		<-c  // line 52
	}()

	time.Sleep(10 * time.Millisecond)
	close(c)  // line 56
}
```

Running this leads to the following trace (indented lines are in the same line 
as the previous line, only for better readability):

```txt
G,1,2;G,2,3;G,3,4;C,4,9,1,R,f,1,2,.../go-patch/src/runtime/mgc.go:180;C,10,11,1,R,f,2,2,.../go-patch/src/runtime/mgc.go:181;G,12,5;C,13,13,2,C,f,0,0,.../go-patch/src/runtime/proc.go:256;G,14,6;G,15,7;G,16,8;C,21,21,4,C,f,0,0,.../main.go:56

C,7,8,1,S,f,2,2,.../go-patch/src/runtime/mgcsweep.go:279
C,5,6,1,S,f,1,2,.../go-patch/src/runtime/mgcscavenge.go:652


C,18,19,4,S,f,1,0,.../main.go:48
C,17,20,4,R,f,1,0,.../main.go:52
```
In this example the file paths are shortened for readability. In the real trace, the full path is given.

The trace includes both the concurrent operations of the program it self, as well
as internal operations used by the go runtime. An explanation of the trace 
file including the explanations for all elements can be found in the `doc`
directory.

## Analysis and Reorder

We can now analyze the created file using the program in the `analyzer`
folder. For now we only support the search for potential send on a closed channel, but we plan to expand the use cases in the future.
The analyzer can also create a new reordered trace, in which a detected possible bug actually occurs. This new trace can then used in the replay, to confirm and simplify the removal of the bug.

> [!WARNING]
> The Reorder is still in development and may result in incorrect traces

The analyzer can take the following command line arguments:

- -t [file]: path to the trace file
- -d [level]: output level, 0: silent, 1: results, 2: errors, 3: info, 4: debug, default: 2
- -f: if set, the analyzer assumes a fifo ordering of messages in the buffer of buffered channels. This is not part of the [Go Memory Mode](https://go.dev/ref/mem), but should follow from the implementation. For this reason, it is only an optional addition.
- -n: If an analysis is run, a rewritten trace can be direcly created in the program. If the rewrite should not be done based on the direct analysis, but on a given trace and result file, -n must be set. In this case no analysis is run. -r and -i must be set.
- -c: Normally, we assume the order of critical sections as fixed. By setting -c, we 
ignore this, and assume we can reorder critical sections of different routines.
This can help us to find bugs, that cannot be found without it, but it can also lead to 
false positives.
- -r [file]: Path to the analysis result file. Only needed if -n is set. The analysis will create two result files `result_readable.log` and `result_machine.log`. Here `result_machine.log` must be used.
- -i [index]: Index of the bug that will be reproduced. Only needed if n is set. 1 based.

If we assume the trace from our example is saved in file `trace.go` and run the analyzer with
```
./analyzer -t "trace.log"
```
it will create the following result, show it in the terminal and print it into an `result_readable.log` file: 
```txt
==================== Summary ====================

-------------------- Critical -------------------
1 Possible send on closed channel:
	close: .../main.go:56
	send: .../main.go:48
-------------------- Warning --------------------
2 Possible receive on closed channel:
	close: .../main.go:56
	recv: .../main.go:42
```
The send can cause a panic of the program, if it occurs. It is therefor an error message (in terminal red).

A receive on a closed channel does not cause a panic, but returns a default value. It can therefor be a desired behavior. For this reason it is only considered a warning (in terminal orange, can be silenced with -w).


## Trace Replay
The trace replay reruns a given program as given in the recorded trace. Please be aware, 
that only the recorded elements are considered for the trace replay. This means, that 
the order of non-recorded operations between two or more routines can still very. 

<!-- The implementation of the trace replay is not finished yet. The following is a short overview over the current state.
- order enforcement for most elements.
	- The operations are started in same global order as in the recorded trace. 
	- This is not yet implemented for the spawn of new routines and atomic operations
- correct channel partner
	- Communication partner of (most) channel operations are identical to the partners in the trace. For selects this cannot be guarantied yet. -->

### How
To start the replay, add the following header at the beginning of the 
main function:

```go
trace := advocate.ReadTrace("trace.log")
runtime.EnableReplay(trace)
defer runtime.WaitForReplayFinish()
```

`"trace.log"` must be replaced with the path to the trace file. Also include the following imports:
```go
"advocate"
"runtime"
```
Now the program can be run with the modified go routine, identical to the recording of the trace (remember to export the new gopath). 

If you want replay and at the same time record the program, make sure to add 
the tracing header before the replay header. Otherwise the program will crash
```go
// init tracing
runtime.InitAdvocate(0)
defer advocate.CreateTrace("trace_new.log")

// init replay
trace := advocate.ReadTrace("trace_old.log")
runtime.EnableReplay(trace)
defer runtime.WaitForReplayFinish()
```

### Warning:
It is the users responsibility of the user to make sure, that the input to 
the program, including e.g. API calls are equal for the recording and the 
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.
