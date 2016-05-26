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
		*buf = append(*buf, byte(iPOS_1BYTE_MARKER|byte(get_bits(x, 6, 0)&0xFF)))
	case x == uint64(iPOS_2BYTE_MAX+1):
	default:
	}
}
