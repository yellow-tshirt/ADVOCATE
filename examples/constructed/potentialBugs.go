package main

/*
import "sync"
import "time"
import "fmt"

*/

import (
	"advocate"
	"flag"
	"runtime"
	"sync"
	"time"
)

// FN = False negative
// FP = False positive
// TN = True negative
// TP = True positive

// =========== Send / Received to/from closed channel ===========

//////////////////////////////////////////////////////////////
// No send of closed due to (must) happens before relations.

// Synchronous channel.
// TN.
func n01() {
	x := make(chan int)
	ch := make(chan int, 1)

	go func() {
		ch <- 1
		x <- 1
	}()

	<-x
	close(ch)
}

// Wait group
// TN
func n02() {
	ch := make(chan int, 1)
	var g sync.WaitGroup

	g.Add(1)

	func() {
		ch <- 1
		g.Done()
	}()

	g.Wait()
	close(ch)

}

// Once
// TN
func n03() {
	var once sync.Once
	ch := make(chan int, 1)
	setup := func() {
		ch <- 1
	}

	once.Do(setup)
	close(ch)

}

// RWMutex
// FN.
/*

T1 -> T2 -> T3 due to sleep statements

RU2 and RU1 sync with L

 => send <HB close


If we reorder critical sections,
we encounter send on closed.

*/
func n04() {
	var m sync.RWMutex
	ch := make(chan int, 1)

	// T1
	go func() {
		m.RLock()
		ch <- 1
		m.RUnlock() // RU1

	}()

	// T2
	go func() {
		time.Sleep(300 * time.Millisecond)
		m.RLock()
		m.RUnlock() // RU2

	}()

	// T3
	time.Sleep(1 * time.Second)
	m.Lock() // L
	close(ch)
	m.Unlock()

}

// TP send on closed
// TP recv on closed
func n05() {
	c := make(chan int)

	go func() {
		c <- 1
	}()

	go func() {
		<-c
	}()

	time.Sleep(300 * time.Millisecond) // prevent actual send on closed channel
	close(c)
}

// TP send on closed
// TP recv on closed
func n06() {
	c := make(chan int, 1)

	go func() {
		c <- 1
		<-c
	}()

	time.Sleep(300 * time.Millisecond) // prevent actual send on closed channel
	close(c)
}

// TN recv/send on closed
func n07() {
	c := make(chan int)

	go func() {
		c <- 1
	}()

	<-c

	close(c)
}

// TP recv on closed
func n08() {
	c := make(chan int)

	go func() {
		time.Sleep(300 * time.Millisecond) // force actual recv on closed channel
		<-c
	}()

	close(c)
	time.Sleep(1 * time.Second) // prevent termination before receive
}

// TP send on closed
// TP recv on closed
func n09() {
	c := make(chan struct{}, 1)
	d := make(chan struct{}, 1)

	go func() {
		time.Sleep(300 * time.Millisecond) // prevent actual send on closed channel
		close(c)
		close(d)
	}()

	go func() {
		select {
		case c <- struct{}{}:
		default:
		}

		select {
		case <-d:
		default:
		}
	}()

	d <- struct{}{}
	<-c

	time.Sleep(1 * time.Second) // prevent termination before receive
}

// FN
func n10() {
	c := make(chan struct{}, 0)

	go func() {
		time.Sleep(200 * time.Millisecond) // prevent actual send on closed channel
		close(c)
	}()

	go func() {
		select {
		case c <- struct{}{}:
		default:
		}
	}()

	time.Sleep(500 * time.Millisecond) // make sure, that the default values are taken
}

// TP
func n11() {
	c := make(chan struct{}, 0)

	go func() {
		time.Sleep(200 * time.Millisecond) // prevent actual send on closed channel
		close(c)
	}()

	go func() {
		select {
		case c <- struct{}{}:
		default:
		}
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-c:
		default:
		}
	}()

	time.Sleep(300 * time.Millisecond) // make sure, that the default values are taken
}

// TN: no send to closed channel because of once
func n12() {
	c := make(chan int, 1)

	once := sync.Once{}

	go func() {
		once.Do(func() {
			c <- 1
		})
	}()

	go func() {
		once.Do(func() {
			close(c)
		})
	}()

	time.Sleep(100 * time.Millisecond)
}

