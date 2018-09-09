package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func SetupSignalHandler() (<-chan struct{}, chan<- error) {
	stopCh := make(chan struct{})
	errCh := make(chan error)
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			log.Info().Msgf("Terminating application due to signal %v", sig)
		case err := <-errCh:
			log.Error().Msgf("Terminating application due to error %v", err)
		}

		close(stopCh)
	}()

	return stopCh, errCh
}
