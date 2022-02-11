package client

import (
	"context"
	"math/rand"
	"time"

	"github.com/eiannone/keyboard"

	"github.com/ticker-es/client-go/eventstream/base"
	_ "github.com/vbauerster/mpb/v7"
)

func (s *Client) PlayEvents(ctx context.Context, events []base.Event, delay func(), progress func()) {
	for _, event := range events {
		if delay != nil {
			delay()
		}
		if _, err := s.Emit(ctx, event); err != nil {
			if ctx.Err() == context.Canceled {
				return
			} else {
				panic(err)
			}
		}
		if progress != nil {
			progress()
		}
	}
}

func ManualSuccession(cancel context.CancelFunc) func() {
	return func() {
		if ch, key, err := keyboard.GetSingleKey(); err == nil {
			if key == 0x03 || ch == 'q' {
				cancel()
				return
			}
		} else {
			panic(err)
		}
	}
}

func FixedDelay(delay int) func() {
	return func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func RandomDelay(delay int) func() {
	return func() {
		p := rand.Intn(delay)
		time.Sleep(time.Duration(p) * time.Millisecond)
	}
}
