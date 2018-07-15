package packet

import (
    "testing"
    "enc_socks/codec"
    "encoding/hex"
)

var codec_key = "hello test"
var codec_name = "xor"
var pkt_data = []byte("hello this is a test!!!")
var pkt_cmd = 1;

func TestEDodec(t *testing.T) {
    c, err := codec.NewFactory().GetCodec(codec_name)
    if err != nil {
        t.Errorf("Get codec fail, err:%s", err.Error())
        return
    }
    c.Init(codec_key)
    t.Logf("Raw hex:%s", hex.EncodeToString(pkt_data))
    pkt := NewRelayPacket(c)
    pkt.SetMsgCmd(int32(pkt_cmd))
    pkt.SetMsgData(pkt_data)
    ret, eerr := pkt.Encode()
    if eerr != nil || ret != 0 {
        t.Errorf("Enc failed, e:%v, ret:%d", eerr, ret)
        return
    }
    t.Logf("Enc hex:%s, data len:%d", hex.EncodeToString(pkt.GetData()), len(pkt.GetData()))
    pkt2 := NewRelayPacket(c)
    pkt2.SetData(pkt.GetData())
    ret2, derr := pkt2.Decode()
    if derr != nil {
        t.Errorf("Decode err, ret:%d, msg:%v", ret2, derr)
        return
    }
    if string(pkt2.GetMsgData()) != string(pkt_data) || pkt2.GetMsgCmd() != pkt.GetMsgCmd() || pkt2.msgRelay.GetCrc32() != pkt.msgRelay.GetCrc32() {
        t.Errorf("After decode, data != raw_data, old hex:%s, new hex:%s", hex.EncodeToString(pkt_data), hex.EncodeToString(pkt2.GetMsgData()))
    }
    t.Logf("After decode, hex decode data:%s, crc32:%u, cmd:%d, pkg serialized:%s", hex.EncodeToString(pkt2.GetMsgData()), pkt2.msgRelay.GetCrc32(), pkt2.msgBody.GetCmd(), pkt2.GetMsgBody())
}

