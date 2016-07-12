package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include "wiredtiger.h"


*/
import "C"
import "unsafe"

func Open(home, config string) (*Connection, error) {
	var w *C.WT_CONNECTION

	homeC := C.CString(home)
	configC := C.CString(config)

	result := int(C.wiredtiger_open(homeC, nil, configC, &w))

	C.free(unsafe.Pointer(homeC))
	C.free(unsafe.Pointer(configC))

	if result == 0 {
		conn := new(Connection)
		conn.w = w
		return conn, nil
	}

	return nil, NewError(result, nil)
}

func Version() (vertxt string, major int, minor int, patch int) {
	var a, b, c C.int

	vertxt = C.GoString(C.wiredtiger_version(&a, &b, &c))
	major = int(a)
	minor = int(b)
	patch = int(c)

	return
}

func strerror(r int) string {
	return C.GoString(C.wiredtiger_strerror(C.int(r)))
}

type Error struct {
	Code int
	Text string
}

func NewError(code int, s *Session) error {
	/*	switch code {
		case WT_DUPLICATE_KEY, WT_NOTFOUND, WT_ROLLBACK, WT_RUN_RECOVERY:
		default:
			panic(code)
		}*/

	if s != nil {
		return &Error{code, s.strerror(code)}
	} else {
		return &Error{code, strerror(code)}
	}
}

func (e *Error) Error() string {
	return e.Text
}

func IsRollbackErr(e error) bool {
	if v, ok := e.(*Error); ok {
		return v.Code == int(C.WT_ROLLBACK)
	}

	return false
}

func IsDupKeyErr(e error) bool {
	if v, ok := e.(*Error); ok {
		return v.Code == int(C.WT_DUPLICATE_KEY)
	}

	return false
}

func IsNotFoundErr(e error) bool {
	if v, ok := e.(*Error); ok {
		return v.Code == int(C.WT_NOTFOUND)
	}

	return false
}

func IsRunRecoveryErr(e error) bool {
	if v, ok := e.(*Error); ok {
		return v.Code == int(C.WT_RUN_RECOVERY)
	}

	return false
}
