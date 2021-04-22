package base

import "math"

type Bracket struct {
	NextSequence int64
	LastSequence int64
}

func (s *Bracket) Sanitize(lastSequence int64) {
	if s.NextSequence < 1 {
		s.NextSequence = 1
	}
	if s.LastSequence <= 0 {
		s.LastSequence = lastSequence
	}
	if s.LastSequence > lastSequence {
		s.LastSequence = lastSequence
	}
}

func All() Bracket {
	return Bracket{
		NextSequence: 1,
		LastSequence: math.MaxInt64,
	}
}

func Range(next, last int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: last,
	}
}

func From(next int64) Bracket {
	return Bracket{
		NextSequence: next,
		LastSequence: math.MaxInt64,
	}
}
