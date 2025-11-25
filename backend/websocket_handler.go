package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WebSocketManager handles WebSocket connections
type WebSocketManager struct {
	app         AppInterface
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
	msgCounter  uint64 // Atomic counter for unique message IDs
}

// WebSocketConnection holds an active WebSocket connection
type WebSocketConnection struct {
	ID             string
	URL            string
	Conn           *websocket.Conn
	Context        context.Context
	Cancel         context.CancelFunc
	AutoReconnect  bool
	ReconnectDelay int // milliseconds
	PingEnabled    bool
	PingInterval   int // milliseconds
	Subprotocol    string
	Headers        map[string]string
	pingTicker     *time.Ticker
	reconnectCount int
	maxReconnects  int
}

// WebSocketConnectRequest holds connection parameters
type WebSocketConnectRequest struct {
	URL            string            `json:"url"`
	Subprotocol    string            `json:"subprotocol"`
	AutoReconnect  bool              `json:"autoReconnect"`
	ReconnectDelay int               `json:"reconnectInterval"` // milliseconds
	PingEnabled    bool              `json:"enablePingPong"`
	PingInterval   int               `json:"pingInterval"` // milliseconds
	Headers        map[string]string `json:"customHeaders"`
}

// WebSocketSendRequest holds a message to send
type WebSocketSendRequest struct {
	ConnectionID string `json:"connectionId"`
	Message      string `json:"message"`
	MessageType  string `json:"messageType"` // "text", "json", "binary"
}

// NewWebSocketManager creates a WebSocket manager
func NewWebSocketManager(app AppInterface) *WebSocketManager {
	return &WebSocketManager{
		app:         app,
		connections: make(map[string]*WebSocketConnection),
		msgCounter:  0,
	}
}

// generateMessageID generates a unique message ID
func (w *WebSocketManager) generateMessageID() string {
	count := atomic.AddUint64(&w.msgCounter, 1)
	return fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), count)
}

// Connect establishes a WebSocket connection
func (w *WebSocketManager) Connect(req WebSocketConnectRequest) (string, error) {
	// Prepare headers
	headers := http.Header{}
	for key, value := range req.Headers {
		headers.Add(key, value)
	}

	// Add subprotocol if specified
	var subprotocols []string
	if req.Subprotocol != "" {
		subprotocols = append(subprotocols, req.Subprotocol)
	}

	// Create dialer with custom config
	dialer := websocket.Dialer{
		Subprotocols:     subprotocols,
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true}, // TODO: Make this configurable
	}

	// Connect
	conn, _, err := dialer.Dial(req.URL, headers)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// Create connection object
	connID := fmt.Sprintf("ws-%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())

	wsConn := &WebSocketConnection{
		ID:             connID,
		URL:            req.URL,
		Conn:           conn,
		Context:        ctx,
		Cancel:         cancel,
		AutoReconnect:  req.AutoReconnect,
		ReconnectDelay: req.ReconnectDelay,
		PingEnabled:    req.PingEnabled,
		PingInterval:   req.PingInterval,
		Subprotocol:    req.Subprotocol,
		Headers:        req.Headers,
		maxReconnects:  10, // Maximum reconnection attempts
	}

	w.mu.Lock()
	w.connections[connID] = wsConn
	w.mu.Unlock()

	// Start message reader goroutine
	go w.readMessages(wsConn)

	// Start ping goroutine if enabled
	if req.PingEnabled && req.PingInterval > 0 {
		go w.sendPings(wsConn)
	}

	// Send system message
	w.emitMessage(StreamMessage{
		ID:        w.generateMessageID(),
		Direction: "system",
		Protocol:  "WebSocket",
		Payload:   fmt.Sprintf("Connected to %s", req.URL),
		Timestamp: time.Now(),
	})

	return connID, nil
}

