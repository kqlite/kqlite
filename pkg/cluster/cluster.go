package cluster

import (
	"sync/atomic"
)

var isPrimary atomic.Bool

func SetPrimary() {
	isPrimary.Store(true)
}

func SetSecondary() {
	isPrimary.Store(false)
}

func IsPrimary() bool {
	return isPrimary.Load()
}
