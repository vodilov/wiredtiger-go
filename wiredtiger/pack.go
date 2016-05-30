package wiredtiger

import "syscall"
import "unicode"
import "strings"

type wtpack struct {
	pfmt     *string
	curIdx   int
	endIdx   int
	curIIdx  int
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
		_, ok := i.(byte)
		if ok == false {
			return 0, int(syscall.EINVAL)
		}

		return 1, 0

	case 'H', 'I', 'L', 'Q', 'r':
		_, ok := i.(byte)
		if ok == false {
			return 0, int(syscall.EINVAL)
		}

		return 1, 0

	case 'R':
		_, ok := i.(uint64)
		if ok == false {
			return 0, int(syscall.EINVAL)
		}

		return 8, 0

	default:
		return 0, int(syscall.EINVAL)
	}
}

func Pack(pfmt string, a ...interface{}) ([]byte, int) {
	var r []byte
	var res int

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return nil, int(syscall.EINVAL)
	}

	for {
		if res = wtp.next(); res != 0 {
			break
		}

	}

	if res != 0 && res != WT_NOTFOUND {
		return nil, res
	}

	return r, 0
}
