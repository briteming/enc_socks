# repo废弃, 迁移到 https://github.com/xxxsen/wcf 

# enc_socks
一个简单的加密通道

## 为什么写这个东西

还是因为无聊啊=。=, 这个东西跟之前的shadowsocks还是有一些区别的, 它仅仅只是一个加密通道, 并不涉及到具体的协议解析, shadowsocks本身是一个前后端分离的socks代理, 会解开socks协议的包并重新封装成自己的协议再进行转发, 而这个东西仅仅只是透传数据, 它自身并无法提供代理, 需要有一个真正的代理服务器才行。
简单的示意图, 嘿嘿。
```
shadowsocks                                   |
+--------------+         +--------------+     |      +---------------+           +--------------------+
|              |         |              |     |      |               |           |                    |
| browser      +--------->  local       +------------>   remote      +----------->   internet         |
|              |         |              |     |      |               |           |                    |
+--------------+         +--------------+     |      +---------------+           +--------------------+
                                              |
                                              |
                                              |
+--------------+         +--------------+     |      +----------------+          +----------------+       +-------------------+
|              |         |              |     |      |                |          |                |       |                   |
|  browser     +--------->   local      +------------>    remote      +---------->    proxy       +------->    internet       |
|              |         |              |     |      |                |          |                |       |                   |
+--------------+         +--------------+     |      +----------------+          +----------------+       +-------------------+
enc_socks                                     |
                                              |
                                              |
                                              |  +-------------+
                                              +--+             |
                                                 |   GFW       |
                                                 |             |
                                                 +-------------+

```

## 编译
1. go get github.com/xxxsen/enc_socks
2. 进到${GOPATH}/src/github.com/xxxsen/enc_socks/cmd 目录
3. 执行go build即可

## 参数
* --type, server类型, 可选**local**, **remote**, 分别启动为本地端和远程端。
* --svr_pem, 服务端加载的pem文件(可以使用cmd目录下的create_tls_data.sh生成)
* --svr_key, 服务端加载的key文件(可以使用cmd目录下的create_tls_data.sh生成)
* --timeout, 链接/读写数据超时的时间, 单位为秒
* --local, 本地监听地址, 例如:"0.0.0.0:8848"
* --remote, 远程server地址, 例如"127.0.0.1:8849"

## 示例

### 本地端
```shell
./cmd --type=local --local="0.0.0.0:8848" --remote="${你服务器的地址}" --user="xxx' --pwd="hello_world" --timeout=3 
```
### 远程端
```shell
./cmd --type=remote --local="0.0.0.0:8849" --remote="${你的代理服务器的地址}" --svr_pem="./server.pem" --svr_key="./server.key" --timeout=3 --user="xxx" --pwd="hello_world"
```
