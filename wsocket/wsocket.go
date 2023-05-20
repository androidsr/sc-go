package wsocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	Socket *Wsocket
)

// 定义回调接口获取用户信息
type UserInfo func(w http.ResponseWriter, r *http.Request) string

func New(upgrader websocket.Upgrader, interval time.Duration, pingFail int, user UserInfo) *Wsocket {
	sSocket := new(Wsocket)
	sSocket.clients = make(map[string]*websocket.Conn, 0)
	sSocket.Data = make(chan Message, 100)
	sSocket.upgrader = upgrader
	sSocket.user = user
	sSocket.Interval = interval
	sSocket.PingFail = pingFail
	Socket = sSocket
	return sSocket
}

type Message struct {
	UserId string
	Data   []byte
}

type Wsocket struct {
	upgrader websocket.Upgrader
	clients  map[string]*websocket.Conn
	user     UserInfo
	Data     chan Message
	PingFail int
	Interval time.Duration
}

// 绑定客户端
func (m *Wsocket) Register(w http.ResponseWriter, r *http.Request) error {
	client, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	userId := m.user(w, r)
	m.clients[userId] = client
	go m.handler(userId, client)
	return nil
}

func (m *Wsocket) handler(userId string, client *websocket.Conn) {
	defer func() {
		delete(m.clients, userId)
		client.Close()
	}()
	maxNotPing := 0
	isRun := true
	go func() {
		ticker := time.NewTicker(m.Interval)
		for {
			select {
			case <-ticker.C:
				maxNotPing++
			default:
				if maxNotPing > m.PingFail {
					client.Close()
					isRun = false
				}
			}
		}
	}()
	for isRun {
		_, message, err := client.ReadMessage()
		if err != nil {
			log.Println("读取WebSocket消息时发生错误：", err)
			continue
		}
		maxNotPing = 0
		if string(message) != "ping" {
			m.Data <- Message{UserId: userId, Data: message}
		}
	}
}

func (m *Wsocket) Write(key string, messageType int, message []byte) {
	if key == "" {
		for _, client := range m.clients {
			err := client.WriteMessage(messageType, message)
			if err != nil {
				delete(m.clients, key)
				client.Close()
				continue
			}
		}
	} else {
		client := m.clients[key]
		err := client.WriteMessage(messageType, message)
		if err != nil {
			delete(m.clients, key)
			client.Close()
		}
	}
}

// 关闭客户端 key为空全部关闭
func (m *Wsocket) Close(key string) {
	if key == "" {
		for _, v := range m.clients {
			v.Close()
		}
	} else {
		client := m.clients[key]
		if client != nil {
			client.Close()
		}
		delete(m.clients, key)
	}
}
