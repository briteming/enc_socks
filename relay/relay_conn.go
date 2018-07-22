package relay

import (
    "net"
    "time"
    "errors"
    "fmt"
    "enc_socks/relay_msg"
    "enc_socks/bytepool"
)

type RelayConn struct {
    conn net.Conn
    rbuf []byte
    rIndex int
    wbuf []byte
    wIndex int
    dbuf []byte
}

var bytePool = bytepool.NewPool(PER_PACKET_DATA_SIZE)

func NewRelayConn(conn net.Conn) *RelayConn {
    return &RelayConn{conn:conn, rbuf:bytePool.Get(), wbuf:bytePool.Get(), dbuf:nil, rIndex:0, wIndex:0}
}

func(this *RelayConn) Read(data []byte) (int, error) {
    if (len(this.dbuf) != 0) {
        ret := copy(data, this.dbuf)
        this.dbuf = this.dbuf[ret:]
        return ret, nil
    }
    cnt, err := this.conn.Read(this.rbuf[this.rIndex:])
    if err != nil {
        return cnt, err
    }
    this.rIndex += cnt
    decodeIndex := 0
    for {
        ck, err := Check(this.rbuf[decodeIndex:this.rIndex])
        if err != nil {
            return 0, errors.New("check buffer failed, err:" + err.Error())
        }
        if ck == 0 {
            break
        }
        pkt := NewPacket()
        pkt.SetPacketData(this.rbuf[decodeIndex:decodeIndex + ck])
        decodeIndex += ck
        err = pkt.Decode()
        if err != nil {
            return 0, errors.New("decode pkt failed, err:" + err.Error())
        }
        if pkt.GetCmd() == int32(relay_msg.CMD_TYPE_CMD_DATA) {
            this.dbuf = append(this.dbuf, pkt.GetBodyData()...)
        }
    }
    if(decodeIndex == 0) {
        return 0, nil  //wait more data...
    }
    if decodeIndex < this.rIndex {
        copy(this.rbuf, this.rbuf[decodeIndex:this.rIndex])
        this.rIndex -= decodeIndex
    } else if(decodeIndex == this.rIndex) {
        this.rIndex = 0
    } else {
        panic(fmt.Sprintf("why decodeIndex > rIndex? decodeIndex:%d, rIndex:%d", decodeIndex, this.rIndex))
    }
    //
    ret := copy(data, this.dbuf)
    this.dbuf = this.dbuf[ret:]
    return ret, nil
}

func(this *RelayConn) Write(data []byte) (int, error) {
    if len(data) >= PER_PACKET_DATA_SIZE {
        data = data[0:PER_PACKET_DATA_SIZE / 3 * 2]
    }
    pkt := NewPacket();
    pkt.SetBodyData(data)
    pkt.SetCmd(int32(relay_msg.CMD_TYPE_CMD_DATA))
    err := pkt.Encode()
    if err != nil {
        return 0, errors.New("pkt encode failed, err:" + err.Error())
    }
    writeData := pkt.GetBodyData()
    writeIndex := 0
    writeTotal := len(writeData)
    for ; writeIndex < writeTotal; {
        cnt, err := this.conn.Write(writeData[writeIndex:])
        if err != nil {
            return cnt, errors.New(fmt.Sprintf("write data failed, err:%s, cnt:%d", err.Error(), cnt))
        }
        writeIndex += cnt
    }
    return len(data), nil
}

func(this *RelayConn) Close() error {
    defer func() {
        bytePool.Put(this.rbuf)
        bytePool.Put(this.wbuf)
    } ()
    err := this.conn.Close()
    if err != nil {
        return err
    }
    if this.rIndex != 0 || this.wIndex != 0 || len(this.dbuf) != 0 {
        return errors.New(fmt.Sprintf("buffer spare in connection, rs:%d, ws, ds:%d", this.rIndex, this.wIndex, len(this.dbuf)))
    }
    return nil
}

func(this *RelayConn) LocalAddr() net.Addr {
    return this.conn.LocalAddr()
}

func(this *RelayConn) RemoteAddr() net.Addr {
    return this.conn.RemoteAddr()
}

func(this *RelayConn) SetDeadline(t time.Time) error {
    return this.conn.SetDeadline(t)
}

func(this *RelayConn) SetReadDeadline(t time.Time) error {
    return this.conn.SetReadDeadline(t)
}

func(this *RelayConn) SetWriteDeadline(t time.Time) error {
    return this.conn.SetWriteDeadline(t)
}