// FN: potential send to closed channel not recorded because of once
func n13() {
	c := make(chan int, 1)

	once := sync.Once{}

	close(c)

	go func() {
		time.Sleep(100 * time.Millisecond) // prevent actual send on closed channel
		once.Do(func() {
			c <- 1
		})
	}()

	go func() {
		once.Do(func() {
			// do nothing
		})
	}()

	time.Sleep(100 * time.Millisecond)
}

// TP: potential send to closed channel recorded with once
func n14() {
	c := make(chan int, 1)

	once := sync.Once{}

	go func() {
		once.Do(func() {
			c <- 1
		})
	}()

	go func() {
		time.Sleep(100 * time.Millisecond) // prevent actual send on closed channel
		once.Do(func() {
			// do nothing
		})
	}()

	time.Sleep(200 * time.Millisecond)
	close(c)
}

// TN: no send possible
func n15() {
	c := make(chan int, 0)
	m := sync.Mutex{}

	go func() {
		time.Sleep(100 * time.Millisecond)
		t := m.TryLock()
		if t {
			c <- 1
			m.Unlock()
		}
	}()

	go func() {
		t := m.TryLock()
		if t {
			<-c
			m.Unlock()
		}
	}()

	time.Sleep(600 * time.Millisecond)
	close(c)
	time.Sleep(300 * time.Millisecond)
}

// TP
func n16() {
	c := make(chan int, 0)
	m := sync.Mutex{}

	go func() {
		t := m.TryLock()
		if t {
			c <- 1
			m.Unlock()
		}
	}()

	go func() {
		<-c
	}()

	time.Sleep(100 * time.Millisecond)
	close(c)
}

// FN
func n17() {
	c := make(chan int, 0)
	m := sync.Mutex{}

	go func() {
		time.Sleep(100 * time.Millisecond)
		t := m.TryLock()
		if t {
			c <- 1
			<-c
			m.Unlock()
		}
	}()

	m.Lock()
	time.Sleep(300 * time.Millisecond)
	close(c)
	m.Unlock()

	time.Sleep(100 * time.Millisecond)
}

// TP
func n18() {
	ch := make(chan int, 1)
	var g sync.WaitGroup

	g.Add(1)

	func() {
		g.Done()
		ch <- 1
	}()

	g.Wait()
	time.Sleep(100 * time.Millisecond)
	close(ch)
}

// FN
func n19() {
	ch := make(chan int, 1)
	m := sync.Mutex{}

	go func() {
		m.Lock()
		ch <- 1
		time.Sleep(100 * time.Millisecond)
		m.Unlock()
	}()

	time.Sleep(100 * time.Millisecond)
	if m.TryLock() {
		close(ch)
		m.Unlock()
	}
	time.Sleep(200 * time.Millisecond)
}

// TP
func n20() {
	ch := make(chan int, 2)

	f := func() {
		ch <- 1
	}

	go func() {
		f()
	}()

	time.Sleep(200 * time.Millisecond)
	close(ch)
}

// ============== Concurrent recv on same channel ==============

// TP
func n21() {
	x := make(chan int)

	go func() {
		<-x
	}()

	go func() {
		<-x
	}()

	x <- 1
	x <- 1

	time.Sleep(300 * time.Millisecond)
}

// TN
func n22() {
	x := make(chan int)

	go func() {
		x <- 1
	}()

	go func() {
		x <- 1
	}()

	<-x
	<-x

	time.Sleep(300 * time.Millisecond)
}

// TN: No concurrent send on same channel
func n23() {
	x := make(chan int, 2)

	go func() {
		x <- 1
		x <- 1
	}()

	<-x
	<-x
}

// TN: No concurrent send on same channel
func n24() {
	x := make(chan int)

	go func() {

		<-x
		<-x
	}()

	x <- 1
	x <- 1
}

// ============== Negative wait counter (Add before done) ==============
// no possible negative wait counter
func n25() {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
}

