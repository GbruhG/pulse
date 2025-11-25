package backend

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	connectionIDCounter uint64
	messageIDCounter    uint64
)

// SSEManager handles Server-Sent Events connections
type SSEManager struct {
	app         AppInterface
	connections map[string]*SSEConnection
	mu          sync.RWMutex
}

// SSEConnection holds an active SSE connection
type SSEConnection struct {
	ID              string
	URL             string
	Client          *http.Client
	Response        *http.Response
	Context         context.Context
	Cancel          context.CancelFunc
	AutoReconnect   bool
	RetryTimeout    int // milliseconds
	LastEventID     string
	WithCredentials bool
	Headers         map[string]string
	EventTypeFilter []string
	reconnectCount  int
	maxReconnects   int
}

// SSEConnectRequest holds SSE connection parameters
type SSEConnectRequest struct {
	URL             string            `json:"url"`
	WithCredentials bool              `json:"withCredentials"`
	RetryTimeout    int               `json:"retryTimeout"`
	LastEventID     string            `json:"lastEventId"`
	AutoReconnect   bool              `json:"autoReconnect"`
	Headers         map[string]string `json:"customHeaders"`
	EventTypeFilter []string          `json:"eventTypeFilter"`
}

// NewSSEManager creates an SSE manager
func NewSSEManager(app AppInterface) *SSEManager {
	return &SSEManager{
		app:         app,
		connections: make(map[string]*SSEConnection),
	}
}

// getNextConnectionID generates a unique connection ID
func getNextConnectionID() string {
	id := atomic.AddUint64(&connectionIDCounter, 1)
	return fmt.Sprintf("sse-%d", id)
}

// getNextMessageID generates a unique message ID
func getNextMessageID() string {
	id := atomic.AddUint64(&messageIDCounter, 1)
	return fmt.Sprintf("msg-%d", id)
}

// Connect establishes an SSE connection
func (s *SSEManager) Connect(req SSEConnectRequest) (string, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 0, // No timeout for streaming
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // TODO: Make configurable
		},
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("GET", req.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	// Add custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add Last-Event-ID if provided
	if req.LastEventID != "" {
		httpReq.Header.Set("Last-Event-ID", req.LastEventID)
	}

	// Make request
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		resp.Body.Close()
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	// Create connection object
	connID := getNextConnectionID()
	ctx, cancel := context.WithCancel(context.Background())

	sseConn := &SSEConnection{
		ID:              connID,
		URL:             req.URL,
		Client:          client,
		Response:        resp,
		Context:         ctx,
		Cancel:          cancel,
		AutoReconnect:   req.AutoReconnect,
		RetryTimeout:    req.RetryTimeout,
		LastEventID:     req.LastEventID,
		WithCredentials: req.WithCredentials,
		Headers:         req.Headers,
		EventTypeFilter: req.EventTypeFilter,
		maxReconnects:   10,
	}

	s.mu.Lock()
	s.connections[connID] = sseConn
	s.mu.Unlock()

	// Start reading events
	go s.readEvents(sseConn)

	// Send system message
	s.emitMessage(StreamMessage{
		ID:        getNextMessageID(),
		Direction: "system",
		Protocol:  "SSE",
		Payload:   fmt.Sprintf("Connected to %s", req.URL),
		Timestamp: time.Now(),
	})

	return connID, nil
}

// Disconnect closes an SSE connection
func (s *SSEManager) Disconnect(connectionID string) error {
	s.mu.Lock()
	conn, ok := s.connections[connectionID]
	if ok {
		delete(s.connections, connectionID)
	}
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("connection not found")
	}

	// Cancel context to stop reader
	conn.Cancel()

	// Close response body
	if conn.Response != nil && conn.Response.Body != nil {
		conn.Response.Body.Close()
	}

	s.emitMessage(StreamMessage{
		ID:        getNextMessageID(),
		Direction: "system",
		Protocol:  "SSE",
		Payload:   "Disconnected",
		Timestamp: time.Now(),
	})

	return nil
}

