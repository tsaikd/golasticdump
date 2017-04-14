package esdump

import (
	"context"
	"os"
	"os/signal"
)

func contextWithOSSignal(parent context.Context, sig ...os.Signal) context.Context {
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, sig...)

	ctx, cancel := context.WithCancel(parent)

	go func(cancel context.CancelFunc) {
		select {
		case sig := <-osSignalChan:
			logger.Info(sig)
			cancel()
		}
	}(cancel)

	return ctx
}
