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
	mu       sync.Mutex
	Conn     *websocket.Conn
	ID       string
	LastPong time.Time // 记录最近一次收到 pong 的时间
}

type WebSocketServer struct {
	clients       map[string]*Client
	clientsMu     sync.RWMutex
	upgrader      *websocket.Upgrader
	engine        *nbhttp.Engine
	r             *gin.Engine
	checkInterval time.Duration
	maxIdle       time.Duration
}

func NewServer(r *gin.Engine, checkInterval, maxIdle time.Duration) *WebSocketServer {
	s := &WebSocketServer{
		clients:       make(map[string]*Client),
		upgrader:      websocket.NewUpgrader(),
		r:             r,
		checkInterval: checkInterval,
		maxIdle:       maxIdle,
	}
	WsServer = s

	// 允许跨域
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// 连接关闭时回调
	s.upgrader.OnClose(s.handleClose)
	// 消息处理：处理 pong
	s.upgrader.OnMessage(func(conn *websocket.Conn, mt websocket.MessageType, data []byte) {
		if string(data) == "pong" {
			if sess, ok := conn.Session().(*Session); ok {
				s.clientsMu.RLock()
				if cli, exists := s.clients[sess.Id]; exists {
					cli.mu.Lock()
					cli.LastPong = time.Now()
					cli.mu.Unlock()
				}
				s.clientsMu.RUnlock()
			}
			return
		}
		// 这里可以加其他业务消息处理
	})

	return s
}

func (s *WebSocketServer) Start(addr string, maxLoad int) error {
	s.r.GET("/ws", s.HandleWebSocket)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{addr},
		MaxLoad:                 maxLoad,
		ReleaseWebsocketPayload: true,
		Handler:                 s.r,
	})
	s.engine = engine

	go s.heartbeatCheck()

	return engine.Start()
}

func (s *WebSocketServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if s.engine != nil {
		_ = s.engine.Shutdown(ctx)
	}
}

func (s *WebSocketServer) GetSize() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

func (s *WebSocketServer) handleClose(conn *websocket.Conn, err error) {
	if sess, ok := conn.Session().(*Session); ok {
		s.clientsMu.Lock()
		if client, exists := s.clients[sess.Id]; exists && client.Conn == conn {
			delete(s.clients, sess.Id)
			fmt.Printf("客户端[%s]已关闭连接\n", sess.Id)
		}
		s.clientsMu.Unlock()
	}
}

func (s *WebSocketServer) heartbeatCheck() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.clientsMu.RLock()
		for id, cli := range s.clients {
			cli.mu.Lock()
			if time.Since(cli.LastPong) > s.maxIdle {
				// 超过最大空闲时间，关闭连接
				cli.Conn.Close()
				cli.mu.Unlock()
				s.clientsMu.RUnlock()
				s.clientsMu.Lock()
				delete(s.clients, id)
				s.clientsMu.Unlock()
				s.clientsMu.RLock()
				fmt.Printf("客户端[%s]心跳超时，已断开\n", id)
			} else {
				// 发送 ping
				_ = cli.Conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				cli.mu.Unlock()
			}
		}
		s.clientsMu.RUnlock()
	}
}

func (s *WebSocketServer) SendToClient(clientID string, message []byte) {
	s.clientsMu.RLock()
	client, ok := s.clients[clientID]
	s.clientsMu.RUnlock()
	if !ok {
		fmt.Printf("未找到指定的客户端: %s\n", clientID)
		return
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	_ = client.Conn.WriteMessage(websocket.TextMessage, message)
}

func (s *WebSocketServer) HandleWebSocket(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Missing id"})
		return
	}
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	conn.SetSession(&Session{Id: id})

	client := &Client{
		Conn:     conn,
		ID:       id,
		LastPong: time.Now(), // 初始为当前时间
	}

	s.clientsMu.Lock()
	// 如果已有旧连接，先关闭它
	if old, ok := s.clients[id]; ok {
		old.Conn.Close()
	}
	s.clients[id] = client
	s.clientsMu.Unlock()

	fmt.Printf("客户端[%s]连接成功\n", id)
}
