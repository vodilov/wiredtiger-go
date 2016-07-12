package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include "wiredtiger.h"

int wiredtiger_session_close(WT_SESSION *session, const char *config) {
	return session->close(session, config);
}

int wiredtiger_session_reconfigure(WT_SESSION *session, const char *config) {
	return session->reconfigure(session, config);
}

const char *wiredtiger_session_strerror(WT_SESSION *session, int error) {
	return session->strerror(session, error);
}

int wiredtiger_session_open_cursor(WT_SESSION *session, const char *uri, WT_CURSOR *to_dup, const char *config, WT_CURSOR **cursorp) {
	int ret;

	if(ret = session->open_cursor(session, uri, to_dup, config, cursorp))
		return ret;

	if (((*cursorp)->flags & WT_CURSTD_DUMP_JSON) == 0)
			(*cursorp)->flags |= WT_CURSTD_RAW;

	return 0;
}

int wiredtiger_session_create(WT_SESSION *session, const char *name, const char *config) {
	return session->create(session, name, config);
}

int wiredtiger_session_compact(WT_SESSION *session, const char *name, const char *config) {
	return session->compact(session, name, config);
}

int wiredtiger_session_drop(WT_SESSION *session, const char *name, const char *config) {
	return session->drop(session, name, config);
}

int wiredtiger_session_join(WT_SESSION *session, WT_CURSOR *join_cursor, WT_CURSOR *ref_cursor, const char *config) {
	return session->join(session, join_cursor, ref_cursor, config);
}

int wiredtiger_session_log_flush(WT_SESSION *session, const char *config) {
	return session->log_flush(session, config);
}

int wiredtiger_session_log_insert_message(WT_SESSION *session, const char *message) {
	if(!message)
		return 0;

	return session->log_printf(session, "%s", message);
}

int wiredtiger_session_rebalance(WT_SESSION *session, const char *uri, const char *config) {
	return session->rebalance(session, uri, config);
}

int wiredtiger_session_rename(WT_SESSION *session, const char *uri, const char *newuri, const char *config) {
	return session->rename(session, uri, newuri, config);
}

int wiredtiger_session_reset(WT_SESSION *session) {
	return session->reset(session);
}

int wiredtiger_session_salvage(WT_SESSION *session, const char *name, const char *config) {
	return session->salvage(session, name, config);
}

int wiredtiger_session_truncate(WT_SESSION *session, const char *name, WT_CURSOR *start, WT_CURSOR *stop, const char *config) {
	return session->truncate(session, name, start, stop, config);
}

int wiredtiger_session_upgrade(WT_SESSION *session, const char *name, const char *config) {
	return session->upgrade(session, name, config);
}

int wiredtiger_session_verify(WT_SESSION *session, const char *name, const char *config) {
	return session->verify(session, name, config);
}

int wiredtiger_session_begin_transaction(WT_SESSION *session, const char *config) {
	return session->begin_transaction(session, config);
}

int wiredtiger_session_commit_transaction(WT_SESSION *session, const char *config) {
	return session->commit_transaction(session, config);
}

int wiredtiger_session_rollback_transaction(WT_SESSION *session, const char *config) {
	return session->rollback_transaction(session, config);
}

int wiredtiger_session_checkpoint(WT_SESSION *session, const char *config) {
	return session->checkpoint(session, config);
}

int wiredtiger_session_snapshot(WT_SESSION *session, const char *config) {
	return session->snapshot(session, config);
}

int wiredtiger_session_transaction_pinned_range(WT_SESSION *session, uint64_t *range) {
	return session->transaction_pinned_range(session, range);
}

int wiredtiger_session_transaction_sync(WT_SESSION *session, const char *config) {
	return session->transaction_sync(session, config);
}

*/
import "C"
import "unsafe"

type Session struct {
	w    *C.WT_SESSION
	conn *Connection
}

// General

func (s *Session) Close(config string) error {
	var configC *C.char

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	result := int(C.wiredtiger_session_close(s.w, configC))

	if result == 0 {
		s.w = nil
		return nil
	}

	return NewError(result, nil)
}

