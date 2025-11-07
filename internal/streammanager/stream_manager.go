package streammanager

import (
	"log"
	"sync"
)

type DeployEvent struct {
	//Status  string `json:"status"`
	//Message string `json:"message"`
	DeploymentId string `json:"deployment_id"`
	Timestamp    string `json:"timestamp"`
	SiteID       string `json:"site_id"`
	Message      string `json:"message"`
	Status       string `json:"status"` // pending, in-progress, completed, failed

}

type StreamManager struct {
	mu      sync.Mutex
	streams map[string][]chan DeployEvent
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string][]chan DeployEvent),
	}
}

func (s *StreamManager) Register(deployID string) chan DeployEvent {
	log.Println("xxxxxxxx SM Register called", deployID)

	ch := make(chan DeployEvent, 10)
	s.mu.Lock()
	s.streams[deployID] = append(s.streams[deployID], ch)

	log.Println("s.streams[deployID]:", s.streams[deployID])
	s.mu.Unlock()
	return ch
}

func (s *StreamManager) Unregister(deployID string, ch chan DeployEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	channels := s.streams[deployID]
	for i, c := range channels {
		if c == ch {
			close(c)
			s.streams[deployID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(s.streams[deployID]) == 0 {
		delete(s.streams, deployID)
	}
}

func (s *StreamManager) Broadcast(deployID string, ev DeployEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("xxxxxxxx SM Brodcast called", deployID, s.streams[deployID])
	for _, ch := range s.streams[deployID] {
		select {
		case ch <- ev:
		default:
			// drop slow clients
		}
	}
}
