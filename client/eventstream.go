package client

import (
	"context"
	"errors"
	"io"

	es "github.com/ticker-es/client-go/eventstream/base"
	"github.com/ticker-es/client-go/rpc"
)

var ErrInvalidClientID = errors.New("invalid clientID")

func (s *Client) Emit(ctx context.Context, event es.Event) (es.Event, error) {
	pub, err := s.eventStreamClient.Emit(ctx, rpc.EventToProto(&event))
	if err != nil {
		return event, err
	}
	if pub != nil {
		event.Sequence = pub.Sequence
		return event, err
	} else {
		return event, errors.New("didn't receive an Ack")
	}
}

func (s *Client) Stream(ctx context.Context, selector *es.Selector, bracket *es.Bracket, handler es.EventHandler) (int64, error) {
	req := &rpc.StreamRequest{
		Bracket:  rpc.BracketToProto(bracket),
		Selector: rpc.SelectorToProto(selector),
	}
	stream, err := s.eventStreamClient.Stream(ctx, req)
	if err != nil {
		return 0, err
	}
	var counter int64
	for {
		ev, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return counter, err
		}

		event := rpc.ProtoToEvent(ev)
		if err := handler(event); err != nil {
			return counter, err
		}
		counter++
	}
	return counter, nil
}

func (s *Client) Subscribe(ctx context.Context, clientID string, sel *es.Selector, handler es.EventHandler) error {
	if clientID == "" {
		return ErrInvalidClientID
	}
	req := &rpc.SubscriptionRequest{
		PersistentClientId: clientID,
		Selector:           rpc.SelectorToProto(sel),
	}
	if sub, err := s.eventStreamClient.Subscribe(ctx, req); err == nil {
		if ackStream, err := s.eventStreamClient.Acknowledge(ctx); err != nil {
			return err
		} else {
			defer ackStream.CloseSend()
			for {
				if ev, err := sub.Recv(); err == nil {
					event := rpc.ProtoToEvent(ev)
					if err := handler(event); err != nil {
						// TODO Check whether to close connection
						return err
					}
					ack := &rpc.Ack{
						PersistentClientId: clientID,
						Sequence:           event.Sequence,
					}
					if err := ackStream.Send(ack); err != nil {
						return err
					}
				} else {
					if err == io.EOF {
						// Server closed the connection
						break
					}
					return err
				}
			}
		}
	} else {
		return err
	}
	return nil
}
