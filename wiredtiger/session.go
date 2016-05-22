package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include "wiredtiger.h"
*/
import "C"

type Session struct {
	w          *C.WT_SESSION
	Connection *Connection
	AppPrivate interface{}
}
