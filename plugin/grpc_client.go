package plugin

import (
	"context"

	pb "example.com/mud/plugin/proto"
)

// grpcClient is the client-side wrapper used by the engine.
type grpcClient struct {
	client pb.OrbisGameClient
}

func (c *grpcClient) GetManifest(ctx context.Context) (*pb.GameManifest, error) {
	return c.client.GetManifest(ctx, &pb.Empty{})
}

func (c *grpcClient) HandleEvent(ctx context.Context, req *pb.EventRequest) (*pb.ActionList, error) {
	return c.client.HandleEvent(ctx, req)
}
