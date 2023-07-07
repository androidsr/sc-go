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
	sSocket.clientList = make([]*websocket.Conn, 0)
	Socket = sSocket
	return sSocket
}

type Message struct {
	UserId string
	Data   []byte
}

type Wsocket struct {
	upgrader   websocket.Upgrader
	clients    map[string]*websocket.Conn
	clientList []*websocket.Conn
	user       UserInfo
	Data       chan Message
	PingFail   int
	Interval   time.Duration
}

// 获取在线客户端数量
func (m *Wsocket) GetSize() int {
	return len(m.clientList) + len(m.clients)
}

// 绑定客户端
func (m *Wsocket) Register(w http.ResponseWriter, r *http.Request) error {
	client, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	var userId string
	if m.user != nil {
		userId = m.user(w, r)
		m.clients[userId] = client
	} else {
		m.clientList = append(m.clientList, client)
	}
	go m.handler(userId, client)
	return nil
}

func (m *Wsocket) removeElement(client *websocket.Conn) {
	result := make([]*websocket.Conn, 0)
	for _, item := range m.clientList {
		if item != client {
			result = append(result, item)
		}
	}
	m.clientList = result
}

func (m *Wsocket) handler(userId string, client *websocket.Conn) {
	defer func() {
		recover()
	}()
	defer func() {
		if userId != "" {
			delete(m.clients, userId)
		} else {
			m.removeElement(client)
		}
		client.Close()
	}()
	maxNotPing := 0
	isRun := true
	go func() {
		ticker := time.NewTicker(m.Interval)
		for isRun {
			<-ticker.C
			maxNotPing++
			if maxNotPing > m.PingFail {
				log.Println("客户端ping超时...")
				ticker.Stop()
				client.Close()
				isRun = false
			}
		}
	}()
	for isRun {
		_, message, err := client.ReadMessage()
		if err != nil {
			maxNotPing = 1000
			isRun = false
			break
		}
		maxNotPing = 0
		if string(message) != "ping" {
			m.Data <- Message{UserId: userId, Data: message}
		} else {
			client.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (m *Wsocket) WriteString(key string, messageType int, message string) {
	m.Write(key, messageType, []byte(message))
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
		for _, client := range m.clientList {
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
		for _, v := range m.clientList {
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