func (s *Session) Reconfigure(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	res := int(C.wiredtiger_session_reconfigure(s.w, configC))

	if res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) strerror(errnum int) string {
	txt := C.wiredtiger_session_strerror(s.w, C.int(errnum))
	if txt != nil {
		return C.GoString(C.wiredtiger_session_strerror(s.w, C.int(errnum)))
	} else {
		return ""
	}

}

func (s *Session) GetConnection() *Connection {
	return s.conn
}

// Cursor handles

func (s *Session) OpenCursor(uri string, to_dup *Cursor, config string) (*Cursor, error) {
	var w *C.WT_CURSOR
	var uriC *C.char = nil
	var configC *C.char = nil
	var wc *C.WT_CURSOR = nil

	if len(uri) > 0 {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if to_dup != nil {
		wc = to_dup.w
	}

	res := int(C.wiredtiger_session_open_cursor(s.w, uriC, wc, configC, &w))

	if res == 0 {
		newcursor := new(Cursor)
		newcursor.w = w
		newcursor.session = s
		newcursor.uri = C.GoString(w.uri)
		newcursor.keyFormat = C.GoString(w.key_format)
		newcursor.valueFormat = C.GoString(w.value_format)

		return newcursor, nil
	}

	return nil, NewError(res, s)
}

// Table operations

func (s *Session) Create(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_create(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Compact(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_compact(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Drop(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_drop(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Join(join_cursor *Cursor, ref_cursor *Cursor, config string) error {
	var w_join_cursor, w_ref_cursor *C.WT_CURSOR
	var configC *C.char

	if join_cursor != nil {
		w_join_cursor = join_cursor.w
	}

	if ref_cursor != nil {
		w_ref_cursor = ref_cursor.w
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_join(s.w, w_join_cursor, w_ref_cursor, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) LogFlush(config string) error {
	var configC *C.char

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_log_flush(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) LogInsertMessage(message string) error {
	var messageC *C.char

	if len(message) > 0 {
		messageC = C.CString(message)
		defer C.free(unsafe.Pointer(messageC))
	} else {
		return nil
	}

	if res := int(C.wiredtiger_session_log_insert_message(s.w, messageC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Rebalance(uri, config string) error {
	var uriC *C.char
	var configC *C.char

	if len(uri) > 0 {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_rebalance(s.w, uriC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Rename(uri, newuri, config string) error {
	var uriC *C.char
	var newuriC *C.char
	var configC *C.char

	if len(uri) > 0 {
		uriC = C.CString(uri)
		defer C.free(unsafe.Pointer(uriC))
	}

	if len(newuri) > 0 {
		newuriC = C.CString(newuri)
		defer C.free(unsafe.Pointer(newuriC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_rename(s.w, uriC, newuriC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Reset() error {
	if res := int(C.wiredtiger_session_reset(s.w)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Salvage(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_salvage(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Truncate(name string, start *Cursor, stop *Cursor, config string) error {
	var sc, ec *C.WT_CURSOR
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if start != nil {
		sc = start.w
	}

	if stop != nil {
		ec = stop.w
	}

	if res := int(C.wiredtiger_session_truncate(s.w, nameC, sc, ec, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Upgrade(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_upgrade(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Verify(name, config string) error {
	var nameC *C.char
	var configC *C.char

	if len(name) > 0 {
		nameC = C.CString(name)
		defer C.free(unsafe.Pointer(nameC))
	}

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_verify(s.w, nameC, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

// Transactions

func (s *Session) BeginTransaction(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_begin_transaction(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) CommitTransaction(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_commit_transaction(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) RollbackTransaction(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_rollback_transaction(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Checkpoint(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_checkpoint(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) Snapshot(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_snapshot(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}

func (s *Session) TransactionPinnedRange(pined_range *uint64) error {
	var pined_rangeC C.uint64_t

	if res := int(C.wiredtiger_session_transaction_pinned_range(s.w, &pined_rangeC)); res != 0 {
		return NewError(res, s)
	}

	*pined_range = uint64(pined_rangeC)
	return nil
}

func (s *Session) TransactionSync(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_session_transaction_sync(s.w, configC)); res != 0 {
		return NewError(res, s)
	}

	return nil
}
