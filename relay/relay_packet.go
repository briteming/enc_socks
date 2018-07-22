package relay


import (
    "enc_socks/relay_msg"
    "errors"
    "github.com/golang/protobuf/proto"
    "encoding/binary"
    "math/rand"
    "time"
    "fmt"
)

//len(4) + 0x2(1) + RelayPacket + 0x3(1)

const (
    STX = 0x2
    ETX = 0x3
    PER_PACKET_DATA_SIZE = 64 * 1024
)

type RelayPacket struct {
    pkt relay_msg.RelayPacket
    data []byte
    rnd *rand.Rand
}

func NewPacket() *RelayPacket {
    return &RelayPacket{data:nil, rnd:rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func(this *RelayPacket) SetBodyData(data []byte) {
    this.pkt.Data = data
}

func(this *RelayPacket) SetCmd(cmd int32) {
    this.pkt.Cmd = proto.Int32(cmd)
}

func(this *RelayPacket) GetCmd() int32 {
    return this.pkt.GetCmd()
}

func(this *RelayPacket) GetBodyData() []byte {
    return this.pkt.Data
}

func(this *RelayPacket) Encode() error {
    dLen := len(this.pkt.Data)
    rndLen := 10
    if dLen < 128 {
        rndLen = 128
    }
    rndLen = this.rnd.Int() % rndLen + 5
    rndData := make([]byte, rndLen)
    rndDataEnd := make([]byte, rndLen)
    for i := 0; i < rndLen; i++ {
        rndData[i] = byte(this.rnd.Int() % 256)
        rndDataEnd[i] = byte(this.rnd.Int() % 256)
    }
    this.pkt.RndPadFront = rndData
    this.pkt.RndPadEnd = rndDataEnd
    mdata, merr := proto.Marshal(&this.pkt)
    if merr != nil {
        return errors.New("encode pkt failed, err:" + merr.Error())
    }
    total := 4 + 2 + len(mdata)
    this.data = make([]byte, total)
    index := 0
    binary.BigEndian.PutUint32(this.data, uint32(total))
    index += 4
    this.data[index] = STX
    index++
    copy(this.data[index:], mdata)
    index += len(mdata)
    this.data[index] = ETX
    return nil
}

func(this *RelayPacket) GetPacketData() []byte {
    return this.data
}

func(this *RelayPacket) SetPacketData(data []byte) {
    this.data = data
}

func(this *RelayPacket) Decode() error {
    xlen, err := Check(this.data)
    if err != nil {
        return err
    }
    if xlen == 0 {
        return errors.New("invalid data, need more data")
    }
    data := this.data[5:xlen - 1]
    err = proto.Unmarshal(data, &this.pkt)
    if err != nil {
        return errors.New("pb pkt decode err, msg:" + err.Error())
    }
    return nil
}

func Check(data []byte) (int, error) {
    if len(data) <= 6 {
        return 0, nil
    }
    xlen := int(binary.BigEndian.Uint32(data))
    if xlen > 256 * 1024 {
        return -1, errors.New(fmt.Sprintf("invalid packet size:%d", xlen))
    }
    if len(data) < xlen {
        return 0, nil
    }
    if data[4] != STX || data[xlen - 1] != ETX {
        return -2, errors.New("invalid packet startswith/endswith")
    }
    return xlen, nil
}

func CreateRelayAuthReq(msg *AuthMsg) *RelayPacket {
    req := &relay_msg.RelayAuthReq{User:&msg.User, Pwd:&msg.Pwd}
    pkt := NewPacket()
    data, err := proto.Marshal(req)
    if err != nil {
        return nil
    }
    pkt.SetBodyData(data)
    pkt.SetCmd(int32(relay_msg.CMD_TYPE_CMD_AUTH))
    return pkt
}

func GetRelayAuthData(pkt *RelayPacket) (*AuthMsg, error) {
    req := &relay_msg.RelayAuthReq{}
    err := proto.Unmarshal(pkt.GetBodyData(), req)
    if err != nil {
        return nil, err
    }
    return NewAuthMsg(req.GetUser(), req.GetPwd()), nil
}
