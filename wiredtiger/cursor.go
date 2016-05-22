package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include "wiredtiger.h"

int wiredtiger_cursor_close(WT_CURSOR *cursor) {
	return cursor->close(cursor);
}

int wiredtiger_cursor_reconfigure(WT_CURSOR *cursor, const char *config) {
	return cursor->reconfigure(cursor, config);
}

// TODO: data access


int wiredtiger_cursor_compare(WT_CURSOR *cursor, WT_CURSOR *other, int *comparep) {
	return cursor->compare(cursor, other, comparep);
}

int wiredtiger_cursor_equals(WT_CURSOR *cursor, WT_CURSOR *other, int *equalp) {
	return cursor->compare(cursor, other, equalp);
}

int wiredtiger_cursor_next(WT_CURSOR *cursor) {
	return cursor->next(cursor);
}

int wiredtiger_cursor_prev(WT_CURSOR *cursor) {
	return cursor->prev(cursor);
}

int wiredtiger_cursor_reset(WT_CURSOR *cursor) {
	return cursor->reset(cursor);
}

int wiredtiger_cursor_search(WT_CURSOR *cursor) {
	return cursor->search(cursor);
}

int wiredtiger_cursor_search_near(WT_CURSOR *cursor, int *exactp) {
	return cursor->search_near(cursor, exactp);
}

int wiredtiger_cursor_insert(WT_CURSOR *cursor) {
	return cursor->insert(cursor);
}

int wiredtiger_cursor_update(WT_CURSOR *cursor) {
	return cursor->update(cursor);
}

int wiredtiger_cursor_remove(WT_CURSOR *cursor) {
	return cursor->remove(cursor);
}
*/
import "C"
import "unsafe"

type Cursor struct {
	w           *C.WT_CURSOR
	Session     *Session
	URI         string
	KeyFormat   string
	ValueFormat string
}

// General

func (c *Cursor) Close() int {
	result := int(C.wiredtiger_cursor_close(c.w))

	if result == 0 {
		c.w = nil
	}

	return result
}

func (c *Cursor) Reconfigure(config string) int {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	return int(C.wiredtiger_cursor_reconfigure(c.w, configC))
}

// Data access
// TODO: implement

func (c *Cursor) GetKey(a ...interface{}) int {
	return -1
}

func (c *Cursor) GetValue(a ...interface{}) int {
	return -1
}

func (c *Cursor) SetKey(a ...interface{}) int {
	return -1
}

func (c *Cursor) SetValue(a ...interface{}) int {
	return -1
}

// Cursor positioning

func (c *Cursor) Compare(other *Cursor) (compare_result, result int) {
	var oc *C.WT_CURSOR
	var compare_resultC C.int

	if other != nil {
		oc = other.w
	}

	result = int(C.wiredtiger_cursor_compare(c.w, oc, &compare_resultC))

	if result == 0 {
		compare_result = int(compare_resultC)
	}

	return
}

func (c *Cursor) Equals(other *Cursor) (compare_result bool, result int) {
	var oc *C.WT_CURSOR
	var compare_resultC C.int

	if other != nil {
		oc = other.w
	}

	result = int(C.wiredtiger_cursor_equals(c.w, oc, &compare_resultC))

	if result == 0 {
		if compare_resultC == 0 {
			compare_result = false
		} else {
			compare_result = true
		}
	}

	return
}

func (c *Cursor) Next() int {
	return int(C.wiredtiger_cursor_next(c.w))
}

func (c *Cursor) Prev() int {
	return int(C.wiredtiger_cursor_prev(c.w))
}

func (c *Cursor) Reset() int {
	return int(C.wiredtiger_cursor_reset(c.w))
}

func (c *Cursor) Search() int {
	return int(C.wiredtiger_cursor_search(c.w))
}

func (c *Cursor) SearchNear() (compare_result, result int) {
	var compare_resultC C.int

	result = int(C.wiredtiger_cursor_search_near(c.w, &compare_resultC))

	if result == 0 {
		compare_result = int(compare_resultC)
	}

	return
}

// Data modification

func (c *Cursor) Insert() int {
	return int(C.wiredtiger_cursor_insert(c.w))
}

func (c *Cursor) Update() int {
	return int(C.wiredtiger_cursor_update(c.w))
}

func (c *Cursor) Remove() int {
	return int(C.wiredtiger_cursor_remove(c.w))
}
