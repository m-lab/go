// +build linux

// Package uuid provides functions that create a consistent globally unique UUID
// for a given TCP socket.
package uuid

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

const (
	// defined in socket.h in the linux kernel
	syscallSoCookie = 57 // syscall.SO_COOKIE does not exist in golang 1.11

	// Whenever there is an error We return this value instead of the empty
	// string. We do this in an effort to detect when client code
	// accidentally uses the returned UUID even when it should not have.
	//
	// This is borne out of past experience, most notably an incident where
	// returning an empty string and an error condition caused the
	// resulting code to create a file named ".gz", which was (thanks to
	// shell-filename-hiding rules) a very hard bug to uncover.  If a file
	// is ever named "INVALID_UUID.gz", it will be much easier to detect
	// that there is a problem versus just ".gz"
	invalidUUID = "INVALID_UUID"
)

var (
	// Only calculate these once - they never change. We use the mtime of /proc as
	// a proxy for the boot time. If the superuser modifies the /proc mount at
	// runtime with something like `sudo touch /proc` while processes using this
	// library are running then this will be wrong.
	cachedPrefixString, cachedPrefixError = getPrefix("/proc")

	// Made into a variable to enable the testing of error handling.
	osHostname = os.Hostname
)

func getBoottime(proxyFile string) (int64, error) {
	// We use the mtime of the passed-in file as a proxy for the boot time.
	//
	// This is potentially brittle, but all existing solutions are worse, as they
	// depend on the stable conversion of the difference of two floating point
	// numbers into an integer, by reading /proc/uptime and then subtracting the
	// first number in that file from time.Now(). If a machine boots up too close
	// to a half-second boundary, then even the old standby `uptime -s` will become
	// inconsistent. On the scale of a single machine boot, that's pretty unlikely,
	// but on M-Lab's scale it will be sure to bite us regularly.
	stat, err := os.Stat(proxyFile)
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
func getPrefix(proxyFile string) (string, error) {
	hostname, err := osHostname()
	if err != nil {
		return "", err
	}
	boottime, err := getBoottime(proxyFile)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%d", hostname, boottime), nil
}

// getCookie returns the cookie (the UUID) associated with a socket. For a given
// boot of a given hostname, this UUID is guaranteed to be unique (until the
// host receives more than 2^64 connections without rebooting).
func getCookie(file *os.File) (uint64, error) {
	var cookie uint64
	cookieLen := uint32(unsafe.Sizeof(cookie))
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
//
// This function will never return the empty string, but the returned string
// value should only be used if the error is nil.
func FromTCPConn(t *net.TCPConn) (string, error) {
	file, err := t.File()
	if err != nil {
		return invalidUUID, err
	}
	defer file.Close()
	return FromFile(file)
}

// FromFile returns a string that is a globally unique identifier for the socket
// represented by the os.File pointer.
//
// This function will never return the empty string, but the returned string
// value should only be used if the error is nil.
func FromFile(file *os.File) (string, error) {
	cookie, err := getCookie(file)
	if err != nil {
		return invalidUUID, err
	}
	return FromCookie(cookie)
}

// FromCookie returns a string that is a globally unique identifier for the
// passed-in socket cookie.
//
// This function will never return the empty string, but the returned string
// value should only be used if the error is nil.
func FromCookie(cookie uint64) (string, error) {
	if cachedPrefixError != nil {
		return invalidUUID, cachedPrefixError
	}
	return fmt.Sprintf("%s_%016X", cachedPrefixString, cookie), nil
}
