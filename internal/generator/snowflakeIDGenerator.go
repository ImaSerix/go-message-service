package generator

import (
	"sync"
	"time"
)

type SnowflakeIDGenerator struct {
	nodeID        int64
	sequence      int64
	lastTimestamp int64
	mutex         sync.Mutex
	epoch         int64
}

type SnowflakeID int64

func NewSnowflakeIDGenerator(nodeID int64, epoch int64) *SnowflakeIDGenerator {
	return &SnowflakeIDGenerator{
		nodeID: nodeID,
		epoch:  epoch,
	}
}

const (
	sequenceBits = 12
	nodeBits     = 10
	maxSequence  = (1 << sequenceBits) - 1
	timeShift    = nodeBits + sequenceBits
	nodeShift    = sequenceBits
)

func (gen *SnowflakeIDGenerator) NewID() SnowflakeID {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	for {
		timeStamp := time.Now().UnixMilli() - gen.epoch
		if timeStamp < gen.lastTimestamp {
			timeStamp = gen.lastTimestamp
		}
		if timeStamp == gen.lastTimestamp {
			gen.sequence++
			if gen.sequence > maxSequence {
				continue
			}
		} else {
			gen.sequence = 0
		}
		gen.lastTimestamp = timeStamp
		id := (timeStamp << timeShift) | (gen.nodeID << nodeShift) | gen.sequence

		return SnowflakeID(id)
	}
}
