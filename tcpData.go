package tcp

import (
	"sync"
	"time"
)

// 用于缓存消息收发记录
type MsgData struct {
	Cmd       string // 消息ID
	StartTime int64  // 发送消息时间戳 nanoseconds
	EndTime   int64  // 接到返回消息时间戳 nanoseconds
	Sus       bool   // 运行结果
	Fn        bool   // 是否完成
}

// 有效数据
func (s *MsgData) GetMiniData() MsgMiniData {
	return MsgMiniData{
		Cmd:      s.Cmd,
		Duration: ((s.EndTime - s.StartTime) / int64(time.Millisecond)),
		// Duration: (s.EndTime - s.StartTime),
		Sus: s.Sus,
	}
}

// 返回的数据
type MsgMiniData struct {
	Cmd      string // 消息ID
	Duration int64  // 消息往返间隔ms
	Sus      bool   // 执行结果
}

// 数据管理器
type MsgDataManager struct {
	// mutex for JSCallBackFunc
	mutex sync.Mutex
	Datas []*MsgData
}

// 发送消息是调用
func (m *MsgDataManager) StoreSendMsg(msg *string) {
	defer func() {
		m.mutex.Unlock()
	}()
	m.mutex.Lock()

	m.Check()

	// log.Println(*msg)

	m.Datas = append(m.Datas, &MsgData{Cmd: *msg, StartTime: time.Now().UnixNano()})
}

// 收到消息时调用
func (m *MsgDataManager) StoreRecvMsg(msg *string, sus bool) {
	defer func() {
		m.mutex.Unlock()
	}()
	m.mutex.Lock()

	m.Check()

	// log.Println(*msg, sus)

	now := time.Now().UnixNano()
	for _, it := range m.Datas {
		if it.Fn {
			continue
		}

		if it.Cmd != *msg {
			continue
		}

		it.EndTime = now
		it.Fn = true
		it.Sus = sus
	}
}

// 非线程安全函数
func (m *MsgDataManager) Check() {
	now := time.Now().UnixNano()
	for _, it := range m.Datas {
		if it.Fn {
			continue
		} else {

			if time.Duration(now-it.StartTime) > time.Second*3 {
				it.EndTime = now
				it.Fn = true
				it.Sus = false
			}
		}
	}
}

// 获取已经处理好的消息数据
func (m *MsgDataManager) GetData() []MsgMiniData {
	defer func() {
		m.mutex.Unlock()
	}()
	m.mutex.Lock()

	m.Check()

	var ret []MsgMiniData

	for i := 0; i < len(m.Datas); {
		if m.Datas[i].Fn {
			ret = append(ret, m.Datas[i].GetMiniData())
			m.Datas = append(m.Datas[:i], m.Datas[i+1:]...)
		} else {
			i++
		}
	}

	// for _, it := range ret {
	// 	log.Println(it.Cmd, it.Sus, it.Duration)
	// }

	return ret
}

// 等待返回消息
func (m *MsgDataManager) NeedWaitServerMsg() (string, bool) {
	defer func() {
		m.mutex.Unlock()
	}()
	m.mutex.Lock()

	m.Check()

	for _, it := range m.Datas {
		if !it.Fn {
			return it.Cmd, true
		}
	}

	return "", false
}
