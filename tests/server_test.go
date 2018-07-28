package tests

import (
    "testing"
    "time"
    "enc_socks"
    "enc_socks/pipeline"
    "strings"
    "io/ioutil"
    "github.com/sirupsen/logrus"
    "os"
    "encoding/json"
)

func TestServer(t *testing.T) {
    addr := "127.0.0.1:1081"
    setting := "/home/sen/GoPath/src/enc_socks/cmd/config.json"


    //加载管道配置
    mp := make(map[string]interface{})
    data, err := ioutil.ReadFile(setting)
    if err != nil {
        logrus.Errorf("Load pipe args file failed, err:%s", err.Error())
    } else {
        err = json.Unmarshal(data, &mp)
        if err != nil {
            logrus.Errorf("Parse pipe file json err, msg:%s", err.Error())
            os.Exit(2)
        }
    }

    cfg := &enc_socks.ServerConfig{
        LocalAddr:"127.0.0.1:8848", RemoteAddr:addr, Timeout:3 * time.Second, ServerType:2}
    args := "auth"
    holder := pipeline.GetHolder()
    //初始化管道
    pipes := strings.Split(args, "#")
    var clientPipe []pipeline.Pipeline
    var serverPipe []pipeline.Pipeline
    for _, pipe := range pipes {
        fac, err := holder.GetByName(pipe)
        if err != nil {
            logrus.Errorf("Get pipe factory failed, err:%s", err.Error())
            os.Exit(1)
        }
        cfg, ok := mp[pipe]
        if !ok {
            logrus.Errorf("Get pipe:%s setting failed, setting not found!", pipe)
            os.Exit(3)
        }
        c := cfg.(map[string]interface{})
        err = fac.Init(&c)
        if err != nil {
            logrus.Errorf("Factory:%s init with setting:%v failed, err:%s", pipe, c, err.Error())
        }
        clientPipe = append(clientPipe, fac.GetCliPipe())
        serverPipe = append(serverPipe, fac.GetSvrPipe())
    }
    svr := enc_socks.NewRelayServer(cfg, serverPipe, clientPipe)
    svr.Start()
}

