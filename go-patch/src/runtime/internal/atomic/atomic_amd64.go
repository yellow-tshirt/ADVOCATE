// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic

import "unsafe"

// Export some functions via linkname to assembly in sync/atomic.
//
//go:linkname Load
// ADVOCATE-CHANGE-START
//go:linkname LoadAdvocate
//go:linkname Load64Advocate
//go:linkname LoadpAdvocate
// ADVOCATE-CHANGE-END
//go:linkname Loadp
//go:linkname Load64

//go:nosplit
//go:noinline
func Load(ptr *uint32) uint32 {
	return *ptr
}

// ADVOCATE-CHANGE-START
//
//go:nosplit
//go:noinline
func LoadAdvocate(ptr *uint32) uint32 {
	AdvocateAtomic32Load(ptr)
	return *ptr
}

// ADVOCATE-CHANGE-END

//go:nosplit
//go:noinline
func Loadp(ptr unsafe.Pointer) unsafe.Pointer {
	return *(*unsafe.Pointer)(ptr)
}

// ADVOCATE-CHANGE-START
//
//go:nosplit
//go:noinline
func LoadpAdvocate(ptr unsafe.Pointer) unsafe.Pointer {
	AdvocateAtomicPtr(ptr)
	return *(*unsafe.Pointer)(ptr)
}

// ADVOCATE-CHANGE-END

//go:nosplit
//go:noinline
func Load64(ptr *uint64) uint64 {
	return *ptr
}

// ADVOCATE-CHANGE-START
//
//go:nosplit
//go:noinline
func Load64Advocate(ptr *uint64) uint64 {
	AdvocateAtomic64Load(ptr)
	return *ptr
}

// ADVOCATE-CHANGE-END

//go:nosplit
//go:noinline
func LoadAcq(ptr *uint32) uint32 {
	return *ptr
}

//go:nosplit
//go:noinline
func LoadAcq64(ptr *uint64) uint64 {
	return *ptr
}

//go:nosplit
//go:noinline
func LoadAcquintptr(ptr *uintptr) uintptr {
	return *ptr
}

//go:noescape
func Xadd(ptr *uint32, delta int32) uint32

//go:noescape
func Xadd64(ptr *uint64, delta int64) uint64

//go:noescape
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr

//go:noescape
func Xchg(ptr *uint32, new uint32) uint32

//go:noescape
func Xchg64(ptr *uint64, new uint64) uint64

//go:noescape
func Xchguintptr(ptr *uintptr, new uintptr) uintptr

//go:nosplit
//go:noinline
func Load8(ptr *uint8) uint8 {
	return *ptr
}

//go:noescape
func And8(ptr *uint8, val uint8)

//go:noescape
func Or8(ptr *uint8, val uint8)

//go:noescape
func And(ptr *uint32, val uint32)

//go:noescape
func Or(ptr *uint32, val uint32)

// NOTE: Do not add atomicxor8 (XOR is not idempotent).

//go:noescape
func Cas64(ptr *uint64, old, new uint64) bool

//go:noescape
func CasRel(ptr *uint32, old, new uint32) bool

//go:noescape
func Store(ptr *uint32, val uint32)

//go:noescape
func Store8(ptr *uint8, val uint8)

//go:noescape
func Store64(ptr *uint64, val uint64)

//go:noescape
func StoreRel(ptr *uint32, val uint32)

//go:noescape
func StoreRel64(ptr *uint64, val uint64)

//go:noescape
func StoreReluintptr(ptr *uintptr, val uintptr)

// StorepNoWB performs *ptr = val atomically and without a write
// barrier.
//
// NO go:noescape annotation; see atomic_pointer.go.
func StorepNoWB(ptr unsafe.Pointer, val unsafe.Pointer)
