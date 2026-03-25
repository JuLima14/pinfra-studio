package sandbox

import (
	"fmt"
	"sync"
)

type PortAllocator struct {
	mu       sync.Mutex
	minPort  int
	maxPort  int
	assigned map[int]bool
}

func NewPortAllocator(min, max int) *PortAllocator {
	return &PortAllocator{minPort: min, maxPort: max, assigned: make(map[int]bool)}
}

func (pa *PortAllocator) Allocate() (int, error) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	for port := pa.minPort; port <= pa.maxPort; port++ {
		if !pa.assigned[port] {
			pa.assigned[port] = true
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", pa.minPort, pa.maxPort)
}

func (pa *PortAllocator) Release(port int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	delete(pa.assigned, port)
}

func (pa *PortAllocator) Reserve(port int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.assigned[port] = true
}
