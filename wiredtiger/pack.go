package wiredtiger

import "syscall"
import "unicode"
import "strings"
import "bytes"

type wtpack struct {
	pfmt     *string
	curIdx   int
	endIdx   int
	repeats  int
	havesize bool
	size     int
	vtype    byte
}

func (p *wtpack) start(pfmt *string) int {
	if len(*pfmt) == 0 {
		*pfmt = "u"
	}

	p.endIdx = len(*pfmt) - 1

	if (*pfmt)[0] == '@' || (*pfmt)[0] == '<' || (*pfmt)[0] == '>' {
		return int(syscall.EINVAL)
	}

	if (*pfmt)[0] == '.' {
		p.curIdx = 1
	}

	if p.curIdx == p.endIdx {
		return int(syscall.EINVAL)
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

	if p.curIdx > p.endIdx {
		return WT_NOTFOUND
	}

	if unicode.IsDigit(rune((*p.pfmt)[p.curIdx])) {
		p.havesize = true
		p.size = 0

		for ; unicode.IsDigit(rune((*p.pfmt)[p.curIdx])) && p.curIdx < p.endIdx; p.curIdx++ {
			p.size *= 10
			p.size += int((*p.pfmt)[p.curIdx] - '0')
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
			return int(syscall.EINVAL)
		}
	case 't':
		/* Bitfield sizes must be between 1 and 8 bits */
		if p.size < 1 || p.size > 8 {
			return int(syscall.EINVAL)
		}
	case 'u', 'U':
		/* Special case for items with a size prefix. */
		if (p.havesize == false) && (p.curIdx != p.endIdx) {
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
		return int(syscall.EINVAL)
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
			return 0, int(syscall.EINVAL)
		}

		if p.vtype == 's' || p.havesize == true {
			return p.size, 0
		} else {
			s := strings.IndexByte(v, 0)
			if s != -1 {
				return s + 1, 0
			}

			return len(v), 0
		}
	case 'u', 'U':
		v, ok := i.([]byte)
		if ok == false {
			return 0, int(syscall.EINVAL)
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

	case 'b', 'B', 't':
		_, ok := i.(byte)
		if ok == false {
			return 0, int(syscall.EINVAL)
		}

		return 1, 0

	case 'h', 'i', 'l', 'q':
		var vc int64

		switch v := i.(type) {
		case int16:
			vc = int64(v)
		case int32:
			vc = int64(v)
		case int64:
			vc = v
		default:
			return 0, int(syscall.EINVAL)
		}

		return vsize_int(vc), 0

	case 'H', 'I', 'L', 'Q', 'r':
		var vc uint64

		switch v := i.(type) {
		case uint16:
			vc = uint64(v)
		case uint32:
			vc = uint64(v)
		case uint64:
			vc = v
		default:
			return 0, int(syscall.EINVAL)
		}

		return vsize_uint(vc), 0

	default:
		return 0, int(syscall.EINVAL)
	}
}

func (p *wtpack) pack(buf *[]byte, i interface{}) {
	switch p.vtype {
	case 'x':
		for p.size > 0 {
			*buf = append(*buf, byte(0))
			p.size--
		}
	case 's':
		v := i.(string)
		switch {
		case p.size == len(v):
			*buf = append(*buf, v...)
		case p.size > len(v):
			pad := p.size - len(v)
			*buf = append(*buf, v...)

			for ; pad != 0; pad-- {
				*buf = append(*buf, byte(0))
			}
		case p.size < len(v):
			*buf = append(*buf, v[:p.size]...)
		}
	case 'S':
		v := i.(string)
		s := strings.IndexByte(v, 0)
		if s == -1 {
			*buf = append(*buf, v...)
			*buf = append(*buf, byte(0))
		} else {
			*buf = append(*buf, v[:s+1]...)
		}
	case 'u', 'U':
		v := i.([]byte)
		s := len(v)
		pad := 0

		switch {
		case p.havesize == true && p.size < s:
			s = p.size
		case p.havesize == true:
			pad = p.size - s
		}

		if p.vtype == 'U' {
			vpack_uint(buf, uint64(s+pad))
		}

		if s > 0 {
			*buf = append(*buf, v[:s]...)
		}

		for ; pad != 0; pad-- {
			*buf = append(*buf, byte(0))
		}
	case 'b':
		v := i.(int8)
		*buf = append(*buf, byte(uint8(v)^0x80))
	case 'B', 't':
		v := i.(byte)
		*buf = append(*buf, v)
	case 'h', 'i', 'l', 'q':
		var vc int64

		switch v := i.(type) {
		case int16:
			vc = int64(v)
		case int32:
			vc = int64(v)
		case int64:
			vc = v
		}

		vpack_int(buf, vc)
	case 'H', 'I', 'L', 'Q', 'r':
		var vc uint64

		switch v := i.(type) {
		case uint16:
			vc = uint64(v)
		case uint32:
			vc = uint64(v)
		case uint64:
			vc = v
		}

		vpack_uint(buf, vc)
	}
}

func (p *wtpack) unpack(buf []byte, bcur *int, bend int, i interface{}) int {
	switch p.vtype {
	case 'x':
		*bcur += p.size
	case 'S', 's':
		var s int
		v, ok := i.(*string)
		if ok == false {
			return int(syscall.EINVAL)
		}

		if p.vtype == 's' || p.havesize == true {
			s = p.size
		} else {
			s = bytes.IndexByte(buf[*bcur:], 0)
			if s != -1 {
				return int(syscall.EINVAL)
			}
		}

		*v = string(buf[*bcur : *bcur+s])
		*bcur += s
	case 'u', 'U':
		var s int
		v, ok := i.(*[]byte)
		if ok == false {
			return int(syscall.EINVAL)
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
			return int(syscall.EINVAL)
		}

		*v = int8(buf[*bcur] ^ 0x80)
		*bcur++

	case 'B', 't':
		v, ok := i.(*uint8)
		if ok == false {
			return int(syscall.EINVAL)
		}

		*v = buf[*bcur]
		*bcur++

	case 'h', 'i', 'l', 'q':
		if vc, r := vunpack_int(buf, bcur, bend); r != 0 {
			return r
		} else {
			switch v := i.(type) {
			case *int16:
				*v = int16(vc)
			case *int32:
				*v = int32(vc)
			case *int64:
				*v = int64(vc)
			default:
				return int(syscall.EINVAL)
			}
		}
	case 'H', 'I', 'L', 'Q', 'r':
		if vc, r := vunpack_uint(buf, bcur, bend); r != 0 {
			return r
		} else {
			switch v := i.(type) {
			case *uint16:
				*v = uint16(vc)
			case *uint32:
				*v = uint32(vc)
			case *uint64:
				*v = uint64(vc)
			default:
				return int(syscall.EINVAL)
			}
		}
	default:
		return int(syscall.EINVAL)
	}

	return 0
}

func Pack(pfmt string, a ...interface{}) ([]byte, int) {
	var res int
	var total int
	var cidx int

	pcnt := len(a)

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return nil, int(syscall.EINVAL)
	}

	res = wtp.next()
	for res == 0 {
		if wtp.vtype == 'x' {
			s, _ := wtp.pack_size(byte(0))
			total += s
			res = wtp.next()
			continue
		}

		if cidx == pcnt {
			res = int(syscall.EINVAL)
			break
		}

		s, r := wtp.pack_size(a[cidx])

		if r != 0 {
			res = r
			break
		}

		total += s
		res = wtp.next()
		cidx++
	}

	if res != 0 && res != WT_NOTFOUND {
		return nil, res
	}

	if total == 0 {
		return nil, int(syscall.EINVAL)
	}

	rarray := make([]byte, 0, total)
	wtp.reset()
	cidx = 0

	// Parameters have allready validated
	res = wtp.next()
	for res == 0 {
		if wtp.vtype == 'x' {
			res = wtp.next()
			continue
		}

		res = wtp.next()
		cidx++
	}

	if res != 0 && res != WT_NOTFOUND {
		return nil, res
	}

	return rarray, 0
}

func UnPack(pfmt string, buf []byte, a ...interface{}) int {
	var res int
	var cidx int
	var bcur int

	pcnt := len(a)
	bend := len(buf)

	if bend == 0 {
		return int(syscall.EINVAL)
	}

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return int(syscall.EINVAL)
	}

	res = wtp.next()
	for res == 0 {
		if wtp.vtype == 'x' {
			res = wtp.unpack(buf, &bcur, bend, byte(0))
			res = wtp.next()
			continue
		}

		if pcnt == 0 {
			res = int(syscall.EINVAL)
			break
		}

		if res = wtp.unpack(buf, &bcur, bend, a[cidx]); res == 0 {
			res = wtp.next()
			cidx++
			pcnt--
		}
	}

	if res != 0 && res != WT_NOTFOUND {
		return res
	}

	return 0
}
