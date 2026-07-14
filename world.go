package gocraft

import "sync"

type chunkPos struct {
	X int32
	Z int32
}

type World struct {
	mu      sync.RWMutex
	columns map[chunkPos]*ChunkColumn
}

func NewWorld() *World {
	return &World{columns: make(map[chunkPos]*ChunkColumn)}
}

func (w *World) LoadColumn(c *ChunkColumn) {
	w.mu.Lock()
	w.columns[chunkPos{c.X, c.Z}] = c
	w.mu.Unlock()
}

func (w *World) UnloadColumn(x, z int32) {
	w.mu.Lock()
	delete(w.columns, chunkPos{x, z})
	w.mu.Unlock()
}

func (w *World) Column(x, z int32) (*ChunkColumn, bool) {
	w.mu.RLock()
	c, ok := w.columns[chunkPos{x, z}]
	w.mu.RUnlock()

	return c, ok
}

func (w *World) Loaded() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return len(w.columns)
}

func (w *World) Block(x, y, z int) (BlockState, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	c, ok := w.columns[chunkPos{int32(x >> 4), int32(z >> 4)}]
	if !ok {
		return 0, false
	}

	return c.Block(x&15, y, z&15), true
}

func (w *World) SetBlock(x, y, z int, state BlockState) {
	w.mu.Lock()
	defer w.mu.Unlock()

	c, ok := w.columns[chunkPos{int32(x >> 4), int32(z >> 4)}]
	if !ok {
		return
	}

	c.SetBlock(x&15, y, z&15, state)
}
