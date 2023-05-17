/*	Copyright (c) 2019, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"sync/atomic"

	"github.com/ErikKassubek/deadlockDetectorGo/src/dedego"
	"github.com/jfcg/sixb"
)

func isSortedF4(slc []float32) int {
	l, h := 0, len(slc)-1
	if NaNoption == NaNlarge {
		for ; l <= h; h-- {
			if x := slc[h]; x == x {
				break
			}
		}
	} else if NaNoption == NaNsmall {
		for ; l <= h; l++ {
			if x := slc[l]; x == x {
				break
			}
		}
	}

	for i := h; i > l; i-- {
		if !(slc[i] >= slc[i-1]) {
			return i
		}
	}
	return 0
}

func insertionF4(slc []float32) {
	for h := 1; h < len(slc); h++ {
		l, val := h, slc[h]
		var pre float32
		goto start
	loop:
		slc[l] = pre
		l--
		if l == 0 {
			goto last
		}
	start:
		pre = slc[l-1]
		if val < pre {
			goto loop
		}
		if l == h {
			continue
		}
	last:
		slc[l] = val
	}
}

func pivotF4(slc []float32, n uint) float32 {

	first, step, _ := minMaxSample(uint(len(slc)), n)

	var sample [nsConc - 1]float32
	for i := int(n - 1); i >= 0; i-- {
		sample[i] = slc[first]
		first += step
	}
	insertionF4(sample[:n])

	return sample[n>>1]
}

func partOneF4(slc []float32, pv float32) int {
	l, h := 0, len(slc)-1
	goto start
second:
	for {
		h--
		if h <= l {
			return l
		}
		if slc[h] <= pv {
			break
		}
	}
swap:
	slc[l], slc[h] = slc[h], slc[l]
next:
	l++
	h--
start:
	if h <= l {
		goto last
	}

	if pv <= slc[h] {
		if pv < slc[l] {
			goto second
		}
		goto next
	}
	for {
		if pv <= slc[l] {
			goto swap
		}
		l++
		if h <= l {
			return l + 1
		}
	}
last:
	if l == h && slc[h] < pv {
		l++
	}
	return l
}

func partTwoF4(slc []float32, l, h int, pv float32) int {
	l--
	if h <= l {
		return -1
	}
	goto start
second:
	for {
		h++
		if h >= len(slc) {
			return l
		}
		if slc[h] <= pv {
			break
		}
	}
swap:
	slc[l], slc[h] = slc[h], slc[l]
next:
	l--
	h++
start:
	if l < 0 {
		return h
	}
	if h >= len(slc) {
		return l
	}

	if pv <= slc[h] {
		if pv < slc[l] {
			goto second
		}
		goto next
	}
	for {
		if pv <= slc[l] {
			goto swap
		}
		l--
		if l < 0 {
			return h
		}
	}
}

func gPartOneF4(ar []float32, pv float32, ch dedego.Chan[int]) {
	ch.Send(partOneF4(ar, pv))
}

func partConF4(slc []float32, ch dedego.Chan[int]) int {

	pv := pivotF4(slc, nsConc-1)
	mid := len(slc) >> 1
	l, h := mid>>1, sixb.MeanI(mid, len(slc))
	func() {
		DedegoRoutineIndex := dedego.SpawnPre()
		go func() {
			dedego.SpawnPost(DedegoRoutineIndex)
			{

				gPartOneF4(slc[l:h:h], pv, ch)
			}
		}()
	}()

	r := partTwoF4(slc, l, h, pv)

	k := l + ch.Receive()

	if r < mid {
		for ; 0 <= r; r-- {
			if pv < slc[r] {
				k--
				slc[r], slc[k] = slc[k], slc[r]
			}
		}
	} else {
		for ; r < len(slc); r++ {
			if slc[r] < pv {
				slc[r], slc[k] = slc[k], slc[r]
				k++
			}
		}
	}
	return k
}

func shortF4(ar []float32) {
start:
	first, step, last := minMaxSample(uint(len(ar)), 3)
	f, pv, l := ar[first], ar[first+step], ar[last]

	if pv < f {
		pv, f = f, pv
	}
	if l < pv {
		if l < f {
			pv = f
		} else {
			pv = l
		}
	}

	k := partOneF4(ar, pv)
	var aq []float32

	if k < len(ar)-k {
		aq = ar[:k:k]
		ar = ar[k:]
	} else {
		aq = ar[k:]
		ar = ar[:k:k]
	}

	if len(aq) > MaxLenIns {
		shortF4(aq)
		goto start
	}
isort:
	insertionF4(aq)

	if len(ar) > MaxLenIns {
		goto start
	}
	if &ar[0] != &aq[0] {
		aq = ar
		goto isort
	}
}

func gLongF4(ar []float32, sv *syncVar) {
	longF4(ar, sv)

	if atomic.AddUint64(&sv.nGor, ^uint64(0)) == 0 {
		sv.done.Send(0)

	}
}

func longF4(ar []float32, sv *syncVar) {
start:
	pv := pivotF4(ar, nsLong-1)
	k := partOneF4(ar, pv)
	var aq []float32

	if k < len(ar)-k {
		aq = ar[:k:k]
		ar = ar[k:]
	} else {
		aq = ar[k:]
		ar = ar[:k:k]
	}

	if len(aq) <= MaxLenRec {

		if len(aq) > MaxLenIns {
			shortF4(aq)
		} else {
			insertionF4(aq)
		}

		if len(ar) > MaxLenRec {
			goto start
		}
		shortF4(ar)
		return
	}

	if sv == nil || gorFull(sv) {
		longF4(aq, sv)
		goto start
	}

	atomic.AddUint64(&sv.nGor, 1)
	func() {
		DedegoRoutineIndex := dedego.SpawnPre()
		go func() {
			dedego.SpawnPost(DedegoRoutineIndex)
			{
				gLongF4(ar, sv)
			}
		}()
	}()
	ar = aq
	goto start
}

func sortF4(ar []float32) {
	l, h := 0, len(ar)-1
	if NaNoption == NaNlarge {
		for l <= h {
			x := ar[h]
			if x != x {
				h--
				continue
			}
			y := ar[l]
			if y != y {
				ar[l], ar[h] = x, y
				h--
			}
			l++
		}
		ar = ar[:h+1]
	} else if NaNoption == NaNsmall {
		for l <= h {
			y := ar[l]
			if y != y {
				l++
				continue
			}
			x := ar[h]
			if x != x {
				ar[l], ar[h] = x, y
				l++
			}
			h--
		}
		ar = ar[l:]
	}

	if len(ar) < 2*(MaxLenRec+1) || MaxGor <= 1 {

		if len(ar) > MaxLenRec {
			longF4(ar, nil)
		} else if len(ar) > MaxLenIns {
			shortF4(ar)
		} else {
			insertionF4(ar)
		}
		return
	}

	sv := syncVar{1,
		dedego.NewChan[int](int(0))}
	for {

		k := partConF4(ar, sv.done)
		var aq []float32

		if k < len(ar)-k {
			aq = ar[:k:k]
			ar = ar[k:]
		} else {
			aq = ar[k:]
			ar = ar[:k:k]
		}

		if len(aq) > MaxLenRec {
			atomic.AddUint64(&sv.nGor, 1)
			func() {
				DedegoRoutineIndex := dedego.SpawnPre()
				go func() {
					dedego.SpawnPost(DedegoRoutineIndex)
					{
						gLongF4(aq, &sv)
					}
				}()
			}()

		} else if len(aq) > MaxLenIns {
			shortF4(aq)
		} else {
			insertionF4(aq)
		}

		if len(ar) < 2*(MaxLenRec+1) || gorFull(&sv) {
			break
		}

	}

	longF4(ar, &sv)

	if atomic.AddUint64(&sv.nGor, ^uint64(0)) != 0 {
		sv.done.Receive()

	}
}
