package wiredtiger

import "syscall"

const (
	iNEG_MULTI_MARKER byte = 0x10
	iNEG_2BYTE_MARKER byte = 0x20
	iNEG_1BYTE_MARKER byte = 0x40
	iPOS_1BYTE_MARKER byte = 0x80
	iPOS_2BYTE_MARKER byte = 0xc0
	iPOS_MULTI_MARKER byte = 0xe0

	iNEG_1BYTE_MIN int64 = (-(1 << 6))
	iNEG_2BYTE_MIN int64 = (-(1 << 13) + iNEG_1BYTE_MIN)
	iPOS_1BYTE_MAX int64 = ((1 << 6) - 1)
	iPOS_2BYTE_MAX int64 = ((1 << 13) + iPOS_1BYTE_MAX)
)

func leading_zeros(x uint64) byte {
	var i byte
	m := uint64(0xFF << 56)

	for i = 0; (x&m) == 0 && i != 8; i++ {
		m >>= 8
	}

	return i
}

func get_bits(x uint64, start uint, end uint) uint64 {
	return (x & ((uint64(1) << (start)) - uint64(1))) >> (end)
}

func vpack_posint(buf *[]byte, x uint64) {
	lz := leading_zeros(x)
	l := 8 - lz

	*buf = append(*buf, byte(iPOS_MULTI_MARKER|(l&0xF)))

	shift := (l - 1) << 3

	for l != 0 {
		*buf = append(*buf, byte((x>>shift)&0xFF))
		l--
		shift -= 8
	}
}

func vpack_negint(buf *[]byte, x uint64) {
	lz := leading_zeros(^x)
	l := 8 - lz

	*buf = append(*buf, byte(iNEG_MULTI_MARKER|(lz&0xF)))

	shift := (l - 1) << 3

	for l != 0 {
		*buf = append(*buf, byte((x>>shift)&0xFF))
		l--
		shift -= 8
	}
}

func vunpack_posint(buf []byte, bcur *int, bend int) (uint64, int) {
	var x uint64

	l := int(buf[*bcur] & 0xf)
	if *bcur+l+1 > bend {
		return 0, int(syscall.EINVAL)
	}

	*bcur++

	for ; l != 0; l-- {
		x = (x << 8) | uint64(buf[*bcur])
		*bcur++
	}

	return x, 0
}

func vunpack_negint(buf []byte, bcur *int, bend int) (uint64, int) {
	var x uint64

	l := int(8 - buf[*bcur]&0xf)
	if *bcur+l+1 > bend {
		return 0, int(syscall.EINVAL)
	}

	*bcur++

	for x = ^uint64(0); l != 0; l-- {
		x = (x << 8) | uint64(buf[*bcur])
		*bcur++
	}

	return x, 0
}

func vpack_uint(buf *[]byte, x uint64) {
	switch {
	case x <= uint64(iPOS_1BYTE_MAX):
		*buf = append(*buf, byte(iPOS_1BYTE_MARKER|byte(get_bits(x, 6, 0)&0xFF)))
	case x <= uint64(iPOS_2BYTE_MAX):
		x -= uint64(iPOS_1BYTE_MAX) + 1
		*buf = append(*buf, byte(iPOS_2BYTE_MARKER|byte(get_bits(x, 13, 8)&0xFF)), byte(get_bits(x, 8, 0)&0xFF))
	case x == uint64(iPOS_2BYTE_MAX+1):
		*buf = append(*buf, byte(iPOS_1BYTE_MARKER|0x01), byte(0))
	default:
		vpack_posint(buf, x)
	}
}

func vpack_int(buf *[]byte, x int64) {
	switch {
	case x < int64(iNEG_2BYTE_MIN):
		vpack_negint(buf, uint64(x))
	case x < int64(iNEG_1BYTE_MIN):
		x -= iNEG_2BYTE_MIN
		*buf = append(*buf, byte(iNEG_2BYTE_MARKER|byte(get_bits(uint64(x), 13, 8)&0xFF)), byte(get_bits(uint64(x), 8, 0)&0xFF))
	case x < 0:
		x -= iNEG_1BYTE_MIN
		*buf = append(*buf, byte(iNEG_1BYTE_MARKER|byte(get_bits(uint64(x), 6, 0)&0xFF)))
	default:
		vpack_uint(buf, uint64(x))
	}
}

func vunpack_uint(buf []byte, bcur *int, bend int) (uint64, int) {
	switch buf[*bcur] & 0xF0 {
	case iPOS_1BYTE_MARKER, iPOS_1BYTE_MARKER | 0x10, iPOS_1BYTE_MARKER | 0x20, iPOS_1BYTE_MARKER | 0x30:
		x := get_bits(uint64(buf[*bcur]), 6, 0)
		*bcur++
		return x, 0
	case iPOS_2BYTE_MARKER, iPOS_2BYTE_MARKER | 0x10:
		if *bcur < bend {
			x := get_bits(uint64(buf[*bcur]), 5, 0) << 8
			*bcur++
			x |= uint64(buf[*bcur])
			x += uint64(iPOS_1BYTE_MAX + 1)
			*bcur++
			return x, 0
		}
	case iPOS_MULTI_MARKER:
		x, r := vunpack_posint(buf, bcur, bend)
		if r == 0 {
			x += uint64(iPOS_2BYTE_MAX + 1)
			return x, 0
		}

		return 0, r
	}

	return 0, int(syscall.EINVAL)
}

func vunpack_int(buf []byte, bcur *int, bend int) (int64, int) {
	switch buf[*bcur] & 0xF0 {
	case iNEG_MULTI_MARKER:
		x, r := vunpack_negint(buf, bcur, bend)

		return int64(x), r

	case iNEG_2BYTE_MARKER, iNEG_2BYTE_MARKER | 0x10:
		if *bcur < bend {
			x := int64(get_bits(uint64(buf[*bcur]), 5, 0) << 8)
			*bcur++
			x |= int64(buf[*bcur])
			x += int64(iPOS_1BYTE_MAX + 1)
			*bcur++
			return x, 0
		}

		return 0, int(syscall.EINVAL)

	case iNEG_1BYTE_MARKER, iNEG_1BYTE_MARKER | 0x10, iNEG_1BYTE_MARKER | 0x20, iNEG_1BYTE_MARKER | 0x30:
		x := int64(get_bits(uint64(buf[*bcur]), 6, 0)) + int64(iNEG_1BYTE_MIN)
		*bcur++
		return x, 0
	}

	x, r := vunpack_uint(buf, bcur, bend)
	return int64(x), r
}

func vsize_posint(x uint64) int {
	return 9 - int(leading_zeros(x))
}

func vsize_negint(x uint64) int {
	return 9 - int(leading_zeros(^x))
}

func vsize_uint(x uint64) int {
	switch {
	case x <= uint64(iPOS_1BYTE_MAX):
		return 1
	case x <= uint64(iPOS_2BYTE_MAX+1):
		return 2
	}

	x -= uint64(iPOS_2BYTE_MAX + 1)
	return vsize_posint(x)
}

func vsize_int(x int64) int {
	switch {
	case x < iNEG_2BYTE_MIN:
		return vsize_negint(uint64(x))
	case x < iNEG_1BYTE_MIN:
		return 2
	case x < 0:
		return 1
	}

	return vsize_uint(uint64(x))
}
