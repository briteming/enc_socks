package relay

import (
    "net"
    "context"
    "github.com/sirupsen/logrus"
    "time"
    "encoding/binary"
    "errors"
)

type RelayAcceptor struct {
    listener net.Listener
    authMap *AuthMap
    conn chan net.Conn
    ctx context.Context
    cancel context.CancelFunc
}

func NewRelayAcceptor(listener net.Listener, authMap *AuthMap) *RelayAcceptor {
    relay := &RelayAcceptor{}
    ctx, cancel := context.WithCancel(context.Background())
    relay.ctx = ctx
    relay.cancel = cancel
    relay.conn = make(chan net.Conn, 50)
    relay.listener = listener
    relay.authMap = authMap
    go func() {
        relay.acceptInner()
    }()
    return relay
}

func isDone(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return true
    default:
        return false
    }
}

func(this *RelayAcceptor) handleConnect(conn net.Conn) {
    buf := make([]byte, 4)
    readIndex := 0
    wouldClose := true
    defer func() {
        if wouldClose {
            conn.Close()
        }
    }()
    for {
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        cnt, err := conn.Read(buf[readIndex:])
        if err != nil {
            logrus.Errorf("Connection read cause err, addr:%s, skip")
            return
        }
        readIndex += cnt
        if isDone(this.ctx) {
            logrus.Errorf("Server exit, skip process connection:%s", conn.RemoteAddr().String())
            return
        }
        if readIndex != len(buf) {
            continue
        }
        if readIndex == 4 {  //data len
            total := int(binary.BigEndian.Uint32(buf))
            if total > PER_PACKET_DATA_SIZE {
                logrus.Errorf("Invalid auth len, len:%d, connection:%s", total, conn.RemoteAddr().String())
                return
            }
            tmp := buf
            buf = make([]byte, total)
            copy(buf, tmp)
        } else {
            cr, err := Check(buf)
            if err != nil || cr == 0 {
                logrus.Errorf("Check buffer failed, len:%d, err:%s",cr,err)
                return
            }
            pkt := NewPacket()
            pkt.SetPacketData(buf)
            err = pkt.Decode()
            if err != nil {
                logrus.Errorf("Check packet and decode failed, err:%s, connection:%s", err.Error(), conn.RemoteAddr().String())
                return
            }
            auth, err := GetRelayAuthData(pkt)
            if err != nil {
                logrus.Errorf("Get auth data failed, err:%s", err.Error())
                return
            }
            ret := this.authMap.Check(auth)
            if !ret {
                logrus.Errorf("Check auth failed, invalid user:%s, pwd:%s", auth.User, auth.Pwd)
                return
            }
            wouldClose = false
            conn.SetReadDeadline(time.Time{})
            this.conn <- conn
            return
        }
    }
}

func(this *RelayAcceptor) acceptInner() {
    defer func() {
        this.cancel()
    }()
    for ; ; {
        conn, err := this.listener.Accept()
        if err != nil {
            logrus.Errorf("Accept conn failed, err:%s", err.Error());
            break
        }
        go func() {
            this.handleConnect(conn)
        } ()
    }
}

func(this *RelayAcceptor) Accept() (net.Conn, error) {
    select {
        case cn := <-this.conn:
            return cn, nil
        case <-this.ctx.Done():
            return nil, errors.New("Server exit, acceptor exit!")

    }
}

func(this *RelayAcceptor) Close() error {
    return this.listener.Close()
}

func(this *RelayAcceptor) Addr() net.Addr {
    return this.listener.Addr()
}