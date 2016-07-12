package wiredtiger

import "testing"
import "os"
import "bytes"

func TestMain(m *testing.M) {
	var res int
	/*	if res = initPackTest(); res != 0 {
		os.Exit(res)
	}*/

	res = m.Run()

	//deinitPackTest()

	os.Exit(res)
}

func intTest(t *testing.T, v int64) {
	var rv int64

	b, e := Pack("q", v)
	b2 := intPackTest(v)

	switch {
	case e != nil:
		t.Errorf("Expected something, got error: %v", e)
	case b == nil:
		t.Error("Expected []byte, nil return.")
	case len(b) == 0:
		t.Errorf("Expected []byte, got empty. len=%d cap=%d", len(b), cap(b))
	default:
		e = UnPack("q", b, &rv)

		switch {
		case e != nil:
			t.Errorf("Expected something, got error: %v ", e)
		case rv != v || bytes.Compare(b, b2) != 0:
			t.Errorf("Returned unexpected value %d != %d\n% x\n% x", v, rv, b, b2)
		default:
			t.Logf("Value %d = % x", v, b)
		}
	}
}

func uintTest(t *testing.T, v uint64) {
	var rv uint64

	b, e := Pack("Q", v)
	b2 := uintPackTest(v)

	switch {
	case e != nil:
		t.Errorf("Expected something, got error %v", e)
	case b == nil:
		t.Error("Expected []byte, nil return.")
	case len(b) == 0:
		t.Errorf("Expected []byte, got empty. len=%d cap=%d", len(b), cap(b))
	default:
		e = UnPack("Q", b, &rv)

		switch {
		case e != nil:
			t.Errorf("Expected something, got error: %v", e)
		case rv != v || bytes.Compare(b, b2) != 0:
			t.Errorf("Returned unexpected value %d != %d\n% x\n% x", v, rv, b, b2)
		default:
			t.Logf("Value %d = % x", v, b)
		}
	}
}

func TestUIntPackPosMax(t *testing.T) {
	uintTest(t, 18446744073709551615)
}

func TestIntPackPosMax(t *testing.T) {
	intTest(t, 9223372036854775807)
}

func TestIntPackPos3bMin(t *testing.T) {
	intTest(t, iPOS_2BYTE_MAX+257)
}

func TestIntPackPos2bSpec(t *testing.T) {
	intTest(t, iPOS_2BYTE_MAX+1)
}

func TestIntPackPos2bMax(t *testing.T) {
	intTest(t, iPOS_2BYTE_MAX)
}

func TestIntPackPos2bMin(t *testing.T) {
	intTest(t, iPOS_1BYTE_MAX+1)
}

func TestIntPackPos1bMax(t *testing.T) {
	intTest(t, iPOS_1BYTE_MAX)
}

func TestIntPack0(t *testing.T) {
	intTest(t, 0)
}

func TestIntPackNeg1bMax(t *testing.T) {
	intTest(t, -1)
}

func TestIntPackNeg1bMin(t *testing.T) {
	intTest(t, iNEG_1BYTE_MIN)
}

func TestIntPackNeg2bMax(t *testing.T) {
	intTest(t, iNEG_1BYTE_MIN-1)
}

func TestIntPackNeg2bMin(t *testing.T) {
	intTest(t, iNEG_2BYTE_MIN)
}

func TestIntPackNeg3bMax(t *testing.T) {
	intTest(t, iNEG_2BYTE_MIN-1)
}

func TestIntPackNegMin(t *testing.T) {
	intTest(t, -9223372036854775808)
}

func TestPack(t *testing.T) {
	b, e := Pack("xbBq3sSuu", int8(-2), uint8(2), int64(-9223372036854775808), "ABCD", "Hello\x00World", []byte{1, 2, 3}, []byte{4, 5, 6, 7})
	b2 := generalPackTest()

	switch {
	case e != nil:
		t.Errorf("Expected something, got error: %v", e)
	case b == nil:
		t.Error("Expected []byte, nil return.")
	case len(b) == 0:
		t.Errorf("Expected []byte, got empty. len=%d cap=%d", len(b), cap(b))
	default:
		var v1 int8
		var v2 uint8
		var v3 int64
		var v4 string
		var v5 string
		var v6 []byte
		var v7 []byte

		e = UnPack("xbBq3sSuu", b, &v1, &v2, &v3, &v4, &v5, &v6, &v7)

		switch {
		case e != nil:
			t.Errorf("Expected something, got error: %v", e)
			fallthrough
		default:
			t.Logf("Values % x", b)
			t.Logf("Values % x", b2)
			t.Logf("Values %d, %d, %d, \"%s\", \"%s\", [% x], [% x]", v1, v2, v3, v4, v5, v6, v7)
		}
	}
}

func TestOpen(t *testing.T) {
	os.Mkdir("data", 0777)
	defer os.RemoveAll("data")

	con, res := Open("data", "create")

	if res != nil {
		t.Errorf("Got error while open database: %v", res)
		return
	}

	defer con.Close("")

	session, er2 := con.OpenSession("")

	if er2 != nil {
		t.Errorf("Got error while open session: %v", er2)
		return
	}

	defer session.Close("")

	er3 := session.Create("table:access", "key_format=S,value_format=S")

	if er3 != nil {
		t.Errorf("Got error while open session: %v", er3)
		return
	}

	cursor, er4 := session.OpenCursor("table:access", nil, "")

	if er4 != nil {
		t.Errorf("Got error while open cursor: %v", er4)
		return
	}

	cursor.Close()

}
