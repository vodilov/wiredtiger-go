package wiredtiger

/*
#cgo LDFLAGS: -lwiredtiger
#include <stdlib.h>
#include <wiredtiger.h>

char buf[4096];
size_t buf_size = 0;

WT_CONNECTION *conn = NULL;
WT_SESSION *session = NULL;


int packtest_init() {
	int ret;

	if (ret = wiredtiger_open(NULL, NULL, "create", &conn))
		return ret;

	if (ret = conn->open_session(conn, NULL, NULL, &session))
		return ret;
}

int packtest_deinit() {
	if (conn)
		return conn->close(conn, NULL);

	return 0;
}

int packtest_intpack(int64_t v) {
    int ret;

	if(ret = wiredtiger_struct_size(session, &buf_size, "q", v))
		return ret;

    return wiredtiger_struct_pack(session, buf, buf_size, "q", v);
}

int packtest_uintpack(uint64_t v) {
    int ret;

	if(ret = wiredtiger_struct_size(session, &buf_size, "Q", v))
		return ret;

    return wiredtiger_struct_pack(session, buf, buf_size, "Q", v);
}

int packtest_general() {
    int ret;
	char vb1[] = {1, 2, 3};
	char vb2[] = {4, 5, 6, 7};
	WT_ITEM wti_vb1, wti_vb2;

	wti_vb1.data = vb1;
	wti_vb1.size = 3;

	wti_vb2.data = vb2;
	wti_vb2.size = 4;

	if(ret = wiredtiger_struct_size(session, &buf_size, "xbBq3sSuu", -2, 2, -9223372036854775808ULL, "ABCD", "Hello", &wti_vb1, &wti_vb2))
		return ret;

    return wiredtiger_struct_pack(session, buf, buf_size, "xbBq3sSuu",  -2, 2, -9223372036854775808ULL, "ABCD", "Hello", &wti_vb1, &wti_vb2);
}


*/
import "C"
import "unicode"
import "strings"
import "bytes"

//import "fmt"

type wtpack struct {
	pfmt     *string
	curIdx   int
	repeats  int
	havesize bool
	size     int
	vtype    byte
}

func (p *wtpack) start(pfmt *string) int {
	if len(*pfmt) == 0 {
		*pfmt = "u"
	}

	if (*pfmt)[0] == '@' || (*pfmt)[0] == '<' || (*pfmt)[0] == '>' {
		return EINVAL
	}

	if (*pfmt)[0] == '.' {
		p.curIdx = 1
	}

	if p.curIdx == len(*pfmt) {
		return EINVAL
	}

	p.pfmt = pfmt

	return 0
}

func (p *wtpack) next() int {
	if p.repeats > 0 {
		p.repeats--
		return 0
	}

pfmt_next:

	if p.curIdx == len(*p.pfmt) {
		return WT_NOTFOUND
	}

	if unicode.IsDigit(rune((*p.pfmt)[p.curIdx])) {
		p.havesize = true
		p.size = 0

		for ; p.curIdx < len(*p.pfmt) && unicode.IsDigit(rune((*p.pfmt)[p.curIdx])); p.curIdx++ {
			p.size *= 10
			p.size += int((*p.pfmt)[p.curIdx] - '0')
		}

		if p.curIdx == len(*p.pfmt) {
			return EINVAL
		}
	} else {
		p.havesize = false
		p.size = 1
	}

	p.vtype = (*p.pfmt)[p.curIdx]

	switch p.vtype {
	case 'S', 'x':
	case 's':
		/* Fixed length strings must be at least 1 byte */
		if p.size < 1 {
			return EINVAL
		}
	case 't':
		/* Bitfield sizes must be between 1 and 8 bits */
		if p.size < 1 || p.size > 8 {
			return EINVAL
		}
	case 'u', 'U':
		/* Special case for items with a size prefix. */
		if (p.havesize == false) && (p.curIdx != len(*p.pfmt)-1) {
			p.vtype = 'U'
		} else {
			p.vtype = 'u'
		}
	case 'b', 'h', 'i', 'B', 'H', 'I', 'l', 'L', 'q', 'Q', 'r', 'R':
		if p.size == 0 {
			p.curIdx++
			goto pfmt_next
		}

		p.havesize = false
		p.repeats = p.size - 1
	default:
		return EINVAL
	}

	p.curIdx++
	return 0
}

