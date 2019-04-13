package grpcx

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/etcd/clientv3"
	"github.com/vsaien/grpcx/config"
	"github.com/vsaien/grpcx/register"
	"github.com/vsaien/log4g"
	"google.golang.org/grpc"
)

var (
	deadSignal = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}
)

type (
	GrpcxServiceFunc func(server *grpc.Server)
	GrpcxServer      struct {
		register       *register.Register
		rpcServiceFunc GrpcxServiceFunc
	}
)

func MustNewGrpcxServer(conf *config.ServiceConf, rpcServiceFunc GrpcxServiceFunc) (*GrpcxServer, error) {
	client3, err := clientv3.New(
		clientv3.Config{
			Endpoints:   conf.Endpoints,
			Username:    conf.UserName,
			Password:    conf.PassWord,
			DialTimeout: config.GrpcxDialTimeout,
		})
	if nil != err {
		return nil, err
	}
	return &GrpcxServer{
		register: register.NewRegister(
			conf.Schema,
			conf.ServerName,
			conf.ServerAddress,
			client3,
		),
		rpcServiceFunc: rpcServiceFunc,
	}, nil
}

func (s *GrpcxServer) Run(serverOptions ...grpc.ServerOption) error {
	listen, err := net.Listen("tcp", s.register.GetServerAddress())
	if nil != err {
		return err
	}
	log4g.InfoFormat(
		"serverAddress [%s] of %s Rpc server has started and full key [%s]",
		s.register.GetServerAddress(),
		s.register.GetServerName(),
		s.register.GetFullAddress(),
	)
	if err := s.register.Register(); err != nil {
		return err
	}
	server := grpc.NewServer(serverOptions...)
	s.rpcServiceFunc(server)
	s.deadNotify()
	if err := server.Serve(listen); nil != err {
		return err
	}
	return nil

}

func (s *GrpcxServer) deadNotify() {
	ch := make(chan os.Signal, 1) //
	signal.Notify(ch, deadSignal...)
	go func() {
		log4g.InfoFormat(
			"serverAddress [%s] of %s Rpc server has existed with got signal [%v] and full key [%s]",
			s.register.GetServerAddress(),
			s.register.GetServerName(),
			<-ch,
			s.register.GetFullAddress(),
		)
		if err := s.register.UnRegister(); err != nil {
			log4g.InfoFormat(
				"serverAddress [%s] of %s Rpc server UnRegister fail and  err %+v and full key [%s]",
				s.register.GetServerAddress(),
				s.register.GetServerName(),
				s.register.GetFullAddress(),
				err,
				s.register.GetFullAddress(),
			)
		}
		os.Exit(1) //
	}()
	return
}
