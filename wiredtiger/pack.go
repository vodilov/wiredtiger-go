package wiredtiger

import "syscall"
import "unicode"

type wtpack struct {
	pfmt     *string
	curIdx   int
	endIdx   int
	repeats  uint32
	havesize bool
	size     uint32
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
			p.size += uint32((*p.pfmt)[p.curIdx] - '0')
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

func Pack(pfmt string, a ...interface{}) ([]byte, int) {
	var r []byte
	var res int

	wtp := new(wtpack)
	if res = wtp.start(&pfmt); res != 0 {
		return nil, int(syscall.EINVAL)
	}

	return r, 0
}