func (p *wtpack) reset() {
	p.curIdx = 0
}

func (p *wtpack) pack_size(i interface{}) (int, int) {
	switch p.vtype {
	case 'x':
		return int(p.size), 0
	case 'S', 's':
		v, ok := i.(string)
		if ok == false {
			return 0, EINVAL
		}

		if p.vtype == 's' || p.havesize == true {
			return p.size, 0
		} else {
			s := strings.IndexByte(v, 0)
			if s != -1 {
				return s + 1, 0
			}

			return len(v) + 1, 0
		}
	case 'u', 'U':
		v, ok := i.([]byte)
		if ok == false {
			panic(0)
			return 0, EINVAL
		}

		s := len(v)
		pad := 0

		switch {
		case p.havesize == true && p.size < s:
			s = p.size
		case p.havesize == true:
			pad = p.size - s
		}

		if p.vtype == 'U' {
			s += vsize_uint(uint64(s + pad))
		}

		return s + pad, 0

	case 'b':
		if _, ok := i.(int8); ok == false {
			panic(0)
			return 0, EINVAL
		}

		return 1, 0
	case 'B', 't':
		if _, ok := i.(byte); ok == false {
			panic(0)
			return 0, EINVAL
		}

		return 1, 0

	case 'h', 'i', 'l', 'q':
		switch v := i.(type) {
		case int:
			return vsize_int(int64(v)), 0
		case int16:
			return vsize_int(int64(v)), 0
		case int32:
			return vsize_int(int64(v)), 0
		case int64:
			return vsize_int(v), 0
		default:
			panic(0)
			return 0, EINVAL
		}

	case 'H', 'I', 'L', 'Q', 'r':
		switch v := i.(type) {
		case uint:
			return vsize_uint(uint64(v)), 0
		case uint16:
			return vsize_uint(uint64(v)), 0
		case uint32:
			return vsize_uint(uint64(v)), 0
		case uint64:
			return vsize_uint(v), 0
		default:
			panic(0)
			return 0, EINVAL
		}

	default:
		panic(0)
		return 0, EINVAL
	}
}

func (p *wtpack) pack(buf []byte, i interface{}) ([]byte, int) {
	switch p.vtype {
	case 'x':
		for p.size > 0 {
			buf = append(buf, byte(0))
			p.size--
		}
	case 's':
		v, ok := i.(string)
		if !ok {
			return buf[:0], EINVAL
		}

		switch {
		case p.size == len(v):
			buf = append(buf, v...)
		case p.size > len(v):
			pad := p.size - len(v)
			buf = append(buf, v...)

			for ; pad != 0; pad-- {
				buf = append(buf, byte(0))
			}
		case p.size < len(v):
			buf = append(buf, v[:p.size]...)
		}
	case 'S':
		v, ok := i.(string)
		if !ok {
			return buf[:0], EINVAL
		}

		s := strings.IndexByte(v, 0)
		if s == -1 {
			buf = append(buf, v...)
			buf = append(buf, byte(0))
		} else {
			buf = append(buf, v[:s+1]...)
		}
	case 'u', 'U':
		v, ok := i.([]byte)
		if !ok {
			return buf[:0], EINVAL
		}

		s := len(v)
		pad := 0

		switch {
		case p.havesize == true && p.size < s:
			s = p.size
		case p.havesize == true:
			pad = p.size - s
		}

		if p.vtype == 'U' {
			buf = vpack_uint(buf, uint64(s+pad))
		}

		if s > 0 {
			buf = append(buf, v[:s]...)
		}

		for ; pad != 0; pad-- {
			buf = append(buf, byte(0))
		}
	case 'b':
		v, ok := i.(int8)
		if !ok {
			return buf[:0], EINVAL
		}

		buf = append(buf, byte(uint8(v)^0x80))
	case 'B', 't':
		v, ok := i.(byte)
		if !ok {
			return buf[:0], EINVAL
		}

		buf = append(buf, v)
	case 'h', 'i', 'l', 'q', 'H', 'I', 'L', 'Q', 'r':
		switch v := i.(type) {
		case int:
			buf = vpack_int(buf, int64(v))
		case int16:
			buf = vpack_int(buf, int64(v))
		case int32:
			buf = vpack_int(buf, int64(v))
		case int64:
			buf = vpack_int(buf, v)
		case uint:
			buf = vpack_uint(buf, uint64(v))
		case uint16:
			buf = vpack_uint(buf, uint64(v))
		case uint32:
			buf = vpack_uint(buf, uint64(v))
		case uint64:
			buf = vpack_uint(buf, v)
		default:
			return buf[:0], EINVAL
		}
	}

	return buf, 0
}