// readEvents reads SSE events
func (s *SSEManager) readEvents(conn *SSEConnection) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[SSE] Reader panic: %v\n", r)
			s.emitMessage(StreamMessage{
				ID:        getNextMessageID(),
				Direction: "error",
				Protocol:  "SSE",
				Payload:   fmt.Sprintf("Reader panic: %v", r),
				Timestamp: time.Now(),
			})
		}
	}()

	scanner := bufio.NewScanner(conn.Response.Body)

	var (
		eventType = "message" // default event type
		data      string
		id        string
		retry     int
	)

	for {
		select {
		case <-conn.Context.Done():
			fmt.Println("[SSE] Context done, stopping reader")
			return
		default:
			if !scanner.Scan() {
				// Connection closed or error
				if err := scanner.Err(); err != nil {
					fmt.Printf("[SSE] Scanner error: %v\n", err)
					s.emitMessage(StreamMessage{
						ID:        getNextMessageID(),
						Direction: "error",
						Protocol:  "SSE",
						Payload:   fmt.Sprintf("Connection error: %v", err),
						Timestamp: time.Now(),
					})
				} else {
					s.emitMessage(StreamMessage{
						ID:        getNextMessageID(),
						Direction: "system",
						Protocol:  "SSE",
						Payload:   "Connection closed by server",
						Timestamp: time.Now(),
					})
				}

				// Attempt reconnection if enabled
				if conn.AutoReconnect && conn.reconnectCount < conn.maxReconnects {
					go s.attemptReconnect(conn)
				}
				return
			}

			line := scanner.Text()

			// Empty line means end of event
			if line == "" {
				if data != "" {
					// Check if we should filter this event type
					shouldEmit := len(conn.EventTypeFilter) == 0
					if !shouldEmit {
						for _, filter := range conn.EventTypeFilter {
							if filter == eventType {
								shouldEmit = true
								break
							}
						}
					}

					if shouldEmit {
						// Format the event data
						payload := data
						if eventType != "message" {
							payload = fmt.Sprintf("[Event: %s] %s", eventType, data)
						}

						s.emitMessage(StreamMessage{
							ID:        getNextMessageID(),
							Direction: "inbound",
							Protocol:  "SSE",
							Payload:   payload,
							Timestamp: time.Now(),
						})

						// Update last event ID
						if id != "" {
							conn.LastEventID = id
						}
					}
				}

				// Reset for next event
				eventType = "message"
				data = ""
				id = ""
				continue
			}

			// Parse SSE field
			if len(line) > 0 && line[0] == ':' {
				// Comment line, ignore
				continue
			}

			colonIndex := -1
			for i, c := range line {
				if c == ':' {
					colonIndex = i
					break
				}
			}

			if colonIndex == -1 {
				// Field with no value
				continue
			}

			field := line[:colonIndex]
			value := ""
			if colonIndex+1 < len(line) {
				value = line[colonIndex+1:]
				// Remove leading space if present
				if len(value) > 0 && value[0] == ' ' {
					value = value[1:]
				}
			}

			switch field {
			case "event":
				eventType = value
			case "data":
				if data != "" {
					data += "\n"
				}
				data += value
			case "id":
				id = value
			case "retry":
				// Parse retry value (milliseconds)
				fmt.Sscanf(value, "%d", &retry)
				if retry > 0 {
					conn.RetryTimeout = retry
				}
			}
		}
	}
}

// attemptReconnect tries to reconnect to the SSE server
func (s *SSEManager) attemptReconnect(conn *SSEConnection) {
	conn.reconnectCount++

	s.emitMessage(StreamMessage{
		ID:        getNextMessageID(),
		Direction: "system",
		Protocol:  "SSE",
		Payload:   fmt.Sprintf("Reconnecting... (attempt %d/%d)", conn.reconnectCount, conn.maxReconnects),
		Timestamp: time.Now(),
	})

	// Wait before reconnecting
	select {
	case <-conn.Context.Done():
		fmt.Println("[SSE] Context cancelled during reconnect delay")
		return
	case <-time.After(time.Duration(conn.RetryTimeout) * time.Millisecond):
		// Continue with reconnect
	}

	// Close old response
	if conn.Response != nil && conn.Response.Body != nil {
		conn.Response.Body.Close()
	}

	// Create new request
	httpReq, err := http.NewRequest("GET", conn.URL, nil)
	if err != nil {
		s.emitMessage(StreamMessage{
			ID:        getNextMessageID(),
			Direction: "error",
			Protocol:  "SSE",
			Payload:   fmt.Sprintf("Reconnection failed: %v", err),
			Timestamp: time.Now(),
		})

		if conn.reconnectCount < conn.maxReconnects {
			go s.attemptReconnect(conn)
		}
		return
	}

	// Add headers
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	for key, value := range conn.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add Last-Event-ID for resumption
	if conn.LastEventID != "" {
		httpReq.Header.Set("Last-Event-ID", conn.LastEventID)
	}

	// Reconnect
	resp, err := conn.Client.Do(httpReq)
	if err != nil {
		s.emitMessage(StreamMessage{
			ID:        getNextMessageID(),
			Direction: "error",
			Protocol:  "SSE",
			Payload:   fmt.Sprintf("Reconnection failed: %v", err),
			Timestamp: time.Now(),
		})

		if conn.reconnectCount < conn.maxReconnects {
			go s.attemptReconnect(conn)
		}
		return
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		s.emitMessage(StreamMessage{
			ID:        getNextMessageID(),
			Direction: "error",
			Protocol:  "SSE",
			Payload:   fmt.Sprintf("Reconnection failed: status %d", resp.StatusCode),
			Timestamp: time.Now(),
		})

		if conn.reconnectCount < conn.maxReconnects {
			go s.attemptReconnect(conn)
		}
		return
	}

	// Update connection
	conn.Response = resp
	conn.reconnectCount = 0

	s.emitMessage(StreamMessage{
		ID:        getNextMessageID(),
		Direction: "system",
		Protocol:  "SSE",
		Payload:   "Reconnected successfully",
		Timestamp: time.Now(),
	})

	// Restart reader
	go s.readEvents(conn)
}

// emitMessage sends a stream message to the frontend
func (s *SSEManager) emitMessage(msg StreamMessage) {
	if s.app == nil || s.app.GetCtx() == nil {
		fmt.Printf("[SSE] ⚠️  Cannot emit message - app context not initialized\n")
		return
	}

	fmt.Printf("[SSE] 🚀 Emitting message to frontend: %+v\n", msg)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[SSE] Event emit panic recovered: %v\n", r)
			}
		}()
		runtime.EventsEmit(s.app.GetCtx(), "stream-message", msg)
		fmt.Println("[SSE] ✅ Message emitted successfully")
	}()
}