// possible negative wait counter
func n26() {
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
	}()

	go func() {
		wg.Add(1)
	}()

	go func() {
		time.Sleep(500 * time.Millisecond) // prevent actual negative wait counter
		wg.Done()
	}()

	wg.Done()
}

func n27() {
	var wg sync.WaitGroup
	c := make(chan int, 0)

	go func() {
		wg.Add(1)
		c <- 1
	}()

	<-c
	wg.Done()
}

// ============== Cyclic deadlock ==============
// cyclic deadlock
func n28() {
	m := sync.Mutex{}
	n := sync.Mutex{}

	go func() {
		m.Lock()
		n.Lock()
		n.Unlock()
		m.Unlock()
	}()

	time.Sleep(100 * time.Millisecond) // prevent deadlock
	n.Lock()
	m.Lock()
	m.Unlock()
	n.Unlock()
}

// cyclic deadlock
func n29() {
	m := sync.Mutex{}
	n := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		m.Lock()
		n.Lock()
		n.Unlock()
		m.Unlock()
		c <- 1
	}()

	<-c
	n.Lock()
	m.Lock()
	m.Unlock()
	n.Unlock()
}

// cyclic deadlock
func n30() {
	m := sync.Mutex{}
	n := sync.Mutex{}
	g := sync.Mutex{}

	go func() {
		g.Lock()
		m.Lock()
		n.Lock()
		n.Unlock()
		m.Unlock()
		g.Unlock()
	}()

	g.Lock()
	n.Lock()
	m.Lock()
	m.Unlock()
	n.Unlock()
	g.Unlock()
}

// ============== Mixed Deadlock ==============
func n31() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}()

	m.Lock()
	<-c
	m.Unlock()
}

func n32() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		c <- 1
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		<-c
	}()

	go func() {
		m.Lock()
		<-c
		m.Unlock()
	}()

	time.Sleep(100 * time.Millisecond)
	m.Lock()
	c <- 1
	m.Unlock()
}

func n33() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
	}()

	m.Lock()
	close(c)
	m.Unlock()
}

func n34() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		m.Lock()
		m.Unlock()
		<-c
	}()

	m.Lock()
	c <- 1
	m.Unlock()
}

func n35() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		m.Lock()
		m.Unlock()
		c <- 1
	}()

	time.Sleep(100 * time.Millisecond) // prevent deadlock
	m.Lock()
	<-c
	m.Unlock()
}

func n36() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		m.Lock()
		m.Unlock()
		close(c)
	}()

	time.Sleep(100 * time.Millisecond) // prevent deadlock
	m.Lock()
	<-c
	m.Unlock()
}

func n37() {
	m := sync.Mutex{}
	c := make(chan int, 0)

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Lock()
		m.Unlock()
		<-c
	}()

	m.Lock()
	close(c)
	m.Unlock()
}

func n38() {
	m := sync.Mutex{}
	c := make(chan int, 1)

	go func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}()

	m.Lock()
	<-c
	m.Unlock()
}

// ============== Leaking ==============

func n39() {
	c := make(chan int, 0)

	go func() {
		c <- 1
	}()

	time.Sleep(100 * time.Millisecond)
}

func n40() {
	c := make(chan int, 0)

	go func() {
		<-c
	}()

	time.Sleep(100 * time.Millisecond)
}

func n41() {
	c := make(chan int, 0)

	go func() {
		close(c)
	}()

	time.Sleep(100 * time.Millisecond)
}

func n42() {
	c := make(chan int, 0)

	go func() {
		<-c
	}()

	go func() {
		c <- 1
	}()

	go func() {
		<-c
	}()

	time.Sleep(100 * time.Millisecond)
}

func n43() {
	c := make(chan int, 0)

	go func() {
		<-c
	}()

	go func() {
		c <- 1
	}()

	go func() {
		c <- 1
	}()

	time.Sleep(100 * time.Millisecond)
}

func n44() {
	w := sync.WaitGroup{}

	go func() {
		time.Sleep(100 * time.Millisecond)
		w.Wait()
	}()

	w.Add(1)

	time.Sleep(100 * time.Millisecond)
}

func n45() {
	m := sync.Mutex{}

	go func() {
		m.Lock()
		m.Lock()
	}()

	time.Sleep(100 * time.Millisecond)
}

