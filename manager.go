package tcpstream

import (
	"math/rand"
	"sync"
	"time"
)

const mapCount = 8

type Manager struct {
	connMaps    [mapCount]connMap
	allConns    []*TcpStream
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
	rwMutex     sync.RWMutex
}

type connMap struct {
	sync.RWMutex
	conns map[uint64]*TcpStream
}

func (s *connMap) RandomSelectFromMap() *TcpStream {
	var array_conns []*TcpStream
	for _, conn := range s.conns {
		array_conns = append(array_conns, conn)
	}

	if len(array_conns) == 0 {
		return nil
	}

	index := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(array_conns))

	return array_conns[index]
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.connMaps); i++ {
		manager.connMaps[i].conns = make(map[uint64]*TcpStream)
	}

	return manager
}

func (m *Manager) Dispose() {
	m.disposeOnce.Do(func() {
		m.disposeFlag = true
		for i := 0; i < mapCount; i++ {
			smap := &m.connMaps[i]
			smap.Lock()
			for _, c := range smap.conns {
				c.Stop()
			}
			smap.Unlock()
		}
		m.disposeWait.Wait()
	})
}

func (m *Manager) Get(id uint64) *TcpStream {
	cmap := &m.connMaps[id%mapCount]
	cmap.RLock()
	defer cmap.RUnlock()
	conn, _ := cmap.conns[id]
	return conn
}

//对所有的connection进行广播处理
func (m *Manager) Broadcast(handler func(*TcpStream)) {

	for i := 0; i < mapCount; i++ {
		cmap := &m.connMaps[i]
		//TODO: 注意这个地方的并发安全问题
		cmap.RLock()
		for _, v := range cmap.conns {
			handler(v)
		}
		cmap.RUnlock()
	}
}

func (m *Manager) Put(conn *TcpStream) {

	cmap := &m.connMaps[conn.ID%mapCount]
	cmap.Lock()
	defer cmap.Unlock()

	cmap.conns[conn.ID] = conn
	m.disposeWait.Add(1)
}

func (m *Manager) Remove(conn *TcpStream) {
	if m.disposeFlag {
		m.disposeWait.Done()
		return
	}
	cmap := &m.connMaps[conn.ID%mapCount]
	cmap.Lock()
	defer cmap.Unlock()
	delete(cmap.conns, conn.ID)
	m.disposeWait.Done()
}
