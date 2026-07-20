package progress

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Spinner struct {
	mu       sync.Mutex
	active   bool
	frames   []string
	frameIdx int
	stopChan chan struct{}
	prefix   string
	done     int
	total    int
	label    string
	isTTY    bool
}

func NewSpinner() *Spinner {
	isTTY := false
	if info, err := os.Stdout.Stat(); err == nil {
		if (info.Mode() & os.ModeCharDevice) != 0 {
			isTTY = true
		}
	}
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stopChan: make(chan struct{}),
		isTTY:    isTTY,
	}
}

func (s *Spinner) Start(prefix string) {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.prefix = prefix
	s.mu.Unlock()

	if !s.isTTY {
		return
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				if !s.active {
					s.mu.Unlock()
					return
				}
				frame := s.frames[s.frameIdx]
				s.frameIdx = (s.frameIdx + 1) % len(s.frames)
				
				var msg string
				if s.total > 0 {
					msg = fmt.Sprintf("\r%s %s  %d / %d files  %s", frame, s.prefix, s.done, s.total, s.label)
				} else {
					msg = fmt.Sprintf("\r%s %s  %d files  %s", frame, s.prefix, s.done, s.label)
				}
				// Carriage return and clear line to avoid trailing characters
				fmt.Print("\r\033[K" + msg)
				s.mu.Unlock()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Spinner) Update(done, total int, label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.done = done
	s.total = total
	s.label = label
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	if s.isTTY {
		s.stopChan <- struct{}{}
		// Clear the line
		fmt.Print("\r\033[K")
	}
}
