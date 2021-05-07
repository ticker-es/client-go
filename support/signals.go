package support

import (
	"context"
	"os"
	"os/signal"
)

func CancelContextOnSignals(parent context.Context, signals ...os.Signal) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, signals...)
	go func() {
		<-signalChannel
		cancel()
	}()
	return ctx, cancel
}
