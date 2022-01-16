package datastore

import "sync"

type Cache struct {
	mu    sync.RWMutex
	codes map[string]string
}
