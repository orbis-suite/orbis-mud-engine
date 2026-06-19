package plugin

import (
	"context"

	pb "example.com/mud/plugin/proto"
)

// grpcServer wraps a GameServer implementation and serves it over gRPC.
type grpcServer struct {
	pb.UnimplementedOrbisGameServer
	impl GameServer
}

func (s *grpcServer) GetManifest(ctx context.Context, req *pb.Empty) (*pb.GameManifest, error) {
	return s.impl.GetManifest(ctx, req)
}

func (s *grpcServer) HandleEvent(ctx context.Context, req *pb.EventRequest) (*pb.ActionList, error) {
	return s.impl.HandleEvent(ctx, req)
}

func (s *grpcServer) EventStream(stream pb.OrbisGame_EventStreamServer) error {
	return s.impl.EventStream(stream)
}
