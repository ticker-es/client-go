package base

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSequenceNotFound = errors.New("sequence not found")
)

type Event struct {
	Sequence   int64                  `json:"sequence,omitempty" yaml:"sequence,omitempty"`
	Aggregate  []string               `json:"aggregate,omitempty" yaml:"aggregate,omitempty"`
	Type       string                 `json:"type,omitempty" yaml:"type,omitempty"`
	OccurredAt time.Time              `json:"occurred_at,omitempty" yaml:"occurred_at,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty" yaml:"payload,omitempty"`
}

type EventHandler func(e *Event) error

type EventStream interface {
	Emit(event *Event) (int64, error)
	LastSequence() int64
	Get(sequence int64) (*Event, error)
	Stream(ctx context.Context, sel Selector, bracket Bracket, handler EventHandler) error
	Subscribe(ctx context.Context, persistentClientID string, sel Selector, handler EventHandler) (Subscription, error)
	// Subscriptions returns all currently known Subscriptions.
	Subscriptions() []Subscription
}

type Subscription interface {
	PersistentID() string
	// ActiveSelector returns the currently active Selector.
	ActiveSelector() Selector
	LastAcknowledgedSequence() (int64, error)
	Acknowledge(sequence int64) error
	// Active returns whether this Subscription is currently active.
	Active() bool
	// InactiveSince returns the time this Subscription last became inactive.
	InactiveSince() time.Time
	// Wait for the Subscription to become inactive (disconnected)
	Wait() error
	// DropOuts returns how often this Subscription has dropped out of the live stream.
	DropOuts() int
	// Shutdown closes this Subscription and removes all associated state. A Subscription can not be resumed after this call.
	Shutdown()
}

type EventStore interface {
	Store(event *Event) (int64, error)
	LastKnownSequence() int64
	Get(sequence int64) (*Event, error)
}

type SequenceStore interface {
	Get(persistentClientID string) (int64, error)
	Store(persistentClientID string, sequence int64) error
}
