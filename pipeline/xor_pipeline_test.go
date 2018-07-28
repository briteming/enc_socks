package pipeline

import "testing"

func TestXor(t *testing.T) {
    basic := "hello this is a test"
    a := []byte(basic)
    b := []byte(basic)
    xor := XorConn{}
    xor.xor(a)
    t.Logf("a == b ? %t\n", string(a) == string(b))
    xor.xor(a)
    t.Logf("a == b ? %t\n", string(a) == string(b))
}
