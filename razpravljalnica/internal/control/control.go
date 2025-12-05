package control

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ControlNode struct {
	api.UnimplementedControlPlaneServer
	address        string
	grpcServer     *grpc.Server
	healthyClients []*shared.ServerNodeClient
	pendingClients []*shared.ServerNodeClient
	deadClients    []*shared.ServerNodeClient
}

var _ api.ControlPlaneServer = (*ControlNode)(nil)

func NewControlNode(host string, port, nServerNodes int) (*ControlNode, error) {
	c := ControlNode{
		address:    fmt.Sprintf("%v:%v", host, port),
		grpcServer: grpc.NewServer(),
	}

	// Prepare server clients
	serverClients := make([]*shared.ServerNodeClient, 0, nServerNodes)
	for i := range nServerNodes {
		address := fmt.Sprintf("%v:%v", host, port+i+1)
		conn, err := grpc.NewClient(address)

		if err != nil {
			for _, client := range serverClients {
				client.Conn.Close()
			}
			return nil, status.Errorf(codes.Internal, "failed to open client connection to %q: %v", address, err)
		}

		serverClients = append(serverClients, &shared.ServerNodeClient{
			Id:       i + 1,
			Conn:     conn,
			Health:   grpc_health_v1.NewHealthClient(conn),
			Internal: api.NewInternalMessageBoardServiceClient(conn),
		})
	}
	c.deadClients = serverClients

	// Register gRPC service
	api.RegisterControlPlaneServer(c.grpcServer, &c)
	reflection.Register(c.grpcServer)

	return &c, nil
}

func (c *ControlNode) Run() error {
	listener, err := net.Listen("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.runServersHealthLoop(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		shared.Logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
		c.grpcServer.GracefulStop()

		allClients := append(c.healthyClients, c.pendingClients...)
		allClients = append(allClients, c.deadClients...)
		for _, client := range allClients {
			client.Conn.Close()
		}
	}()

	shared.Logger.Info("control running", "address", c.address)

	if err := c.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("control stopped gracefully")
	return nil
}

func (c *ControlNode) runServersHealthLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			shared.Logger.InfoContext(ctx, "stopping health loop")
			return

		case <-ticker.C:
			shared.Logger.InfoContext(ctx, "running health check loop")
			cycleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

			c.checkServersHealth(cycleCtx)
			c.linkServers(cycleCtx)
			c.repairServers(cycleCtx)

			cancel()
		}
	}
}

func (c *ControlNode) checkServersHealth(ctx context.Context) {
	var newHealthyClients, newDeadClients, newPendingClients []*shared.ServerNodeClient

	var wg sync.WaitGroup

	for _, client := range c.healthyClients {
		defer wg.Done()
		_, err := client.Health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})

		if err != nil {
			newDeadClients = append(newDeadClients, client)
			shared.Logger.WarnContext(ctx, "server node is dead", "address", client.Conn.Target())
		} else {
			newHealthyClients = append(newHealthyClients, client)
			shared.Logger.InfoContext(ctx, "server node is healthy", "address", client.Conn.Target())
		}
	}

	for _, client := range c.pendingClients {
		defer wg.Done()
		res, err := client.Health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			newDeadClients = append(newDeadClients, client)
			shared.Logger.WarnContext(ctx, "pending server node is dead", "address", client.Conn.Target())
		} else if res.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			newHealthyClients = append(newHealthyClients, client)
			shared.Logger.InfoContext(ctx, "pending server node is ready", "address", client.Conn.Target())
		} else {
			newPendingClients = append(newPendingClients, client)
			shared.Logger.InfoContext(ctx, "pending server node is healthy", "address", client.Conn.Target())
		}
	}

	c.healthyClients = newHealthyClients
	c.pendingClients = newPendingClients
	c.deadClients = newDeadClients
}

func (c *ControlNode) linkServers(ctx context.Context) {
	for i, client := range c.healthyClients {
		isHead := i == 0
		isTail := i == len(c.healthyClients)-1

		req := api.RewireRequest{}
		if !isHead && !isTail {
			downstream := c.healthyClients[i-1]
			req.DownstreamId = shared.AnyPtr(int64(downstream.Id))
			req.DownstreamAddress = shared.AnyPtr(downstream.Conn.Target())
			upstream := c.healthyClients[i+1]
			req.UpstreamId = shared.AnyPtr(int64(upstream.Id))
			req.UpstreamAddress = shared.AnyPtr(upstream.Conn.Target())
		} else if isHead && !isTail {
			upstream := c.healthyClients[i+1]
			req.UpstreamId = shared.AnyPtr(int64(upstream.Id))
			req.UpstreamAddress = shared.AnyPtr(upstream.Conn.Target())
		} else if !isHead && isTail {
			downstream := c.healthyClients[i-1]
			req.DownstreamId = shared.AnyPtr(int64(downstream.Id))
			req.DownstreamAddress = shared.AnyPtr(downstream.Conn.Target())
		}

		shared.Logger.InfoContext(ctx, "rewiring server node", "address", client.Conn.Target())
		go client.Internal.Rewire(ctx, &req)
	}
}

func (c *ControlNode) repairServers(ctx context.Context) {
	for _, client := range c.deadClients {
		go func() {
			cmd := exec.Command(path.Join("..", "..", "build"), "--id", strconv.Itoa(client.Id), "--address", client.Conn.Target(), "--control", c.address)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Start(); err != nil {
				shared.Logger.ErrorContext(ctx, "failed to start server node", "address", client.Conn.Target())
			} else {
				shared.Logger.InfoContext(ctx, "starting server node", "address", client.Conn.Target())
			}
		}()
	}
}
