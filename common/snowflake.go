package common

import (
	"sync"
	"time"
)

const (
	epoch          = int64(1704038400000) // 2024-01-01 00:00:00 UTC
	workerBits     = 5
	maxWorker      = -1 ^ (-1 << workerBits)
	sequenceBits   = 12
	sequenceMask   = -1 ^ (-1 << sequenceBits)
	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Snowflake struct {
	mu        sync.Mutex
	workerID  int64
	sequence  int64
	lastStamp int64
}

func NewSnowflake(workerID int64) *Snowflake {
	if workerID < 0 || workerID > maxWorker {
		workerID = 0
	}
	return &Snowflake{workerID: workerID}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastStamp {
		now = s.lastStamp
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = now
	return ((now - epoch) << timestampShift) | (s.workerID << workerShift) | s.sequence
}

var defaultSF = NewSnowflake(1)

func GenID() int64 {
	return defaultSF.NextID()
}
