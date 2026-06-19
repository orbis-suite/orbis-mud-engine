// Package sdk provides types and helpers for implementing Orbis game plugins.
package sdk

import (
	"context"

	pb "example.com/mud/plugin/proto"
)

// Game is the interface implemented by a game binary.
type Game interface {
	GetManifest() *Manifest
	HandleEvent(e *Event) []Action
}

// Adapter wraps a Game and implements the plugin.GameServer gRPC interface.
type Adapter struct {
	impl Game
}

func NewAdapter(g Game) *Adapter {
	return &Adapter{impl: g}
}

func (a *Adapter) GetManifest(ctx context.Context, _ *pb.Empty) (*pb.GameManifest, error) {
	m := a.impl.GetManifest()
	return m.toProto(), nil
}

func (a *Adapter) HandleEvent(ctx context.Context, req *pb.EventRequest) (*pb.ActionList, error) {
	ev := eventFromProto(req)
	acts := a.impl.HandleEvent(ev)
	protoActs := make([]*pb.Action, 0, len(acts))
	for _, act := range acts {
		protoActs = append(protoActs, act.toProto())
	}
	return &pb.ActionList{Actions: protoActs}, nil
}

func (a *Adapter) EventStream(stream pb.OrbisGame_EventStreamServer) error {
	// Consume and ignore engine updates for now
	for {
		if _, err := stream.Recv(); err != nil {
			return nil
		}
	}
}
