package enc_socks

import (
    "net"
    "context"
    "enc_socks/relay"
    log "github.com/sirupsen/logrus"
    "sync"
    "io"
    "time"
    "reflect"
    "enc_socks/bytepool"
    "enc_socks/pipeline"
)

type RelayServer struct {
    config *ServerConfig
    bytePool *bytepool.BytePool
    svrPipe []pipeline.Pipeline
    cliPipe []pipeline.Pipeline
}

func NewRelayServer(cfg *ServerConfig, psvr []pipeline.Pipeline, pcli []pipeline.Pipeline) *RelayServer {
    return &RelayServer{config:cfg, bytePool:bytepool.NewPool(relay.PER_PACKET_DATA_SIZE), svrPipe:psvr, cliPipe:pcli}
}

func(this *RelayServer) loadTarget() (net.Conn, error) {
    conn, err := net.DialTimeout("tcp", this.config.RemoteAddr, this.config.Timeout)
    if err != nil {
        return conn, err
    }
    if this.config.ServerType == SERVER_TYPE_REMOTE {
        return conn, err
    } else {
        tmp := conn
        for _, pipe := range this.cliPipe {
            tmp, err = pipe.Process(tmp)
            if err != nil {
                return conn, err
            }
        }
        return tmp, nil
    }
}

func(this *RelayServer) Start() {
    listener, err := net.Listen("tcp", this.config.LocalAddr)
    if err != nil {
        log.Errorf("Listen addr:%s failed, err:%s", this.config.LocalAddr, err.Error())
        return
    }
    sessionId := uint32(0)
    for {
        tmp, err := listener.Accept()
        if err != nil {
            log.Errorf("Recv conn failed, err:%s, wait", err.Error())
            time.Sleep(10 * time.Millisecond)
            continue
        }
        sessionId++
        go func() {
            conn := tmp
            if this.config.ServerType == SERVER_TYPE_REMOTE {
                for _, pipe := range this.svrPipe {
                    conn, err = pipe.Process(conn)
                    if err != nil {
                        log.Errorf("Client check failed, pipe:%s, err:%s", reflect.TypeOf(pipe).String(), err.Error())
                        tmp.Close()
                        return
                    }
                }
            }
            log.Infof("Recv connection from local, addr:%s, mark sessionid:%d", conn.RemoteAddr(), sessionId)
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

func(this *RelayServer) datacopy(lhs net.Conn, rhs net.Conn) (int, error) {
    buf := this.bytePool.Get()
    defer func() {
        this.bytePool.Put(buf)
    } ()
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
            err := local.Close()
            if err != nil {
                log.Errorf("Close local connection failed, err:%s, session:%d", err.Error(), sessionId)
            }
        }
        if remote != nil && !reflect.ValueOf(remote).IsNil() {
            err := remote.Close()
            if err != nil {
                log.Errorf("Close remote connection failed, err:%s, session:%d", err.Error(), sessionId)
            }
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
            _, err := this.datacopy(remote, local)
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
            _, err := this.datacopy(local, remote)
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

