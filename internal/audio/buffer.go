package audio

import (
	"sync"
)

type Buffer struct {
	mu      sync.Mutex
	data    []byte
	maxSize int
	notify  chan struct{}
	closed  bool
}

func NewBuffer(maxSize int) *Buffer {
	return &Buffer{
		data:    make([]byte, 0, maxSize),
		maxSize: maxSize,
		notify:  make(chan struct{}, 1),
	}
}

func (b *Buffer) Write(chunk []byte) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0
	}

	available := b.maxSize - len(b.data)
	if available <= 0 {
		return 0
	}

	toWrite := len(chunk)
	if toWrite > available {
		toWrite = available
	}

	b.data = append(b.data, chunk[:toWrite]...)

	select {
	case b.notify <- struct{}{}:
	default:
	}

	return toWrite
}

func (b *Buffer) Read(size int) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data) == 0 {
		return nil
	}

	toRead := size
	if toRead > len(b.data) {
		toRead = len(b.data)
	}

	result := make([]byte, toRead)
	copy(result, b.data[:toRead])
	b.data = b.data[toRead:]

	return result
}

func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}

func (b *Buffer) Wait() <-chan struct{} {
	return b.notify
}

func (b *Buffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	close(b.notify)
}

func (b *Buffer) IsClosed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.closed
}
