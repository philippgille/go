// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm

package atomic

import (
	"runtime/internal/sys"
	"unsafe"
)

type spinlock struct {
	v uint32
}

//go:nosplit
func (l *spinlock) lock() {
	for {
		if Cas(&l.v, 0, 1) {
			return
		}
	}
}

//go:nosplit
func (l *spinlock) unlock() {
	Store(&l.v, 0)
}

var locktab [57]struct {
	l   spinlock
	pad [sys.CacheLineSize - unsafe.Sizeof(spinlock{})]byte
}

func addrLock(addr *uint64) *spinlock {
	return &locktab[(uintptr(unsafe.Pointer(addr))>>3)%uintptr(len(locktab))].l
}

// Atomic add and return new value.
//go:nosplit
func Xadd(val *uint32, delta int32) uint32 {
	for {
		oval := *val
		nval := oval + uint32(delta)
		if Cas(val, oval, nval) {
			return nval
		}
	}
}

//go:noescape
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr

//go:nosplit
func Xchg(addr *uint32, v uint32) uint32 {
	for {
		old := *addr
		if Cas(addr, old, v) {
			return old
		}
	}
}

//go:nosplit
func Xchguintptr(addr *uintptr, v uintptr) uintptr {
	return uintptr(Xchg((*uint32)(unsafe.Pointer(addr)), uint32(v)))
}

//go:nosplit
func StorepNoWB(addr unsafe.Pointer, v unsafe.Pointer) {
	for {
		old := *(*unsafe.Pointer)(addr)
		if Casp1((*unsafe.Pointer)(addr), old, v) {
			return
		}
	}
}

//go:nosplit
func Store(addr *uint32, v uint32) {
	for {
		old := *addr
		if Cas(addr, old, v) {
			return
		}
	}
}

//go:nosplit
func Cas64(addr *uint64, old, new uint64) bool {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var ok bool
	addrLock(addr).lock()
	if *addr == old {
		*addr = new
		ok = true
	}
	addrLock(addr).unlock()
	return ok
}

//go:nosplit
func Xadd64(addr *uint64, delta int64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr + uint64(delta)
	*addr = r
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func Xchg64(addr *uint64, v uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	*addr = v
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func Load64(addr *uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	addrLock(addr).unlock()
	return r
}

//go:nosplit
func Store64(addr *uint64, v uint64) {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	addrLock(addr).lock()
	*addr = v
	addrLock(addr).unlock()
}

//go:nosplit
func Or8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8) // little endian
	for {
		old := *addr32
		if Cas(addr32, old, old|word) {
			return
		}
	}
}

//go:nosplit
func And8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8)    // little endian
	mask := uint32(0xFF) << ((uaddr & 3) * 8) // little endian
	word |= ^mask
	for {
		old := *addr32
		if Cas(addr32, old, old&word) {
			return
		}
	}
}

//go:nosplit
func armcas(ptr *uint32, old, new uint32) bool

//go:noescape
func Load(addr *uint32) uint32

//go:noescape
func Loadp(addr unsafe.Pointer) unsafe.Pointer