func (p *wtpack) unpack(buf []byte, bcur *int, bend int, i interface{}) int {
	switch p.vtype {
	case 'x':
		*bcur += p.size
	case 'S', 's':
		var s int
		v, ok := i.(*string)
		if ok == false {
			return EINVAL
		}

		if p.vtype == 's' || p.havesize == true {
			s = p.size
			*v = string(buf[*bcur : *bcur+s])
			*bcur += s
		} else {
			s = bytes.IndexByte(buf[*bcur:], 0)
			switch {
			case s == 0:
				*v = ""
				*bcur++
			case s > 0:
				*v = string(buf[*bcur : *bcur+s])
				*bcur += s + 1
			default:
				return EINVAL
			}
		}
	case 'u', 'U':
		var s int
		v, ok := i.(*[]byte)
		if ok == false {
			return EINVAL
		}

		switch {
		case p.havesize == true:
			s = p.size
		case p.vtype == 'U':
			if su, r := vunpack_uint(buf, bcur, bend); r != 0 {
				return r
			} else {
				s = int(su)
			}

		default:
			s = bend - *bcur
		}

		*v = buf[*bcur : *bcur+s]
		*bcur += s

	case 'b':
		v, ok := i.(*int8)
		if ok == false {
			return EINVAL
		}

		*v = int8(buf[*bcur] ^ 0x80)
		*bcur++

	case 'B', 't':
		v, ok := i.(*uint8)
		if ok == false {
			return EINVAL
		}

		*v = buf[*bcur]
		*bcur++

	case 'h', 'i', 'l', 'q':
		if vc, r := vunpack_int(buf, bcur, bend); r != 0 {
			return r
		} else {
			switch v := i.(type) {
			case *int:
				*v = int(vc)
			case *int16:
				*v = int16(vc)
			case *int32:
				*v = int32(vc)
			case *int64:
				*v = int64(vc)
			case *uint:
				*v = uint(vc)
			case *uint16:
				*v = uint16(vc)
			case *uint32:
				*v = uint32(vc)
			case *uint64:
				*v = uint64(vc)
			default:
				return EINVAL
			}
		}
	case 'H', 'I', 'L', 'Q', 'r':
		if vc, r := vunpack_uint(buf, bcur, bend); r != 0 {
			return r
		} else {
			switch v := i.(type) {
			case *int:
				*v = int(vc)
			case *int16:
				*v = int16(vc)
			case *int32:
				*v = int32(vc)
			case *int64:
				*v = int64(vc)
			case *uint:
				*v = uint(vc)
			case *uint16:
				*v = uint16(vc)
			case *uint32:
				*v = uint32(vc)
			case *uint64:
				*v = uint64(vc)
			default:
				return EINVAL
			}
		}
	default:
		return EINVAL
	}

	return 0
}

func Pack(session *Session, pfmt string, buf []byte, a ...interface{}) ([]byte, error) {
	var res int
	var cidx int

	buf = buf[:0]

	pcnt := len(a)

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return nil, NewError(res, session)
	}

	// Parameters have allready validated
	res = wtp.next()
	for res == 0 {
		if wtp.vtype == 'x' {
			buf, res = wtp.pack(buf, byte(0))
			res = wtp.next()
			continue
		}

		if cidx == pcnt {
			res = EINVAL
			buf = buf[:0]
			break
		}

		if buf, res = wtp.pack(buf, a[cidx]); res != 0 {
			break
		}

		res = wtp.next()
		cidx++
	}

	if res != 0 && res != WT_NOTFOUND {
		return buf, NewError(res, session)
	}

	return buf, nil
}