func main() {

	testCase := flag.Int("c", 0, "Test to run")
	replay := flag.Bool("r", false, "Replay")
	timeout := flag.Int("t", 0, "Timeout")
	flag.Parse()

	if replay != nil && !*replay {
		// init tracing
		runtime.InitAdvocate(0)
		defer advocate.CreateTrace("trace.log")
	} else {
		// init replay
		trace := advocate.ReadTrace("trace.log")
		runtime.EnableReplay(trace)
		defer runtime.WaitForReplayFinish()
	}

	const n = 45
	testFuncs := [n]func(){n01, n02, n03, n04, n05, n06, n07, n08, n09, n10,
		n11, n12, n13, n14, n15, n16, n17, n18, n19, n20,
		n21, n22, n23, n24, n25, n26, n27, n28, n29, n30, n31, n32, n33, n34, n35,
		n36, n37, n38, n39, n40, n41, n42, n43, n44, n45}

	testNames := [n]string{
		"Test 01: N - Synchronous channel",
		"Test 02: N - Wait group",
		"Test 03: N - Once",
		"Test 04: P - RWMutex",
		"Test 05: P - send/recv on closed",
		"Test 06: P - send/recv on closed",
		"Test 07: N - recv/send on closed",
		"Test 08: P - recv on closed",
		"Test 09: P - send/recv on closed",
		"Test 10: P - potential send on closed",
		"Test 11: P - potential send on closed",
		"Test 12: N - no send to closed channel because of once",
		"Test 13: P - potential send to closed channel not recorded because of once",
		"Test 14: P - potential send to closed channel recorded with once",
		"Test 15: N - send on close, no send possible",
		"Test 16: P - send on close",
		"Test 17: P - send on close",
		"Test 18: P - send on close",
		"Test 19: N - send on close",
		"Test 20: P - send on close",
		"Test 21: P - concurrent recv on same channel",
		"Test 22: N - concurrent send on same channel",
		"Test 23: N - no concurrent send on same channel",
		"Test 24: N - no concurrent recv on same channel",
		"Test 25: N - no negative wait counter",
		"Test 26: P - negative wait counter",
		"Test 27: N - no negative wait counter",
		"Test 28: P - cyclic deadlock",
		"Test 29: N - no cyclic deadlock because of channel",
		"Test 30: N - no cyclic deadlock because of guard lock",
		"Test 31: N - no cyclic deadlock because of channel",
		"Test 32: P - Mixed deadlock, MD2-1, send/recv, unbuffered",
		"Test 33: P - Mixed deadlock, MD2-1, close/recv, unbuffered",
		"Test 34: P - Mixed deadlock, MD2-2/3, send/recv, unbuffered",
		"Test 35: P - Mixed deadlock, MD2-2/3, send/recv, unbuffered",
		"Test 36: P - Mixed deadlock, MD2-2/3, close/recv, unbuffered",
		"Test 37: N - No mixed deadlock, MD2-2/3, close/recv, unbuffered",
		"Test 38: P - Mixed deadlock, send/recv, buffered",
		"Test 39: P - Leaking channel send, no alternative",
		"Test 40: P - Leaking channel recv, no alternative",
		"Test 41: N - No leaking channel close",
		"Test 42: P - Leaking channel recv, with alternative",
		"Test 43: P - Leaking channel send, with alternative",
		"Test 44: P - Leaking wait group",
		"Test 45: P - Leaking mutex, doubble locking",
	}

	// cancel test if time has run out
	go func() {
		if timeout != nil && *timeout != 0 {
			time.Sleep(time.Duration(*timeout) * time.Second)
			advocate.CreateTrace("trace.log")
			panic("Timeout")
		}
	}()

	if testCase != nil && *testCase != 0 {
		println(testNames[*testCase-1])
		testFuncs[*testCase-1]()
		time.Sleep(1 * time.Second)
	} else {
		for i := 0; i < n; i++ {
			println(testNames[i])
			testFuncs[i]()
			println("Done: ", i+1, " of ", n)
			time.Sleep(1 * time.Second)
		}
	}
}
