package packet


import (
    "github.com/xxxsen/enc_socks/relay_msg"
    "errors"
    "hash/crc32"
    "github.com/golang/protobuf/proto"
    "encoding/binary"
    "github.com/xxxsen/enc_socks/utils"
    "github.com/xxxsen/enc_socks/codec"
    "fmt"
)

type RelayPacket struct {
    msgRelay relay_msg.RelayMsg
    msgBody relay_msg.RelayBody
    data []byte
    table *crc32.Table
    code codec.Codec
}

func NewRelayPacket(xcodec codec.Codec) *RelayPacket {
    return &RelayPacket{data:nil, table:crc32.MakeTable(crc32.IEEE), code:xcodec}
}

func(this *RelayPacket) SetData(data []byte) {
    this.data = data
}

func(this *RelayPacket) Encode() (int, error) {
    if this.msgBody.Cmd == nil || this.msgBody.Data == nil {
        return -1, errors.New("need more params")
    }
    //msg relay
    {
        data, err := proto.Marshal(&this.msgBody)
        if err != nil {
            return -2, errors.New("encode msg body err, msg:" + err.Error())
        }
        ck := crc32.Checksum(data, this.table)
        this.msgRelay.Body = data
        this.msgRelay.Crc32 = proto.Uint32(ck)
        this.msgRelay.Rnd = utils.NewRand().CreateBytes(int(utils.NewRand().CreateUInt32(0)))
    }
    {
        data, err := proto.Marshal(&this.msgRelay)
        if err != nil {
            return -3, errors.New("encode msg relay err, msg:" + err.Error())
        }
        if this.code != nil {
            var err error
            data, err = this.code.Encode(data)
            if err != nil {
                return -4, err
            }
        }
        total := len(data) + 4
        this.data = make([]byte, total)
        binary.BigEndian.PutUint32(this.data, uint32(total))
        copy(this.data[4:], data)
    }
    return 0, nil
}

func(this *RelayPacket) Decode() (int, error) {
    if this.data == nil || len(this.data) == 0 {
        return -1, errors.New("data empty")
    }
    if len(this.data) <= 4 {
        return -2, errors.New("data len invalid")
    }
    v := binary.BigEndian.Uint32(this.data)
    if len(this.data) < int(v) {
        return -3, errors.New("data len too short")
    }
    data := this.data[4:]
    if this.code != nil {
        dataLen, err := this.code.Check(data)
        if err != nil || dataLen <= 0 {
            return -4, errors.New(fmt.Sprint("codec check err, msg:%v", err))
        }
        data, err = this.code.Decode(data)
        if err != nil {
            return -10, errors.New(fmt.Sprintf("codec decode err, msg:%v", err))
        }
    }
    if err := proto.Unmarshal(data, &this.msgRelay); err != nil {
        return -5, errors.New(fmt.Sprintf("relay msg decode err, msg:%s, data len:%d", err.Error(), len(data)))
    }
    ckSum := crc32.Checksum(this.msgRelay.Body, this.table)
    if this.msgRelay.Crc32 == nil || ckSum != *this.msgRelay.Crc32 {
        return -6, errors.New("crc32 error")
    }
    if err := proto.Unmarshal(this.msgRelay.Body, &this.msgBody); err != nil {
        return -7, errors.New("relay body decode err, msg:" + err.Error())
    }
    if this.msgBody.Cmd == nil {
        return -8, errors.New("relay cmd nil")
    }
    return 0, nil
}

func(this *RelayPacket) GetData() []byte {
    return this.data
}

func(this *RelayPacket) GetMsgCmd() int32 {
    return this.msgBody.GetCmd()
}

func(this *RelayPacket) GetMsgData() []byte {
    return this.msgBody.GetData()
}

func(this *RelayPacket) SetMsgCmd(cmd int32) {
    this.msgBody.Cmd = proto.Int32(cmd)
}

func(this *RelayPacket) SetMsgData(data []byte) {
    this.msgBody.Data = data
}

func(this *RelayPacket) GetMsgBody() relay_msg.RelayBody {
    return this.msgBody
}

/*
    (ret > 0) - 数据未收全
    (ret > 0) - 数据已收全
    (ret < 0) -包错误
*/
func CheckRelay(data []byte) (int, error) {
    if len(data) <= 4 {
        return 0, nil
    }
    v := binary.BigEndian.Uint32(data)
    if v < 0 || v > 256 * 1024 {
        return -1, errors.New("packet len invalid");
    }
    if v <= uint32(len(data)) {
        return int(v), nil;
    }
    return 0, nil
}

func CheckRelayCodec(data []byte, codec codec.Codec) (int, error) {
    xlen, xerr := CheckRelay(data)
    if xerr != nil || xlen <= 0 {
        return xlen, xerr
    }
    len, err := codec.Check(data[4:])
    if len <= 0 || err != nil {
        msg := ""
        if err != nil {
            msg = err.Error()
        }
        return -2, errors.New("codec check fail, err:" + msg)
    }
    return xlen, nil
}
