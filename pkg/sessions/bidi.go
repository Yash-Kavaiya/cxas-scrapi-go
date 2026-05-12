package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
)

const (
	wsBaseURL       = "wss://ces.googleapis.com/ws/google.cloud.ces.v1.SessionService/BidiRunSession"
	audioChunkSize  = 3200
	audioSampleRate = 16000
	chunkDelay      = 100 * time.Millisecond
)

// BidiSession manages a WebSocket bidirectional session for streaming audio.
// Use NewBidiSession to create a session, SendText/SendAudio to send input,
// Outputs to retrieve received messages, and Close to terminate.
type BidiSession struct {
	cfg     BidiConfig
	conn    *websocket.Conn
	mu      sync.Mutex
	outputs []map[string]interface{}
	done    chan struct{}
	sendErr error
	recvErr error
}

// agentTurnManager tracks how many audio bytes we've received in the current agent turn.
type agentTurnManager struct {
	mu            sync.Mutex
	bytesReceived int
	turnCompleted bool
}

func (m *agentTurnManager) add(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bytesReceived += n
}

func (m *agentTurnManager) complete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.turnCompleted = true
}

func (m *agentTurnManager) isDone() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.turnCompleted
}

// NewBidiSession opens a WebSocket BidiSession connection.
func NewBidiSession(ctx context.Context, cfg BidiConfig, authCfg auth.Config) (*BidiSession, error) {
	ts, err := auth.NewTokenSource(ctx, authCfg)
	if err != nil {
		return nil, err
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}

	wsURL := fmt.Sprintf("%s/locations/%s", wsBaseURL, cfg.Location)
	header := map[string][]string{
		"Authorization": {"Bearer " + tok.AccessToken},
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	s := &BidiSession{cfg: cfg, conn: conn, done: make(chan struct{})}
	if err := s.sendSessionConfig(); err != nil {
		conn.Close()
		return nil, err
	}

	go s.receiveLoop()
	return s, nil
}

func (s *BidiSession) sendSessionConfig() error {
	rate := s.cfg.AudioSampleRate
	if rate == 0 {
		rate = audioSampleRate
	}
	cfg := map[string]interface{}{
		"session_config": map[string]interface{}{
			"app":     s.cfg.AppName,
			"session": s.cfg.SessionID,
			"audio_config": map[string]interface{}{
				"sample_rate_hertz": rate,
			},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

func (s *BidiSession) receiveLoop() {
	defer close(s.done)
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			s.mu.Lock()
			s.recvErr = err
			s.mu.Unlock()
			return
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			s.mu.Lock()
			s.recvErr = fmt.Errorf("malformed message: %w", err)
			s.mu.Unlock()
			return
		}
		s.mu.Lock()
		s.outputs = append(s.outputs, msg)
		s.mu.Unlock()
	}
}

// SendText sends a text input over the WebSocket.
func (s *BidiSession) SendText(text string) error {
	msg := map[string]interface{}{"text": text}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

// SendAudio streams raw audio bytes in chunks with a delay between each.
func (s *BidiSession) SendAudio(audio []byte) error {
	for i := 0; i < len(audio); i += audioChunkSize {
		end := i + audioChunkSize
		if end > len(audio) {
			end = len(audio)
		}
		chunk := audio[i:end]
		msg := map[string]interface{}{"audio": chunk}
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
		time.Sleep(chunkDelay)
	}
	return nil
}

// Outputs returns all messages received so far (thread-safe).
func (s *BidiSession) Outputs() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]map[string]interface{}, len(s.outputs))
	copy(result, s.outputs)
	return result
}

// Close closes the WebSocket connection.
func (s *BidiSession) Close() error {
	return s.conn.Close()
}

// Wait blocks until the receive goroutine exits (connection closed or error).
func (s *BidiSession) Wait() error {
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recvErr
}
