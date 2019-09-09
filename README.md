一个简单的TCP长链的网络库，可以很简便的开发自己的server和client, 以此来实现基于tcp长链的主动请求，推送等业务；

- TcpStream支持两种角色： 客户端主动发起连接的Connection，服务端Accept后派生出来的TcpStream;
- 内部协议的实现集成了tcp数据自动分包的功能，同时Header预留了Seq和MsgType的字段给业务方使用；Seq主要用于多路复用的场景下，request与resp的对应关系；MsgType可用于业务包类型的序列化和反序列化；
- client TcpStream默认支持自动重连，连接断开后(比如server重启)会每隔5s尝试一次重连；
- tcpstream支持销毁（stop)的功能，被stop的client不再自动重连，如需要重新建连需要再new一个TcpStream;
- 没有默认心跳包的支持, 如需要业务层自己实现；

Sync调用功能：
- client与server建立一个连接后，可通过RequestID的映射功能，实现请求方类似RPC调用功能；
- 基于一条连接可以实现双向RPC调用的效果；这样就解决了server向client发起rpc调用复杂的问题，client端可以不用再启用rpcServer.
- 同步调用接口的业务包序列化和反序列化有自己来实现；


一个client和server的例子

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

同步调用客户端的例子

- 创建一个客户端

```golang
	c := NewSyncClient("127.0.0.1:7001")
```

基于已经存在tcpstream的创建，client端和server都可以

```golang
	c := NewSyncFromTcpStream(tcpstream)
```

- 调用
```go

	if resp, err := c.Call(&m, time.Duration(2*time.Second)); err != nil {
		fmt.Println("call fail ", err.Error())
	} else {
		//反序列化resp，do something
	}

```