package cluster

import (
	"sync/atomic"
)

var isPrimary atomic.Bool

func SetRole(asPrimary bool) {
	isPrimary.Store(asPrimary)
}

func IsPrimary() bool {
	return isPrimary.Load()
}
