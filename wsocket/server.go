package wsocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	WsServer *WebSocketServer
)

type Session struct {
	Id string `json:"id"`
}

type Client struct {
	mu    sync.Mutex
	Conn  *websocket.Conn
	ID    string
	Alive bool
}

type WebSocketServer struct {
	clients  map[string]*Client
	upgrader *websocket.Upgrader
	engine   *nbhttp.Engine
	r        *gin.Engine
}

func NewServer(r *gin.Engine) *WebSocketServer {
	WsServer = &WebSocketServer{
		clients:  make(map[string]*Client),
		upgrader: websocket.NewUpgrader(),
		r:        r,
	}
	WsServer.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	WsServer.upgrader.OnClose(WsServer.handleClose)
	return WsServer
}

func (s *WebSocketServer) OnMessage(handleMessage func(conn *websocket.Conn, messageType websocket.MessageType, data []byte)) {
	s.upgrader.OnMessage(handleMessage)
}

func (s *WebSocketServer) Start(addrs string, maxLoad int) error {
	s.r.GET("/ws", s.HandleWebSocket)
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{addrs},
		MaxLoad:                 maxLoad,
		ReleaseWebsocketPayload: true,
		Handler:                 s.r,
	})
	go s.HeartbeatCheck()
	err := engine.Start()
	if err != nil {
		return err
	}
	return nil
}

func (s *WebSocketServer) GetSize() int {
	return len(s.clients)
}
func (s *WebSocketServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	s.engine.Shutdown(ctx)
}

func (s *WebSocketServer) handleClose(conn *websocket.Conn, err error) {
	session := conn.Session().(*Session)
	clientID := session.Id
	client := s.clients[clientID]
	client.mu.Lock()
	conn.Close()
	delete(s.clients, clientID)
	client.mu.Unlock()
}

func (s *WebSocketServer) HeartbeatCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for id, client := range s.clients {
			client.mu.Lock()
			if !client.Alive {
				client.Conn.Close()
				delete(s.clients, id)
				fmt.Printf("移除失效的客户端：%s\n", id)
			} else {
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					client.Alive = false
					client.Conn.Close()
					delete(s.clients, id)
					fmt.Printf("移除失效的客户端: %s\n", id)
				}
			}
			client.mu.Unlock()
		}
	}
}

func (s *WebSocketServer) SendToClient(clientID string, message []byte) {
	if client, ok := s.clients[clientID]; ok {
		client.mu.Lock()
		client.Conn.WriteMessage(websocket.TextMessage, message)
		client.mu.Unlock()
	} else {
		fmt.Printf("未找到指定的客户端: %s \n", clientID)
	}
}

func (s *WebSocketServer) HandleWebSocket(c *gin.Context) {
	id := c.Query("id") // 从请求的参数中获取用户ID
	if id == "" {
		c.JSON(400, gin.H{"error": "Missing user_id"})
		return
	}
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	conn.SetSession(&Session{Id: id})
	client := &Client{
		Conn:  conn,
		ID:    id,
		Alive: true,
	}
	s.clients[id] = client
	fmt.Println("连接成功...")
}
