package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include "wiredtiger.h"
*/
import "C"

type Cursor struct {
	w *C.WT_CURSOR
}