func UnPack(session *Session, pfmt string, buf []byte, a ...interface{}) error {
	var res int
	var cidx int
	var bcur int

	pcnt := len(a)
	bend := len(buf)

	if bend == 0 {
		return NewError(EINVAL, session)
	}

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return NewError(EINVAL, session)
	}

	res = wtp.next()
	for res == 0 {
		if wtp.vtype == 'x' {
			res = wtp.unpack(buf, &bcur, bend, byte(0))
			res = wtp.next()
			continue
		}

		if pcnt == 0 {
			res = EINVAL
			break
		}

		if res = wtp.unpack(buf, &bcur, bend, a[cidx]); res == 0 {
			res = wtp.next()
			cidx++
			pcnt--
		}
	}

	if res != 0 && res != WT_NOTFOUND {
		return NewError(res, session)
	}

	return nil
}

func initPackTest() int {
	return int(C.packtest_init())
}

func deinitPackTest() int {
	return int(C.packtest_deinit())
}

func getResultPackTest() []byte {
	b := make([]byte, 0, int(C.buf_size))

	for i := 0; i < int(C.buf_size); i++ {
		b = append(b, byte(C.buf[i]))
	}

	return b
}

func intPackTest(v int64) []byte {
	C.packtest_intpack(C.int64_t(v))

	return getResultPackTest()
}

func uintPackTest(v uint64) []byte {
	C.packtest_uintpack(C.uint64_t(v))

	return getResultPackTest()
}

func generalPackTest() []byte {
	C.packtest_general()

	return getResultPackTest()
}

func PackInterface(a ...interface{}) []byte {
	var buf []byte
	if a == nil || len(a) == 0 {
		return nil
	}

	lastArg := len(a) - 1

	for i, arg := range a {
		switch v := arg.(type) {
		case string:
			s := strings.IndexByte(v, 0)
			if s != -1 {
				buf = append(buf, v[:s+1]...)
				buf = append(buf, byte(0))
			} else {
				buf = append(buf, v...)
			}
		case []byte:
			if i != lastArg {
				buf = vpack_uint(buf, uint64(len(v)))
			}
			buf = append(buf, v...)
		case int8:
			buf = append(buf, byte(uint8(v)^0x80))
		case byte:
			buf = append(buf, v)
		case int:
			buf = vpack_int(buf, int64(v))
		case int16:
			buf = vpack_int(buf, int64(v))
		case int32:
			buf = vpack_int(buf, int64(v))
		case int64:
			buf = vpack_int(buf, v)
		case uint:
			buf = vpack_uint(buf, uint64(v))
		case uint16:
			buf = vpack_uint(buf, uint64(v))
		case uint32:
			buf = vpack_uint(buf, uint64(v))
		case uint64:
			buf = vpack_uint(buf, v)
		default:
			return nil
		}
	}

	return buf
}

func UnPackInterface(session *Session, buf []byte, a ...interface{}) error {
	var bcur int

	if len(buf) == 0 || a == nil || len(a) == 0 {
		return nil
	}

	lastArg := len(a) - 1
	bend := len(buf)

	for i, arg := range a {
		switch v := arg.(type) {
		case *string:
			s := bytes.IndexByte(buf[bcur:], 0)
			switch {
			case s == 0:
				*v = ""
				bcur++
			case s > 0:
				*v = string(buf[bcur : bcur+s])
				bcur += s + 1
			default:
				return NewError(EINVAL, session)
			}
		case *[]byte:
			var s int

			if i != lastArg {
				if su, r := vunpack_uint(buf, &bcur, bend); r != 0 {
					return NewError(r, session)
				} else {
					s = int(su)
				}
			} else {
				s = bend - bcur
			}

			*v = buf[bcur : bcur+s]
			bcur += s
		case *int8:
			*v = int8(buf[bcur] ^ 0x80)
			bcur++
		case *byte:
			*v = buf[bcur]
			bcur++
		case *int:
			if vc, r := vunpack_int(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = int(vc)
			}
		case *int16:
			if vc, r := vunpack_int(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = int16(vc)
			}
		case *int32:
			if vc, r := vunpack_int(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = int32(vc)
			}
		case *int64:
			if vc, r := vunpack_int(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = vc
			}
		case *uint:
			if vc, r := vunpack_uint(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = uint(vc)
			}
		case *uint16:
			if vc, r := vunpack_uint(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = uint16(vc)
			}
		case *uint32:
			if vc, r := vunpack_uint(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = uint32(vc)
			}
		case *uint64:
			if vc, r := vunpack_uint(buf, &bcur, bend); r != 0 {
				return NewError(r, session)
			} else {
				*v = vc
			}
		default:
			return NewError(EINVAL, session)
		}
	}

	return nil
}
