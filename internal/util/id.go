package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var counter uint64

func NewSessionID() string {
	ts := time.Now().UnixNano()
	c := atomic.AddUint64(&counter, 1)
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("sess_%d_%d_%s", ts, c, hex.EncodeToString(b))
}

func NewRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
