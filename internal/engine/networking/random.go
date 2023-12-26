// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import (
	"crypto/rand"
	"sync"
	"unsafe"
)

var (
	randomPool      = make([]byte, 10000)
	randomPoolIndex = len(randomPool)
	randomPoolLock  = sync.Mutex{}
)

func mustReadRandUnsafe() {
	_, err := rand.Read(randomPool)
	if err != nil {
		panic(err)
	}
	randomPoolIndex = 0
}

func randBytes(b []byte) {
	// Lock the pool.
	randomPoolLock.Lock()
	defer randomPoolLock.Unlock()

remainingCheck:
	// Get the remaining number of bytes in the pool.
	remaining := len(randomPool) - randomPoolIndex

	// If there are enough bytes in the pool, copy them to the byte slice.
	if remaining >= len(b) {
		copy(b, randomPool[randomPoolIndex:randomPoolIndex+len(b)])
		randomPoolIndex += len(b)
		return
	}

	// If there are zero, refill the pool and jump back to the start.
	if remaining == 0 {
		mustReadRandUnsafe()
		goto remainingCheck
	}

	// If there are not enough bytes in the pool, copy the remaining bytes
	// and refill the pool, then reslice.
	copy(b, randomPool[randomPoolIndex:])
	mustReadRandUnsafe()
	b = b[remaining:]
	goto remainingCheck
}

const strChoice = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	// Lock the pool.
	randomPoolLock.Lock()
	defer randomPoolLock.Unlock()

	// Process the slice.
	b := make([]byte, n)
	choiceLen := uint8(len(strChoice))
	for i := range b {
		// Make sure there is data in the pool.
		if len(randomPool) == randomPoolIndex {
			mustReadRandUnsafe()
		}

		// Get a random byte from the pool.
		randByte := randomPool[randomPoolIndex]
		randomPoolIndex++

		// Get a random index from the choice string.
		b[i] = strChoice[randByte%choiceLen]
	}

	// Unsafely convert the byte slice to a string since it will never be mutated again.
	return unsafe.String(&b[0], n)
}
