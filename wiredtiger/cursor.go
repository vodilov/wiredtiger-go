package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include <wiredtiger.h>

#define WT_SIZE_ZERO (size_t)((size_t)(SIZE_MAX) >> 1)

typedef const void* CPVOID;

int wiredtiger_cursor_close(WT_CURSOR *cursor) {
	return cursor->close(cursor);
}

int wiredtiger_cursor_reconfigure(WT_CURSOR *cursor, const char *config) {
	int ret;

	if(ret = cursor->reconfigure(cursor, config))
		return ret;

	if ((cursor->flags & WT_CURSTD_DUMP_JSON) == 0)
			cursor->flags |= WT_CURSTD_RAW;

	return 0;
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

int wiredtiger_cursor_search(WT_CURSOR *cursor, const void *data, size_t size) {
	if (size != 0) {
		WT_ITEM key;
		key.data = data;
		key.size = size == WT_SIZE_ZERO ? 0 : size;
		cursor->set_key(cursor, &key);
	}

	return cursor->search(cursor);
}

int wiredtiger_cursor_search_near(WT_CURSOR *cursor, const void *data, size_t size, int *exactp) {
	if (size != 0) {
		WT_ITEM key;
		key.data = data;
		key.size = size == WT_SIZE_ZERO ? 0 : size;
		cursor->set_key(cursor, &key);
	}

	return cursor->search_near(cursor, exactp);
}

int wiredtiger_cursor_insert(WT_CURSOR *cursor, const void *key_data, size_t key_size, const void *val_data, size_t val_size) {

	if (key_size != 0) {
		WT_ITEM key;
		key.data = key_data;
		key.size = key_size == WT_SIZE_ZERO ? 0 : key_size;
		cursor->set_key(cursor, &key);
	}

	if (val_size != 0) {
		WT_ITEM value;
		value.data = val_data;
		value.size = val_size == WT_SIZE_ZERO ? 0 : val_size;
		cursor->set_value(cursor, &value);
	}

	return cursor->insert(cursor);
}

int wiredtiger_cursor_update(WT_CURSOR *cursor, const void *key_data, size_t key_size, const void *val_data, size_t val_size) {

	if (key_size != 0) {
		WT_ITEM key;
		key.data = key_data;
		key.size = key_size == WT_SIZE_ZERO ? 0 : key_size;
		cursor->set_key(cursor, &key);
	}

	if (val_size != 0) {
		WT_ITEM value;
		value.data = val_data;
		value.size = val_size == WT_SIZE_ZERO ? 0 : val_size;
		cursor->set_value(cursor, &value);
	}

	return cursor->update(cursor);
}

int wiredtiger_cursor_remove(WT_CURSOR *cursor, const void *data, size_t size) {
	if (size != 0) {
		WT_ITEM key;
		key.data = data;
		key.size = size == WT_SIZE_ZERO ? 0 : size;
		cursor->set_key(cursor, &key);
	}

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
	keySetExt   bool
	valuePack   []byte
	valueSetExt bool
}

// General

func (c *Cursor) Close() error {
	if res := int(C.wiredtiger_cursor_close(c.w)); res != 0 {
		return NewError(res, c.session)
	}

	return nil
}

func (c *Cursor) Reconfigure(config string) error {
	var configC *C.char = nil

	if len(config) > 0 {
		configC = C.CString(config)
		defer C.free(unsafe.Pointer(configC))
	}

	if res := int(C.wiredtiger_cursor_reconfigure(c.w, configC)); res != 0 {
		return NewError(res, c.session)
	}

	return nil
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

func (c *Cursor) GetKey(a ...interface{}) error {
	var v C.WT_ITEM

	if res := int(C.wiredtiger_cursor_get_key(c.w, &v)); res != 0 {
		return NewError(res, c.session)
	}

	return UnPack(c.session, c.keyFormat, C.GoBytes(unsafe.Pointer(v.data), C.int(v.size)), a...)
}

func (c *Cursor) GetValue(a ...interface{}) error {
	var v C.WT_ITEM

	if res := int(C.wiredtiger_cursor_get_value(c.w, &v)); res != 0 {
		return NewError(res, c.session)
	}

	return UnPack(c.session, c.valueFormat, C.GoBytes(unsafe.Pointer(v.data), C.int(v.size)), a...)
}

func (c *Cursor) SetKey(a ...interface{}) error {
	var res error
	c.keyPack, res = Pack(c.session, c.keyFormat, c.keyPack, a...)

	if res != nil {
		c.keySetExt = false
		return res
	}

	c.keySetExt = true
	return nil
}

func (c *Cursor) SetValue(a ...interface{}) error {
	var res error

	c.valuePack, res = Pack(c.session, c.valueFormat, c.valuePack, a...)

	if res != nil {
		c.valueSetExt = false
		return res
	}

	c.valueSetExt = true
	return nil
}

// Cursor positioning

func (c *Cursor) Compare(other *Cursor) (int, error) {
	var oc *C.WT_CURSOR
	var compare_resultC C.int

	if other != nil {
		oc = other.w
	}

	if res := int(C.wiredtiger_cursor_compare(c.w, oc, &compare_resultC)); res != 0 {
		return 0, NewError(res, c.session)
	}

	return int(compare_resultC), nil
}

func (c *Cursor) Equals(other *Cursor) (bool, error) {
	var oc *C.WT_CURSOR
	var compare_resultC C.int

	if other != nil {
		oc = other.w
	}

	if res := int(C.wiredtiger_cursor_equals(c.w, oc, &compare_resultC)); res != 0 {
		return false, NewError(res, c.session)
	}

	if compare_resultC == 0 {
		return false, nil
	}

	return true, nil
}

func (c *Cursor) Next() error {
	if res := int(C.wiredtiger_cursor_next(c.w)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}
	return nil
}

func (c *Cursor) Prev() error {
	if res := int(C.wiredtiger_cursor_prev(c.w)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}
	return nil
}

func (c *Cursor) Reset() error {
	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}

	if res := int(C.wiredtiger_cursor_reset(c.w)); res != 0 {
		return NewError(res, c.session)
	}

	return nil
}

func (c *Cursor) Search() error {
	var key_data unsafe.Pointer
	var key_size C.size_t

	if len(c.keyPack) > 0 {
		key_data = unsafe.Pointer(&c.keyPack[0])
		key_size = C.size_t(len(c.keyPack))
	} else if c.keySetExt {
		key_size = C.WT_SIZE_ZERO
	}

	if res := int(C.wiredtiger_cursor_search(c.w, key_data, key_size)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}
	return nil
}

func (c *Cursor) SearchNear() (int, error) {
	var key_data unsafe.Pointer
	var key_size C.size_t
	var compare_resultC C.int

	if len(c.keyPack) > 0 {
		key_data = unsafe.Pointer(&c.keyPack[0])
		key_size = C.size_t(len(c.keyPack))
	} else if c.keySetExt {
		key_size = C.WT_SIZE_ZERO
	}

	if res := int(C.wiredtiger_cursor_search_near(c.w, key_data, key_size, &compare_resultC)); res != 0 {
		return 0, NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}

	return int(compare_resultC), nil
}

// Data modification

func (c *Cursor) Insert() error {
	var key_data, value_data unsafe.Pointer
	var key_size, value_size C.size_t

	if len(c.keyPack) > 0 {
		key_data = unsafe.Pointer(&c.keyPack[0])
		key_size = C.size_t(len(c.keyPack))
	} else if c.keySetExt {
		key_size = C.WT_SIZE_ZERO
	}

	if len(c.valuePack) > 0 {
		value_data = unsafe.Pointer(&c.valuePack[0])
		value_size = C.size_t(len(c.valuePack))
	} else if c.valueSetExt {
		value_size = C.WT_SIZE_ZERO
	}

	if res := int(C.wiredtiger_cursor_insert(c.w, key_data, key_size, value_data, value_size)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}

	return nil
}

func (c *Cursor) Update() error {
	var key_data, value_data unsafe.Pointer
	var key_size, value_size C.size_t

	if len(c.keyPack) > 0 {
		key_data = unsafe.Pointer(&c.keyPack[0])
		key_size = C.size_t(len(c.keyPack))
	} else if c.keySetExt {
		key_size = C.WT_SIZE_ZERO
	}

	if len(c.valuePack) > 0 {
		value_data = unsafe.Pointer(&c.valuePack[0])
		value_size = C.size_t(len(c.valuePack))
	} else if c.valueSetExt {
		value_size = C.WT_SIZE_ZERO
	}

	if res := int(C.wiredtiger_cursor_update(c.w, key_data, key_size, value_data, value_size)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}

	return nil
}

func (c *Cursor) Remove() error {
	var key_data unsafe.Pointer
	var key_size C.size_t

	if len(c.keyPack) > 0 {
		key_data = unsafe.Pointer(&c.keyPack[0])
		key_size = C.size_t(len(c.keyPack))
	} else if c.keySetExt {
		key_size = C.WT_SIZE_ZERO
	}

	if res := int(C.wiredtiger_cursor_remove(c.w, key_data, key_size)); res != 0 {
		return NewError(res, c.session)
	}

	if c.keySetExt {
		c.keyPack = c.keyPack[:0]
		c.keySetExt = false
	}

	if c.valueSetExt {
		c.valuePack = c.valuePack[:0]
		c.valueSetExt = false
	}

	return nil
}
