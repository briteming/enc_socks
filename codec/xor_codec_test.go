package codec

import (
    "testing"
    "encoding/hex"
)

func TestXor(t *testing.T) {
    c, err := NewFactory().GetCodec("xor")
    if err != nil {
        t.Errorf("Get codec xor fail, err:" + err.Error())
        return
    }
    raw := []byte("are you ok?")
    c.Init("hello world")
    t.Logf("Raw data:%s", hex.EncodeToString(raw))
    enc, eerr := c.Encode([]byte(raw))
    if eerr != nil {
        t.Errorf("Encode fail, err:" + eerr.Error())
        return
    }
    t.Logf("Encode data:enc:%s", hex.EncodeToString(enc))
    dec, derr := c.Decode(enc)
    if derr != nil {
        t.Errorf("Decode fail, err:" + derr.Error())
        return
    }
    if string(dec) != string(raw) {
        t.Errorf("Decode but result not equal, raw:%s, dec:%s", hex.EncodeToString(raw), hex.EncodeToString(dec))
        return
    }
    t.Logf("Decode succ, data hex:%s", hex.EncodeToString(dec))
}
