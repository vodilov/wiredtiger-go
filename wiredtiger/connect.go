package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include "wiredtiger.h"

int wiredtiger_connection_open_session(WT_CONNECTION *connection, WT_EVENT_HANDLER *errhandler,	const char *config, WT_SESSION **sessionp) {
	return connection->open_session(connection, errhandler, config, sessionp);
}
*/
import "C"

type Connection struct {
	w *C.WT_CONNECTION
}

func (c *Connection) OpenSession(config string) (cursor *Session, result int) {
	var w *C.WT_SESSION

	cursor = nil
	result = int(C.wiredtiger_connection_open_session(c.w, nil, C.CString(config), &w))

	return
}
