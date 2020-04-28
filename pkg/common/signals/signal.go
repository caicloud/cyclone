package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Used to ensure GracefulShutdown is called at most once. When GracefulShutdown is called,
// this channel would be closed, if another call to close it again, panic would occur.
var once = make(chan struct{})

// GracefulShutdown catches signals of Interrupt, SIGINT, SIGTERM, SIGQUIT and cancel a context.
// If any signals caught, it will call the CancelFunc to cancel a context. If a second signal
// caught, exit directly with code 1.
func GracefulShutdown(cancel context.CancelFunc) {
	close(once)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-c
		log.WithField("signal", s).Info("System signal caught, cancel context.")
		cancel()
		s = <-c
		log.WithField("signal", s).Info("Another system signal caught, exit directly.")
		os.Exit(1)
	}()
}
