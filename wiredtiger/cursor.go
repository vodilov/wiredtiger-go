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

int wiredtiger_cursor_get_key(WT_CURSOR *cursor, WT_ITEM *v) {
	return cursor->get_key(cursor, v);
}

int wiredtiger_cursor_get_value(WT_CURSOR *cursor, WT_ITEM *v) {
	return cursor->get_value(cursor, v);
}

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

int wiredtiger_cursor_search(WT_CURSOR *cursor, WT_ITEM *key) {
	cursor->set_key(cursor, key);
	return cursor->search(cursor);
}

int wiredtiger_cursor_search_near(WT_CURSOR *cursor, WT_ITEM *key, int *exactp) {
	cursor->set_key(cursor, key);
	return cursor->search_near(cursor, exactp);
}

int wiredtiger_cursor_insert(WT_CURSOR *cursor, WT_ITEM *key, WT_ITEM *value) {
	cursor->set_key(cursor, key);
	cursor->set_value(cursor, value);
	return cursor->insert(cursor);
}

int wiredtiger_cursor_update(WT_CURSOR *cursor, WT_ITEM *key, WT_ITEM *value) {
	cursor->set_key(cursor, key);
	cursor->set_value(cursor, value);
	return cursor->update(cursor);
}

int wiredtiger_cursor_remove(WT_CURSOR *cursor, WT_ITEM *key) {
	cursor->set_key(cursor, key);
	return cursor->remove(cursor);
}
*/
import "C"
import "unsafe"

type Cursor struct {
	w           *C.WT_CURSOR
	session     *Session
	uri         string
	keyFormat   string
	valueFormat string
	keyPack     []byte
	valuePack   []byte
}

// General

func (c *Cursor) Close() int {
	result := int(C.wiredtiger_cursor_close(c.w))

	if result == 0 {
		c = nil
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

func (c *Cursor) GetSession() *Session {
	return c.session
}

func (c *Cursor) GetUri() string {
	return c.uri
}

func (c *Cursor) GetKeyFormat() string {
	return c.keyFormat
}

func (c *Cursor) GetValueFormat() string {
	return c.valueFormat
}

// Data access
// TODO: implement

func (c *Cursor) GetKey(a ...interface{}) int {
	if c.keyPack == nil {
		var v C.WT_ITEM

		if result := int(C.wiredtiger_cursor_get_key(c.w, &v)); result != 0 {
			return result
		}

		c.keyPack = C.GoBytes(unsafe.Pointer(v.data), C.int(v.size))
	}

	return UnPack(c.keyFormat, c.keyPack, a...)
}

func (c *Cursor) GetValue(a ...interface{}) int {
	if c.valuePack == nil {
		var v C.WT_ITEM

		if result := int(C.wiredtiger_cursor_get_value(c.w, &v)); result != 0 {
			return result
		}

		c.valuePack = C.GoBytes(unsafe.Pointer(v.data), C.int(v.size))
	}

	return UnPack(c.valueFormat, c.valuePack, a...)
}

func (c *Cursor) SetKey(a ...interface{}) int {
	b, res := Pack(c.keyFormat, a...)

	if res == 0 {
		c.keyPack = b
	}

	return res
}

func (c *Cursor) SetValue(a ...interface{}) int {
	b, res := Pack(c.valueFormat, a...)

	if res == 0 {
		c.valuePack = b
	}

	return res
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
	res := int(C.wiredtiger_cursor_next(c.w))

	if res == 0 {
		c.keyPack = nil
		c.valuePack = nil
	}

	return res
}

func (c *Cursor) Prev() int {
	res := int(C.wiredtiger_cursor_prev(c.w))

	if res == 0 {
		c.keyPack = nil
		c.valuePack = nil
	}

	return res
}

func (c *Cursor) Reset() int {
	c.keyPack = nil
	c.valuePack = nil
	return int(C.wiredtiger_cursor_reset(c.w))
}

func (c *Cursor) Search() int {
	var key C.WT_ITEM

	if c.keyPack != nil && len(c.keyPack) > 0 {
		key.data = unsafe.Pointer(&c.keyPack[0])
		key.size = C.size_t(len(c.keyPack))
	}

	res := int(C.wiredtiger_cursor_search(c.w, &key))

	if res == 0 {
		c.valuePack = nil
	}

	return res
}

func (c *Cursor) SearchNear() (compare_result, result int) {
	var compare_resultC C.int
	var key C.WT_ITEM

	if c.keyPack != nil && len(c.keyPack) > 0 {
		key.data = unsafe.Pointer(&c.keyPack[0])
		key.size = C.size_t(len(c.keyPack))
	}

	result = int(C.wiredtiger_cursor_search_near(c.w, &key, &compare_resultC))

	if result == 0 {
		compare_result = int(compare_resultC)
		c.keyPack = nil
		c.valuePack = nil
	}

	return
}

// Data modification

func (c *Cursor) Insert() int {
	var key, val C.WT_ITEM

	if c.keyPack != nil && len(c.keyPack) > 0 {
		key.data = unsafe.Pointer(&c.keyPack[0])
		key.size = C.size_t(len(c.keyPack))
	}

	if c.valuePack != nil && len(c.valuePack) > 0 {
		val.data = unsafe.Pointer(&c.valuePack[0])
		val.size = C.size_t(len(c.valuePack))
	}

	return int(C.wiredtiger_cursor_insert(c.w, &key, &val))
}

func (c *Cursor) Update() int {
	var key, val C.WT_ITEM

	if c.keyPack != nil && len(c.keyPack) > 0 {
		key.data = unsafe.Pointer(&c.keyPack[0])
		key.size = C.size_t(len(c.keyPack))
	}

	if c.valuePack != nil && len(c.valuePack) > 0 {
		val.data = unsafe.Pointer(&c.valuePack[0])
		val.size = C.size_t(len(c.valuePack))
	}

	return int(C.wiredtiger_cursor_update(c.w, &key, &val))
}

func (c *Cursor) Remove() int {
	var key C.WT_ITEM

	if c.keyPack != nil && len(c.keyPack) > 0 {
		key.data = unsafe.Pointer(&c.keyPack[0])
		key.size = C.size_t(len(c.keyPack))
	}

	res := int(C.wiredtiger_cursor_remove(c.w, &key))

	if res == 0 {
		c.valuePack = nil
	}

	return res
}
