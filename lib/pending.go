package lib

import "slices"

type sequenced[T any] struct {
	sequence int32
	value    T
}

type Pending[T any] struct {
	last    int32
	entries []sequenced[T]
}

func (p *Pending[T]) Push(value T) int32 {
	p.last++
	p.entries = append(p.entries, sequenced[T]{
		sequence: p.last,
		value:    value,
	})

	return p.last
}

func (p *Pending[T]) Ack(sequence int32) []T {
	count := 0
	for _, entry := range p.entries {
		if entry.sequence > sequence {
			break
		}
		count++
	}
	if count == 0 {
		return nil
	}

	settled := make([]T, count)
	for index := range count {
		settled[index] = p.entries[index].value
	}
	p.entries = slices.Delete(p.entries, 0, count)

	return settled
}

func (p *Pending[T]) Each(visit func(*T)) {
	for index := range p.entries {
		visit(&p.entries[index].value)
	}
}

func (p *Pending[T]) Len() int {
	return len(p.entries)
}
