package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

// Session is a single active PTY session.
type Session struct {
	ID       string
	ptmx     *os.File
	cmd      *exec.Cmd
	mu       sync.Mutex
	lastSeen time.Time
}

// Manager manages PTY sessions with idle-timeout reaping.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*Session
	maxSess  int
	timeout  time.Duration
}

// NewManager creates a Manager. idleTimeout is how long a session survives without a WS connection.
func NewManager(maxSessions int, idleTimeout time.Duration) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		maxSess:  maxSessions,
		timeout:  idleTimeout,
	}
	go m.reaper()
	return m
}

// Create allocates a new PTY session. It runs bash as the current user.
func (m *Manager) Create(id string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; exists {
		return m.sessions[id], nil
	}
	if len(m.sessions) >= m.maxSess {
		return nil, fmt.Errorf("max terminal sessions (%d) reached", m.maxSess)
	}

	shell, err := exec.LookPath("bash")
	if err != nil {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	s := &Session{ID: id, ptmx: ptmx, cmd: cmd, lastSeen: time.Now()}
	m.sessions[id] = s

	go func() {
		_ = cmd.Wait()
		_ = ptmx.Close()
		m.mu.Lock()
		delete(m.sessions, id)
		m.mu.Unlock()
	}()

	return s, nil
}

// Get returns an existing session and refreshes its idle timer.
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if ok {
		s.lastSeen = time.Now()
	}
	return s, ok
}

// Close terminates a session immediately.
func (m *Manager) Close(id string) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if ok {
		s.terminate()
	}
}

// Count returns the number of active sessions.
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// CloseAll terminates all active sessions (called on server shutdown).
func (m *Manager) CloseAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.Close(id)
	}
}

func (m *Manager) reaper() {
	for range time.Tick(15 * time.Second) {
		m.mu.Lock()
		now := time.Now()
		var dead []*Session
		for id, s := range m.sessions {
			if now.Sub(s.lastSeen) > m.timeout {
				dead = append(dead, s)
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()
		for _, s := range dead {
			s.terminate()
		}
	}
}

// Write sends bytes to the PTY (terminal input from client).
func (s *Session) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.ptmx.Write(data)
	return err
}

// Resize adjusts the PTY window size.
func (s *Session) Resize(rows, cols uint16) error {
	return pty.Setsize(s.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
}

// ReadLoop reads PTY output and calls fn for each chunk, until the PTY closes or ctx is done.
// Uses a 300ms read deadline so ctx cancellation is noticed promptly.
func (s *Session) ReadLoop(ctx context.Context, fn func([]byte)) error {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_ = s.ptmx.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			fn(buf[:n])
		}
		if err != nil {
			if isTimeout(err) {
				continue // deadline expired — check ctx and try again
			}
			if err == io.EOF || errors.Is(err, os.ErrClosed) {
				return nil
			}
			return err
		}
	}
}

func isTimeout(err error) bool {
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func (s *Session) terminate() {
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Signal(syscall.SIGHUP)
	}
	_ = s.ptmx.Close()
}
