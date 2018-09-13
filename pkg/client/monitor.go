package client

import (
	"context"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"google.golang.org/grpc"
)

func ConnectToMonitor(monitors []string) (monitor.Client, error) {
	cc, err := grpc.Dial(monitors[0], grpc.WithTimeout(15*time.Second), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return monitor.FromClientConn(cc), nil
}

func WatchVolumes(client monitor.Client, ctx context.Context, revision int64) (chan *api.Volume, error) {
	ch := make(chan *api.Volume, 10)
	watch, err := client.Watch(ctx)
	if err != nil {
		return nil, err
	}
	err = watch.Send(&api.WatchVolumesRequest{
		Header: &api.WatchRequestHeader{
			StartRevision: revision,
			End:           false,
		},
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			vols, err := watch.Recv()
			if err != nil {
				break
			}
			for _, v := range vols.Volumes {
				select {
				case ch <- v:
				case <-ctx.Done():
				}
			}
		}
		close(ch)
	}()

	return ch, nil
}
