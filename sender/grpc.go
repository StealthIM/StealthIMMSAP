package sender

import (
	pb "StealthIMMSAP/StealthIM.MSAP"
	"StealthIMMSAP/config"
	"context"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

var cfg config.Config

type server struct {
	pb.StealthIMMSAPSyncServer
}

func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.Pong, error) {
	return &pb.Pong{}, nil
}

// Start 启动 GRPC 服务
func Start(rCfg config.Config) {
	cfg = rCfg
	lis, err := net.Listen("tcp", rCfg.SyncGRPCProxy.Host+":"+strconv.Itoa(rCfg.SyncGRPCProxy.Port))
	if err != nil {
		log.Fatalf("[SYNCER]Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStealthIMMSAPSyncServer(s, &server{})
	log.Printf("[SYNCER]Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("[SYNCER]Failed to serve: %v", err)
	}
}
