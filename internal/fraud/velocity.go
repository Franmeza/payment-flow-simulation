package fraud

import (
	"sync"
	"time"
)

const (
	velocityWindow   = 60 * time.Second
	velocityMaxCount = 3
)

// Velocity tracks per-card transaction timestamps in a sliding window.
// It is safe for concurrent use.
type Velocity struct {
	mu         sync.Mutex // lock that allows only one goroutine at a time into a critical section.
	timestamps map[string][]time.Time // 
}

// NewVelocity returns a ready-to-use Velocity checker.
func NewVelocity() *Velocity {
	return &Velocity{
		timestamps: make(map[string][]time.Time),  
	}
}

/* Check records this transaction attempt and returns true when the card has
reached or exceeded velocityMaxCount attempts within the velocityWindow.
Declined attempts count toward the limit so attackers cannot hammer the
endpoint without consequence.*/
func (v *Velocity) Check(cardUID string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-velocityWindow)

	existing := v.timestamps[cardUID]
	fresh := existing[:0]
	for _, t := range existing {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}

	fresh = append(fresh, now)
	v.timestamps[cardUID] = fresh

	return len(fresh) >= velocityMaxCount
}
