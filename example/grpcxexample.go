package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/vsaien/grpcx"
	"github.com/vsaien/grpcx/config"
	"github.com/vsaien/grpcx/example/proto"
	"google.golang.org/grpc"
)

func main() {
	for i := 0; i < 3; i++ {
		go Server(i)
	}
	Client()
}

func Server(count int) {
	conf := &config.ServiceConf{
		EtcdAuth:      config.EtcdAuth{},
		Schema:        "www.vector.com",
		ServerName:    "knowing",
		Endpoints:     []string{"127.0.0.1:2379"},
		ServerAddress: "127.0.0.1:2000" + strconv.Itoa(count),
	}
	demo := &RegionHandlerServer{ServerAddress: conf.ServerAddress}
	rpcServer, err := grpcx.MustNewGrpcxServer(conf, func(server *grpc.Server) {
		proto.RegisterRegionHandlerServer(server, demo)
	})
	if err != nil {
		panic(err)
	}
	log.Fatal(rpcServer.Run())
}

type RegionHandlerServer struct {
	ServerAddress string
}

func (s *RegionHandlerServer) GetListenAudio(ctx context.Context, r *proto.FindRequest) (*proto.HasOptionResponse, error) {

	has := []*proto.HasOption(nil)
	for _, t := range r.Tokens {
		has = append(has, &proto.HasOption{Token: t + s.ServerAddress, Listen: 1})
	}
	res := &proto.HasOptionResponse{
		Items: has,
	}
	return res, nil
}

func Client() {
	conf := &config.ClientConf{
		EtcdAuth:  config.EtcdAuth{},
		Target:    "www.vector.com:///knowing",
		Endpoints: []string{"127.0.0.1:2379"},
		WithBlock: false,
	}

	r, err := grpcx.MustNewGrpcxClient(conf)
	if err != nil {
		panic(err)
	}
	conn, err := r.NextConnection()
	if err != nil {
		panic(err)
	}
	regionHandlerClient := proto.NewRegionHandlerClient(conn)
	for {
		res, err := regionHandlerClient.GetListenAudio(
			context.Background(),
			&proto.FindRequest{Tokens: []string{"a_"}},
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
		time.Sleep(2 * time.Second)
	}
}
