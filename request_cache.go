package tcpstream

import (
	"errors"
	"fmt"
	"sync"
)

const mapNum = 8

type requestMap struct {
	mutex    sync.Mutex
	respChan map[uint64]chan *Message
}

type RequestManager struct {
	requestMaps [mapNum]requestMap
	current     uint32
}

func NewRequestManager() *RequestManager {
	manager := &RequestManager{}
	for i := 0; i < mapNum; i++ {
		manager.requestMaps[i].respChan = make(map[uint64]chan *Message)
	}

	return manager
}

func (r *RequestManager) cache(id uint64, msg chan *Message) {
	m := &r.requestMaps[id%mapNum]
	m.mutex.Lock()
	m.respChan[id] = msg
	m.mutex.Unlock()
}

func (r *RequestManager) shoot(id uint64, msg *Message) (err error ){
	m := &r.requestMaps[id%mapNum]

	m.mutex.Lock()
	respChans, exist := m.respChan[id]
	if exist {
		delete(m.respChan, id)
		m.mutex.Unlock()
		//如果此时无法写入，说明resp_chan已经不可用，要立马返回
		select {
		case respChans <- msg:
			//只使用一次，写入完成后关闭；
			close(respChans)
		default:
			err = errors.New(fmt.Sprint("Default fail !!!!! request id : ", id))
		}
	} else {
		m.mutex.Unlock()
		err = errors.New("default fail ")
	}
	return
}

func (r *RequestManager) dispose(id uint64) {
	m := &r.requestMaps[id%mapNum]
	m.mutex.Lock()
	respChans, exist := m.respChan[id]
	if exist {
		delete(m.respChan, id)
		m.mutex.Unlock()
		close(respChans)
	} else {
		//r.disposeWait.Done()
		m.mutex.Unlock()
	}
}
