// +build linux

// Package uuid provides functions that create a consistent globally unique UUID
// for a given TCP socket.
package uuid

import (
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

const (
	// defined in socket.h in the linux kernel
	syscallSoCookie = 57 // syscall.SO_COOKIE does not exist in golang 1.11
)

var cachedPrefixString = ""
var cacheMutex = sync.Mutex{}

func getBoottime() (int64, error) {
	// We use the mtime of /proc as a proxy for the boot time. If the superuser
	// modifies the /proc mount at runtime with something like `sudo touch /proc`
	// while processes using this library are running then this will be wrong.
	//
	// All existing solutions are worse, however, as they depend on the stable
	// conversion of the difference of two floating point numbers into an integer,
	// by reading /proc/uptime and then subtracting the first number in that file
	// from time.Now(). If a machine boots up too close to a half-second boundary,
	// then even the old standby `uptime -s` will become inconsistent. On the scale
	// of a single machine boot, that's pretty unlikely, but on M-Lab's scale it
	// will be sure to bite us regularly.
	stat, err := os.Stat("/proc")
	if err != nil {
		return 0, err
	}
	return stat.ModTime().Unix(), err
}

// getPrefix returns a prefix string which contains the hostname and boot time
// of the machine, which globally uniquely identifies the socket uuid namespace.
// This function is cached because that pair should be constant for a given
// instance of the program, unless the boot time changes (how?) or the hostname
// changes (why?) while this program is running.
func getPrefix() (string, error) {
	if cachedPrefixString == "" {
		cacheMutex.Lock()
		defer cacheMutex.Unlock()
		// Check twice because check-then-lock has an implicit race condition. We use
		// check-then-lock because it means that the common path needs no mutex
		// locking, which is nice. This code also ensures that we only stop trying to
		// set up the prefix when we don't get an error.
		if cachedPrefixString == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return "", err
			}
			boottime, err := getBoottime()
			if err != nil {
				return "", err
			}
			cachedPrefixString = fmt.Sprintf("%s_%d", hostname, boottime)
		}
	}
	return cachedPrefixString, nil
}

// getCookie returns the cookie (the UUID) associated with a socket. For a given
// boot of a given hostname, this UUID is guaranteed to be unique (until the
// host receives more than 2^64 connections without rebooting).
func getCookie(t *net.TCPConn) (uint64, error) {
	var cookie uint64
	cookieLen := uint32(unsafe.Sizeof(cookie))
	file, err := t.File()
	if err != nil {
		return 0, err
	}
	defer file.Close()
	// GetsockoptInt does not work for 64 bit integers, which is what the UUID is.
	// So we crib from the GetsockoptInt implementation and ndt-server/tcpinfox,
	// and call the syscall manually.
	_, _, errno := syscall.Syscall6(
		uintptr(syscall.SYS_GETSOCKOPT),
		uintptr(int(file.Fd())),
		uintptr(syscall.SOL_SOCKET),
		uintptr(syscallSoCookie),
		uintptr(unsafe.Pointer(&cookie)),
		uintptr(unsafe.Pointer(&cookieLen)),
		uintptr(0))

	if errno != 0 {
		return 0, fmt.Errorf("Error in Getsockopt. Errno=%d", errno)
	}
	return cookie, nil
}

// FromTCPConn returns a string that is a globally unique identifier for the
// socket held by the passed-in TCPConn (assuming hostnames are unique).
func FromTCPConn(t *net.TCPConn) (string, error) {
	cookie, err := getCookie(t)
	if err != nil {
		return "", err
	}
	return FromCookie(cookie)
}

// FromCookie returns a string that is a globally unique identifier for the
// passed-in socket cookie.
func FromCookie(cookie uint64) (string, error) {
	prefix, err := getPrefix()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%016X", prefix, cookie), nil
}
