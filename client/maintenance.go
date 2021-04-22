package client

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Client) PrintServerState(ctx context.Context) {
	if state, err := s.maintenanceClient.GetServerState(ctx, &emptypb.Empty{}); err == nil {
		fmt.Printf("uptime: %5ds   |   active connections: %3d   |   events stored: %8d\n", state.Uptime, state.ConnectionCount, state.EventCount)
	}
}
