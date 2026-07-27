package audit

import (
	"sync"
	"time"
)

type Entry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	Success   bool      `json:"success"`
}

type Log struct {
	mu      sync.RWMutex
	entries []Entry
	max     int
	seq     int64
}

func New(max int) *Log {
	if max <= 0 {
		max = 500
	}
	return &Log{max: max}
}

func (l *Log) Record(actor, action, resource, detail string, success bool) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq++
	e := Entry{
		ID:        formatID(l.seq),
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		Success:   success,
	}
	l.entries = append([]Entry{e}, l.entries...)
	if len(l.entries) > l.max {
		l.entries = l.entries[:l.max]
	}
	return e
}

func (l *Log) List(limit int) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}
	out := make([]Entry, limit)
	copy(out, l.entries[:limit])
	return out
}

func formatID(n int64) string {
	return time.Now().UTC().Format("20060102150405") + "-" + itoa(n)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
