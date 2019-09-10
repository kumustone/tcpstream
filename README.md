一个简单的TCP长链的网络库，可以帮助你很简便的开发基于tcp socket连接的业务。同时封装了一个类似于rpc功能的同步调用接口，可以模拟双向RPC调用，服务端主动推送，广播等业务功能。

整个库的核心在于tcpstream，主要支持一下功能：
- 支持Tcp字节流自动分包功能，使用者感知到的是msg层面的数据，不用关心底层的数据收发；
- 支持client主动发起的Tcpstream，和Server端Accept派生出来的TcpStream，除了重连功能外，二者其他的调用接口完全相同；SyncClient基于同一条TcpStream实现双向RPC调用；
- 支持多路复用，多个RPC请求使用一条socket连接即可，内部通信是全异步；在用户端看来SynClient是同步功能；
- Client的tcpstream默认支持自动重连功能，连接断开后每隔5s向对端发起重连，如果不想重连调用stop接口销毁次tcpstream即可；
- 不支持默认的心跳包功能，如业务方需要自己在上层实现；



Sync调用功能：
- client与server建立一个连接后，可通过RequestID的映射功能，实现请求方类似RPC调用功能；
- 基于一条连接可以实现双向RPC调用的效果；这样就解决了server向client发起rpc调用复杂的问题，client端可以不用再启用rpcServer.
- 同步调用接口的业务包序列化和反序列化有自己来实现；



同步调用客户端的例子

- 创建一个客户端

```golang
	c := NewSyncClient("127.0.0.1:7001")
```

基于已经存在tcpstream的创建，client端和server都可以

```golang
	c := NewSyncFromTcpStream(tcpstream)
```

- 同步调用，输入是要发送的msg，接收到的是对端返回的msg；然后对msg反序列化即可实现自己的业务；
```go

	if resp, err := c.Call(&m, time.Duration(2*time.Second)); err != nil {
		fmt.Println("call fail ", err.Error())
	} else {
		//反序列化resp，do something
	}

```


一个client和server的例子

server
```go

type Handler struct {
	name string
}

func (h *handler) OnData(conn *TcpStream, msg *Message) error {
	_ = conn.Write(&Message{
		//因为是多路复用的，返回的seq必须要与client请求的seq对应
		Header: ProtoHeader{
			Seq: msg.Header.Seq,
		},
		Body: []byte("I am server"),
	})
	return nil
}
func (h *handler) OnConn(conn *TcpStream)    {}
func (h *handler) OnDisConn(conn *TcpStream) {}

func main() {
	tcpstream.NewTCPServer(address, &Handler{}).Serve()
	select {
	}
}

```

client：
```go

var address = "127.0.0.1:7001"

type Handler struct{}

func (h *Handler) OnData(conn *tcpstream.TcpStream, msg *tcpstream.Message) error {
	fmt.Println("receive from server: ", string(msg.Body))
	resp := &tcpstream.Message{
		Header: msg.Header,
		Body:   []byte("I am client"),
	}
	time.Sleep(time.Second * 1)
	return conn.Write(resp)
}

func (h *Handler) OnConn(conn *tcpstream.TcpStream) {
	fmt.Println("Connection to ", conn.RemoteAddr)

	_ = conn.Write(&tcpstream.Message{
		Body: []byte("I am client"),
	})
}

func (h *Handler) OnDisConn(conn *tcpstream.TcpStream) {
	fmt.Println("DisConnection to ", conn.RemoteAddr)
}

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	for i := 0; i < 100; i++ {
		go func() {
			client := tcpstream.NewConnTcpStream(address, &Handler{})
			client.Start()
			time.Sleep(10 * time.Second)
			client.Stop()
		}()
	}
	select {}
}
```
