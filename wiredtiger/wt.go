package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include "wiredtiger.h"


*/
import "C"
import "unsafe"

func Open(home, config string) (conn *Connection, result int) {
	var w *C.WT_CONNECTION

	homeC := C.CString(home)
	configC := C.CString(config)

	result = int(C.wiredtiger_open(homeC, nil, configC, &w))

	C.free(unsafe.Pointer(homeC))
	C.free(unsafe.Pointer(configC))

	if result == 0 {
		conn = new(Connection)
		conn.w = w
		return
	}

	return
}

func Version() (vertxt string, major int, minor int, patch int) {
	var a, b, c C.int

	vertxt = C.GoString(C.wiredtiger_version(&a, &b, &c))
	major = int(a)
	minor = int(b)
	patch = int(c)

	return
}

func Error(r int) string {
	return C.GoString(C.wiredtiger_strerror(C.int(r)))
}
