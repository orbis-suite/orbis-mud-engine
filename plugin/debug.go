package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "example.com/mud/plugin/proto"
)

const DebugAddrFile = ".orbis-debug-addr"

type debugAddr struct {
	Addr string `json:"addr"`
}

// ServeDebug starts a standalone gRPC server for the game, writes its address
// to DebugAddrFile, and blocks until the context is cancelled or the process exits.
// Use this instead of plugin.Serve() when running the game binary under a debugger.
func ServeDebug(ctx context.Context, impl GameServer) error {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("debug serve: listen: %w", err)
	}

	addr := lis.Addr().String()
	data, _ := json.Marshal(debugAddr{Addr: addr})
	if err := os.WriteFile(DebugAddrFile, data, 0o644); err != nil {
		return fmt.Errorf("debug serve: write addr file: %w", err)
	}
	defer os.Remove(DebugAddrFile)

	fmt.Printf("[orbis debug] game listening on %s\n", addr)
	fmt.Printf("[orbis debug] waiting for engine to connect...\n")

	s := grpc.NewServer()
	pb.RegisterOrbisGameServer(s, &grpcServer{impl: impl})

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve(lis) }()

	select {
	case <-ctx.Done():
		s.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

// ConnectDebug connects directly to a game binary that was started with ServeDebug.
// It waits up to timeout for DebugAddrFile to appear (giving the debugger time to
// start the game process).
func ConnectDebug(timeout time.Duration) (GameClient, func(), error) {
	addr, err := waitForDebugAddr(timeout)
	if err != nil {
		return nil, nil, err
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("debug connect: %w", err)
	}

	cleanup := func() { conn.Close() }
	return &grpcClient{client: pb.NewOrbisGameClient(conn)}, cleanup, nil
}

func waitForDebugAddr(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	fmt.Printf("[orbis debug] waiting up to %s for game to start (launch game binary with -debug)...\n", timeout)

	for time.Now().Before(deadline) {
		data, err := os.ReadFile(DebugAddrFile)
		if err == nil {
			var a debugAddr
			if json.Unmarshal(data, &a) == nil && a.Addr != "" {
				fmt.Printf("[orbis debug] connecting to game at %s\n", a.Addr)
				return a.Addr, nil
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	return "", fmt.Errorf("timed out waiting for game debug server (is the game running with -debug?)")
}

// Ensure grpc import used
var _ = context.Background
