package plugin

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "example.com/mud/plugin/proto"
)

// Handshake is shared between engine and game binary to verify compatibility.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ORBIS_PLUGIN",
	MagicCookieValue: "orbis-mud-engine-v1",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"game": &GameGRPCPlugin{},
}

// GameGRPCPlugin is the plugin.Plugin implementation for gRPC.
type GameGRPCPlugin struct {
	plugin.Plugin
	Impl GameServer
}

func (p *GameGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterOrbisGameServer(s, &grpcServer{impl: p.Impl})
	return nil
}

func (p *GameGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &grpcClient{client: pb.NewOrbisGameClient(c)}, nil
}

// Satisfy net/rpc plugin interface (unused but required by go-plugin)
func (p *GameGRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) { return nil, nil }
func (p *GameGRPCPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, nil
}

// GameServer is the interface implemented by the game binary.
type GameServer interface {
	GetManifest(context.Context, *pb.Empty) (*pb.GameManifest, error)
	HandleEvent(context.Context, *pb.EventRequest) (*pb.ActionList, error)
	EventStream(pb.OrbisGame_EventStreamServer) error
}

// GameClient is the interface used by the engine to call the game binary.
type GameClient interface {
	GetManifest(ctx context.Context) (*pb.GameManifest, error)
	HandleEvent(ctx context.Context, req *pb.EventRequest) (*pb.ActionList, error)
}