// SendMessage sends a message through WebSocket
func (w *WebSocketManager) SendMessage(req WebSocketSendRequest) error {
	w.mu.RLock()
	conn, ok := w.connections[req.ConnectionID]
	w.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found: %s", req.ConnectionID)
	}

	var err error
	var messageType int

	switch req.MessageType {
	case "text", "json":
		messageType = websocket.TextMessage
		err = conn.Conn.WriteMessage(messageType, []byte(req.Message))
	case "binary":
		messageType = websocket.BinaryMessage
		// TODO: Decode base64 or hex string to bytes
		err = conn.Conn.WriteMessage(messageType, []byte(req.Message))
	default:
		messageType = websocket.TextMessage
		err = conn.Conn.WriteMessage(messageType, []byte(req.Message))
	}

	if err != nil {
		w.emitMessage(StreamMessage{
			ID:        w.generateMessageID(),
			Direction: "error",
			Protocol:  "WebSocket",
			Payload:   fmt.Sprintf("Failed to send: %s", err.Error()),
			Timestamp: time.Now(),
		})
		return err
	}

	// Emit outbound message
	w.emitMessage(StreamMessage{
		ID:        w.generateMessageID(),
		Direction: "outbound",
		Protocol:  "WebSocket",
		Payload:   req.Message,
		Timestamp: time.Now(),
	})

	return nil
}

// Disconnect closes a WebSocket connection
func (w *WebSocketManager) Disconnect(connectionID string) error {
	w.mu.Lock()
	conn, ok := w.connections[connectionID]
	if ok {
		delete(w.connections, connectionID)
	}
	w.mu.Unlock()

	if !ok {
		return fmt.Errorf("connection not found")
	}

	// CRITICAL: Cancel context FIRST to stop ALL goroutines immediately
	conn.Cancel()

	// Stop ping ticker if running (non-blocking)
	if conn.pingTicker != nil {
		conn.pingTicker.Stop()
	}

	// Force close immediately in goroutine - don't wait for graceful close
	go func() {
		// Try graceful close with very short timeout
		closeDone := make(chan bool, 1)
		go func() {
			conn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			closeDone <- true
		}()

		select {
		case <-closeDone:
			// Graceful close succeeded
		case <-time.After(100 * time.Millisecond):
			// Timeout - force close immediately
			fmt.Println("[WS] Graceful close timeout, forcing close")
		}

		// Force close the underlying connection
		conn.Conn.Close()
	}()

	// Send disconnect message immediately without waiting
	w.emitMessage(StreamMessage{
		ID:        w.generateMessageID(),
		Direction: "system",
		Protocol:  "WebSocket",
		Payload:   "Disconnected",
		Timestamp: time.Now(),
	})

	return nil
}

// readMessages reads messages from the WebSocket
func (w *WebSocketManager) readMessages(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[WS] Reader panic: %v\n", r)
			w.emitMessage(StreamMessage{
				ID:        w.generateMessageID(),
				Direction: "error",
				Protocol:  "WebSocket",
				Payload:   fmt.Sprintf("Reader panic: %v", r),
				Timestamp: time.Now(),
			})
		}
	}()

	for {
		select {
		case <-conn.Context.Done():
			fmt.Println("[WS] Context done, stopping reader")
			return
		default:
			messageType, message, err := conn.Conn.ReadMessage()
			if err != nil {
				// Connection closed or error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("[WS] Connection error: %s\n", err.Error())
					w.emitMessage(StreamMessage{
						ID:        w.generateMessageID(),
						Direction: "error",
						Protocol:  "WebSocket",
						Payload:   fmt.Sprintf("Connection error: %s", err.Error()),
						Timestamp: time.Now(),
					})
				} else {
					fmt.Println("[WS] Connection closed")
					w.emitMessage(StreamMessage{
						ID:        w.generateMessageID(),
						Direction: "system",
						Protocol:  "WebSocket",
						Payload:   "Connection closed",
						Timestamp: time.Now(),
					})
				}

				// Check if context was cancelled (user disconnected)
				select {
				case <-conn.Context.Done():
					fmt.Println("[WS] Context cancelled - user disconnected, not reconnecting")
					return
				default:
					// Context not cancelled, attempt reconnection if enabled
					if conn.AutoReconnect && conn.reconnectCount < conn.maxReconnects {
						go w.attemptReconnect(conn)
					}
				}

				return
			}

			// Handle different message types
			var payload string
			switch messageType {
			case websocket.TextMessage:
				payload = string(message)
			case websocket.BinaryMessage:
				payload = fmt.Sprintf("[Binary data: %d bytes]", len(message))
			case websocket.PingMessage:
				payload = "[Ping]"
			case websocket.PongMessage:
				payload = "[Pong]"
			case websocket.CloseMessage:
				payload = "[Close]"
			}

			// Emit inbound message - THE FIX IS HERE
			w.emitMessage(StreamMessage{
				ID:        w.generateMessageID(),
				Direction: "inbound",
				Protocol:  "WebSocket",
				Payload:   payload,
				Timestamp: time.Now(),
			})
		}
	}
}

