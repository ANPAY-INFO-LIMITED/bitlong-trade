package custodyMutex

import (
	"sync"
	"sync/atomic"
	"time"
)

type CustodyMutex struct {
	lock       *sync.Mutex
	expiryTime int64
}

var (
	custodyLocks sync.Map
	cleanupOnce  sync.Once
)

const expiryInterval = 5 * time.Minute

func GetCustodyMutex(key string) *sync.Mutex {

	cleanupOnce.Do(func() {
		go clearExpiredLocks()
	})

	timeNow := time.Now()

	newMutex := &CustodyMutex{
		lock:       &sync.Mutex{},
		expiryTime: timeNow.Add(expiryInterval).Unix(),
	}

	actual, loaded := custodyLocks.LoadOrStore(key, newMutex)
	if loaded {

		existingMutex := actual.(*CustodyMutex)
		existingMutex.UpdateExpiryTime(timeNow)
		return existingMutex.lock
	}

	return newMutex.lock
}

func clearExpiredLocks() {
	ticker := time.NewTicker(expiryInterval)
	defer ticker.Stop()

	for range ticker.C {
		cleanExpiredLocks()
	}
}

func cleanExpiredLocks() {
	now := time.Now().Unix()
	custodyLocks.Range(func(key, value interface{}) bool {
		cm := value.(*CustodyMutex)

		if atomic.LoadInt64(&cm.expiryTime) < now {
			custodyLocks.Delete(key)
		}
		return true
	})
}

func (cm *CustodyMutex) UpdateExpiryTime(timeNow time.Time) {
	newExpiry := timeNow.Add(expiryInterval).Unix()
	atomic.StoreInt64(&cm.expiryTime, newExpiry)
}
