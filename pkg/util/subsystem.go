package util

import (
	"sync"
)

type Subsystem interface {
	Init() error
	Start(*SubsystemControl)
}

type SubsystemControl struct {
	Stop      <-chan struct{}
	WaitGroup *sync.WaitGroup
	errCh     chan<- error
}

func SetupSubsystemControl() *SubsystemControl {
	stopCh, errCh := SetupSignalHandler()

	return &SubsystemControl{
		errCh:     errCh,
		Stop:      stopCh,
		WaitGroup: &sync.WaitGroup{},
	}
}

func InitSubsystems(systems ...Subsystem) {
	for _, system := range systems {
		if err := system.Init(); err != nil {
			panic(err)
		}
	}
}

func StartSubsystems(systems ...Subsystem) {
	ctl := SetupSubsystemControl()

	for _, system := range systems {
		system.Start(ctl)
	}

	ctl.WaitGroup.Wait()
}

func (ctl *SubsystemControl) Error(err error) {
	select {
	case ctl.errCh <- err:
	default:
	}
}
