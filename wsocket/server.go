package wsocket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
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
	clients   map[string]*Client
	clientsMu sync.RWMutex

	upgrader *websocket.Upgrader
	engine   *nbhttp.Engine
	r        *gin.Engine

	// 用于优雅停止心跳
	cancel context.CancelFunc

	// 用户的消息处理函数（由 OnMessage 设置）
	msgHandler   func(conn *websocket.Conn, messageType websocket.MessageType, data []byte)
	msgHandlerMu sync.RWMutex
}

// NewServer 初始化 server：绑定 OnClose 与 内部 OnMessage wrapper（负责刷新 Alive 并转发到用户 handler）
func NewServer(r *gin.Engine) *WebSocketServer {
	s := &WebSocketServer{
		clients:  make(map[string]*Client),
		upgrader: websocket.NewUpgrader(),
		r:        r,
	}

	// 允许跨域（按需修改）
	s.upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// 关闭事件：在连接真正对应 map 中的 client 时才删除
	s.upgrader.OnClose(s.handleClose)

	// 内部 OnMessage wrapper：用于刷新 Alive，并把消息转发给用户 handler（如果有）
	s.upgrader.OnMessage(func(conn *websocket.Conn, mt websocket.MessageType, data []byte) {
		// 刷新对应 client 的 Alive
		if sess, ok := conn.Session().(*Session); ok {
			s.clientsMu.RLock()
			if cli, exists := s.clients[sess.Id]; exists {
				cli.mu.Lock()
				cli.Alive = true
				cli.mu.Unlock()
			}
			s.clientsMu.RUnlock()
		}
		// 转发到用户 handler（线程安全读）
		s.msgHandlerMu.RLock()
		h := s.msgHandler
		s.msgHandlerMu.RUnlock()
		if h != nil {
			h(conn, mt, data)
		}
	})

	return s
}

// OnMessage: 用户注册消息处理器（会被内部 wrapper 调用）
func (s *WebSocketServer) OnMessage(
	handleMessage func(conn *websocket.Conn, messageType websocket.MessageType, data []byte),
) {
	s.msgHandlerMu.Lock()
	s.msgHandler = handleMessage
	s.msgHandlerMu.Unlock()
}

// Start: 启动 nbio 引擎并启动心跳 goroutine
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

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.heartbeat(ctx)

	return engine.Start()
}

func (s *WebSocketServer) GetSize() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// Close: 优雅关闭（停止心跳、关闭所有连接、shutdown engine）
func (s *WebSocketServer) Close() {
	if s.cancel != nil {
		s.cancel()
	}
	s.clientsMu.Lock()
	for _, c := range s.clients {
		_ = c.Conn.Close()
	}
	s.clients = make(map[string]*Client)
	s.clientsMu.Unlock()

	if s.engine != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.engine.Shutdown(ctx)
	}
}

// handleClose: nbio 在连接关闭时会回调。仅在 map 中该 id 的连接仍然指向这个 conn 时才删除（防止误删新连接）
func (s *WebSocketServer) handleClose(conn *websocket.Conn, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("handleClose panic: %v", r)
		}
	}()

	if sess, ok := conn.Session().(*Session); ok {
		s.clientsMu.Lock()
		if client, exists := s.clients[sess.Id]; exists {
			// 只有当 map 中的 client 正是这个 conn 时才删
			client.mu.Lock()
			if client.Conn == conn {
				delete(s.clients, sess.Id)
				client.mu.Unlock()
				s.clientsMu.Unlock()
				log.Printf("client %s disconnected (handleClose)", sess.Id)
				return
			}
			client.mu.Unlock()
		}
		s.clientsMu.Unlock()
	}
}

// heartbeat: 不依赖 OnPong；策略：如果 client.Alive 为 false（上轮未被刷新），则移除；否则将 Alive 置 false 并发送 Ping，写失败也移除
func (s *WebSocketServer) heartbeat(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("heartbeat panic: %v", r)
		}
	}()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var toRemove []string

			s.clientsMu.RLock()
			for id, c := range s.clients {
				c.mu.Lock()
				if !c.Alive {
					// 上一轮没有被刷新，标记为待移除
					toRemove = append(toRemove, id)
					c.mu.Unlock()
					continue
				}
				// 先把 Alive 置为 false，等待下一轮有消息来刷新
				c.Alive = false
				// 发送 ping 检测写是否正常（若写失败，也加入待移除）
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					toRemove = append(toRemove, id)
				}
				c.mu.Unlock()
			}
			s.clientsMu.RUnlock()

			if len(toRemove) > 0 {
				s.clientsMu.Lock()
				for _, id := range toRemove {
					if cli, ok := s.clients[id]; ok {
						_ = cli.Conn.Close()
						delete(s.clients, id)
						log.Printf("removed inactive client %s", id)
					}
				}
				s.clientsMu.Unlock()
			}
		}
	}
}

// SendToClient: 发送消息给指定 id（线程安全）
func (s *WebSocketServer) SendToClient(id string, msg []byte) {
	s.clientsMu.RLock()
	c, ok := s.clients[id]
	s.clientsMu.RUnlock()
	if !ok {
		log.Printf("client %s not found", id)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.Conn.WriteMessage(websocket.TextMessage, msg)
}

// HandleWebSocket: 升级为 websocket，并**安全替换**同 id 的旧连接（先替换 map，再在 map 锁释放后关闭旧连接）
func (s *WebSocketServer) HandleWebSocket(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("HandleWebSocket panic: %v", r)
		}
	}()

	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id"})
		return
	}
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	conn.SetSession(&Session{Id: id})

	newClient := &Client{
		Conn:  conn,
		ID:    id,
		Alive: true,
	}

	// 注意：不要在持有 clientsMu 的时候调用 old.Conn.Close()（可能导致与 handleClose 的锁争用死锁）
	var oldConn *websocket.Conn
	s.clientsMu.Lock()
	if old, ok := s.clients[id]; ok {
		oldConn = old.Conn
	}
	s.clients[id] = newClient
	s.clientsMu.Unlock()

	// 在锁外关闭旧连接（如果有）
	if oldConn != nil {
		go func(cn *websocket.Conn, id string) {
			_ = cn.Close()
			log.Printf("old connection for %s closed due to reconnect", id)
		}(oldConn, id)
	}

	log.Printf("client %s connected", id)
}
