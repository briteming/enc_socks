package enc_socks

import (
    "net"
    "context"
    "enc_socks/relay"
    log "github.com/sirupsen/logrus"
    "sync"
    "io"
    "time"
    "crypto/tls"
    "reflect"
)

type RelayServer struct {
    config *ServerConfig
}

func NewRelayServer(cfg *ServerConfig) *RelayServer {
    return &RelayServer{config:cfg}
}

func(this *RelayServer) loadServer() (net.Listener, error) {
    if this.config.ServerType == SERVER_TYPE_REMOTE {
        cert, err := tls.LoadX509KeyPair(this.config.TlsServerPemAddr, this.config.TlsServerKeyAddr)
        if err != nil {
            return nil, err
        }
        config := &tls.Config{Certificates: []tls.Certificate{cert}}
        ln, err := tls.Listen("tcp", this.config.LocalAddr, config)
        if err != nil {
            return ln, err
        }
//        ln, _ := net.Listen("tcp", this.config.LocalAddr)
        return relay.NewRelayAcceptor(ln, &this.config.UserInfo), nil
    } else {
        return net.Listen("tcp", this.config.LocalAddr)
    }
}

func(this *RelayServer) loadTarget() (net.Conn, error) {
    if this.config.ServerType == SERVER_TYPE_REMOTE {
        return net.DialTimeout("tcp", this.config.RemoteAddr, this.config.Timeout)
    } else {
        conf := &tls.Config{
            InsecureSkipVerify: true,
        }
        conn, err := tls.Dial("tcp", this.config.RemoteAddr, conf)
        if err != nil {
            return conn, err
        }
//        conn, _ := net.DialTimeout("tcp", this.config.RemoteAddr, this.config.Timeout)
        return relay.DialWithConn(conn, &this.config.User, this.config.Timeout)
    }
}

func(this *RelayServer) Start() {
    listener, err := this.loadServer()
    if err != nil {
        log.Errorf("Listen addr:%s failed, err:%s", this.config.LocalAddr, err.Error())
        return
    }
    var sessionId uint32 = 0;
    log.Printf("Server start on addr:%s", this.config.LocalAddr)
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Errorf("Rcv conn failed, err:%s", err)
            listener.Close()
            return
        }
        sessionId++
        log.Infof("Recv connection from local, addr:%s, mark sessionid:%d", conn.RemoteAddr(), sessionId)
        go func() {
            this.handleConnection(conn, sessionId)
        }()
    }
}

func isDone(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return true
    default:
        return false
    }
}

func datacopy(lhs net.Conn, rhs net.Conn) (int, error) {
    buf := make([]byte, relay.PER_PACKET_DATA_SIZE * 2)
    cnt, err := lhs.Read(buf)
    if cnt > 0 {
        data := buf[0:cnt]
        dataLen := len(data)
        writeIndex := 0
        for ; writeIndex < dataLen; {
            wcnt, err := rhs.Write(data[writeIndex:])
            if err != nil {
                return wcnt, err
            }
            writeIndex += wcnt
        }
    }
    return cnt, err
}

func(this *RelayServer) handleConnection(local net.Conn, sessionId uint32) {
    remote, err := this.loadTarget()
    defer func() {
        if local != nil && !reflect.ValueOf(local).IsNil() {
            local.Close()
        }
        if remote != nil && !reflect.ValueOf(remote).IsNil() {
            remote.Close()
        }
    }()
    if err != nil {
        log.Errorf("Create remote connection failed, session, err:%s:%d", err.Error(), sessionId)
        return
    }
    log.Infof("Dial connection to remote success, remote:%s", remote.LocalAddr())

    var waiter sync.WaitGroup
    waiter.Add(2)
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        defer func() {
            waiter.Done()
            cancel()
        }()
        for ; ; {
            remote.SetReadDeadline(time.Now().Add(this.config.Timeout))
            _, err := datacopy(remote, local)
            if err != nil {
                if err == io.EOF {
                    log.Infof("Local read eof, sessionid:%d", sessionId)
                    break
                } else if err, ok := err.(net.Error); ok && err.Timeout() {
                    if isDone(ctx) {
                        return
                    }
                } else {
                    log.Errorf("Read local write remote failed, err:%s, sessionid:%d", err.Error(), sessionId)
                    return
                }
            }
        }
    }()
    go func() {
        defer func() {
            waiter.Done()
            cancel()
        }()
        for ; ; {
            _, err := datacopy(local, remote)
            if err != nil {
                if err == io.EOF {
                    log.Infof("Remote read eof, sessionid:%d", sessionId)
                    break
                } else if err, ok := err.(net.Error); ok && err.Timeout() {
                    if isDone(ctx) {
                        return
                    }
                } else {
                    log.Errorf("Read remote write local failed, err:%s, sessionid:%d", err.Error(), sessionId)
                    return
                }
            }
        }
    }()
    waiter.Wait()
    log.Infof("Read/Write finish, sessionid:%d", sessionId)
}