// sendPings sends periodic ping messages
func (w *WebSocketManager) sendPings(conn *WebSocketConnection) {
	conn.pingTicker = time.NewTicker(time.Duration(conn.PingInterval) * time.Millisecond)
	defer conn.pingTicker.Stop()

	for {
		select {
		case <-conn.Context.Done():
			return
		case <-conn.pingTicker.C:
			err := conn.Conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				w.emitMessage(StreamMessage{
					ID:        w.generateMessageID(),
					Direction: "error",
					Protocol:  "WebSocket",
					Payload:   fmt.Sprintf("Ping failed: %s", err.Error()),
					Timestamp: time.Now(),
				})
				return
			}
		}
	}
}

// attemptReconnect tries to reconnect to the WebSocket
func (w *WebSocketManager) attemptReconnect(conn *WebSocketConnection) {
	// Check if context was cancelled before attempting reconnect
	select {
	case <-conn.Context.Done():
		fmt.Println("[WS] Context cancelled - aborting reconnect attempt")
		return
	default:
	}

	conn.reconnectCount++

	w.emitMessage(StreamMessage{
		ID:        w.generateMessageID(),
		Direction: "system",
		Protocol:  "WebSocket",
		Payload:   fmt.Sprintf("Reconnecting... (attempt %d/%d)", conn.reconnectCount, conn.maxReconnects),
		Timestamp: time.Now(),
	})

	// Wait before reconnecting (but check for cancellation)
	select {
	case <-conn.Context.Done():
		fmt.Println("[WS] Context cancelled during reconnect delay")
		return
	case <-time.After(time.Duration(conn.ReconnectDelay) * time.Millisecond):
		// Continue with reconnect
	}

	// Prepare headers
	headers := http.Header{}
	for key, value := range conn.Headers {
		headers.Add(key, value)
	}

	// Add subprotocol if specified
	var subprotocols []string
	if conn.Subprotocol != "" {
		subprotocols = append(subprotocols, conn.Subprotocol)
	}

	// Create dialer
	dialer := websocket.Dialer{
		Subprotocols:     subprotocols,
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
	}

	// Reconnect
	newConn, _, err := dialer.Dial(conn.URL, headers)
	if err != nil {
		w.emitMessage(StreamMessage{
			ID:        w.generateMessageID(),
			Direction: "error",
			Protocol:  "WebSocket",
			Payload:   fmt.Sprintf("Reconnection failed: %s", err.Error()),
			Timestamp: time.Now(),
		})

		// Try again if we haven't hit max attempts
		if conn.reconnectCount < conn.maxReconnects {
			go w.attemptReconnect(conn)
		}
		return
	}

	// Update connection
	conn.Conn = newConn
	conn.reconnectCount = 0 // Reset counter on successful reconnection

	w.emitMessage(StreamMessage{
		ID:        w.generateMessageID(),
		Direction: "system",
		Protocol:  "WebSocket",
		Payload:   "Reconnected successfully",
		Timestamp: time.Now(),
	})

	// Restart reader
	go w.readMessages(conn)

	// Restart pings if enabled
	if conn.PingEnabled && conn.PingInterval > 0 {
		go w.sendPings(conn)
	}
}

// emitMessage sends a stream message to the frontend
func (w *WebSocketManager) emitMessage(msg StreamMessage) {
	// CRITICAL FIX: Check if context is valid before emitting
	if w.app == nil || w.app.GetCtx() == nil {
		fmt.Printf("[WS] ⚠️  Cannot emit message - app context not initialized yet\n")
		return
	}

	// Emit in a separate goroutine to prevent blocking
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[WS] Event emit panic recovered: %v\n", r)
			}
		}()
		runtime.EventsEmit(w.app.GetCtx(), "stream-message", msg)
	}()
}
