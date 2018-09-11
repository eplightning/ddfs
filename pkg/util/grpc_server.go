package util

import (
	"net"

	"google.golang.org/grpc"
)

const MaxMessageSize = 16 * 1024 * 1024

type GrpcServer struct {
	listen  net.Listener
	Server  *grpc.Server
	address string
}

func NewGrpcServer(address string) *GrpcServer {
	return &GrpcServer{
		address: address,
	}
}

func (s *GrpcServer) Init() error {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.listen = listen
	s.Server = grpc.NewServer(
		grpc.MaxRecvMsgSize(MaxMessageSize),
		grpc.MaxSendMsgSize(MaxMessageSize),
	)
	return nil
}

func (s *GrpcServer) Start(ctl *SubsystemControl) {
	ctl.WaitGroup.Add(2)
	go func() {
		defer ctl.WaitGroup.Done()

		err := s.Server.Serve(s.listen)
		if err != nil && err != grpc.ErrServerStopped {
			ctl.Error(err)
		}
	}()
	go func() {
		defer ctl.WaitGroup.Done()
		<-ctl.Stop
		s.Server.Stop()
	}()
}
