package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include "wiredtiger.h"


int wiredtiger_connection_close(WT_CONNECTION *connection, const char *config) {
	return connection->close(connection, config);
}

int wiredtiger_connection_reconfigure(WT_CONNECTION *connection, const char *config) {
	return connection->reconfigure(connection, config);
}

const char *wiredtiger_connection_get_home(WT_CONNECTION *connection) {
	return connection->get_home(connection);
}

int wiredtiger_connection_configure_method(WT_CONNECTION *connection, const char *method, const char *uri, const char *config, const char *type, const char *check) {
	return connection->configure_method(connection, method, uri, config, type, check);
}

int wiredtiger_connection_is_new(WT_CONNECTION *connection) {
	return connection->is_new(connection);
}

int wiredtiger_connection_open_session(WT_CONNECTION *connection, WT_EVENT_HANDLER *errhandler,	const char *config, WT_SESSION **sessionp) {
	return connection->open_session(connection, errhandler, config, sessionp);
}
*/
import "C"
import "unsafe"

type Connection struct {
	w *C.WT_CONNECTION
}

// General
func (c *Connection) Close(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_connection_close(c.w, configC)); res == 0 {
		c.w = nil
		return nil
	} else {
		return NewError(res, nil)
	}
}

func (c *Connection) Reconfigure(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_connection_reconfigure(c.w, configC)); res != 0 {
		return NewError(res, nil)
	}

	return nil
}

func (c *Connection) GetHome() string {
	return C.GoString(C.wiredtiger_connection_get_home(c.w))
}

func (c *Connection) ConfigureMethod(method, uri, config, mtype, check string) error {
	var methodC *C.char = nil
	var uriC *C.char = nil
	var configC *C.char = nil
	var mtypeC *C.char = nil
	var checkC *C.char = nil

	if len(method) > 0 {
		methodC = C.CString(method)
		defer C.free(unsafe.Pointer(methodC))
	}

	if len(uri) > 0 {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if len(mtype) > 0 {
		mtypeC = C.CString(mtype)
		defer C.free(unsafe.Pointer(mtypeC))
	}

	if len(check) > 0 {
		checkC = C.CString(check)
		defer C.free(unsafe.Pointer(checkC))
	}

	if res := int(C.wiredtiger_connection_configure_method(c.w, methodC, uriC, configC, mtypeC, checkC)); res != 0 {
		return NewError(res, nil)
	}

	return nil
}

func (c *Connection) IsNew() bool {
	if res := int(C.wiredtiger_connection_is_new(c.w)); res == 0 {
		return false
	}

	return true
}

// Session handles
func (c *Connection) OpenSession(config string) (*Session, error) {
	var w *C.WT_SESSION
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_connection_open_session(c.w, nil, configC, &w)); res != 0 {
		return nil, NewError(res, nil)
	}

	newsession := new(Session)
	newsession.w = w
	newsession.conn = c

	return newsession, nil
}

// TODO: Extension
