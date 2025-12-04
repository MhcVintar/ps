package control

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var healthCheckPeriod = 30 * time.Second

type chainClient struct {
	id       string
	conn     *grpc.ClientConn
	health   grpc_health_v1.HealthClient
	internal api.InternalMessageBoardServiceClient
}

type ControlNode struct {
	api.UnimplementedControlPlaneServer
	host           string
	port           int
	grpcServer     *grpc.Server
	healthyClients []*chainClient // This is in the order of the actual chain
	pendingClient  *chainClient
	deadClients    []*chainClient
}

var _ api.ControlPlaneServer = (*ControlNode)(nil)

func NewControlNode(host string, port, wantedChainSize int) (*ControlNode, error) {
	n := ControlNode{
		host:           host,
		port:           port,
		grpcServer:     grpc.NewServer(),
		healthyClients: nil,
		pendingClient:  nil,
	}

	// Prepare chain clients
	chainClients := make([]*chainClient, 0, wantedChainSize)
	for i := range wantedChainSize {
		cleanup := func() {
			for _, client := range chainClients {
				client.conn.Close()
			}
		}

		address := fmt.Sprintf("%v:%v", host, port+i+1)
		conn, err := grpc.NewClient(address)

		if err != nil {
			cleanup()
			return nil, status.Errorf(codes.Internal, "failed to open client connection to %q: %v", address, err)
		}

		id, err := uuid.NewUUID()
		if err != nil {
			cleanup()
			return nil, status.Errorf(codes.Internal, "failed to generate client id: %v", address, err)
		}

		chainClients = append(chainClients, &chainClient{
			id:       id.String(),
			conn:     conn,
			health:   grpc_health_v1.NewHealthClient(conn),
			internal: api.NewInternalMessageBoardServiceClient(conn),
		})
	}
	n.deadClients = chainClients

	// Register gRPC service
	api.RegisterControlPlaneServer(n.grpcServer, &n)
	reflection.Register(n.grpcServer)

	return &n, nil
}

func (c *ControlNode) Run() error {
	address := fmt.Sprintf("%v:%v", c.host, c.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.runChainHealthLoop(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		shared.Logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
		c.grpcServer.GracefulStop()
	}()

	shared.Logger.Info("control running", "address", address)

	if err := c.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("control stopped gracefully")
	return nil
}

func (c *ControlNode) runChainHealthLoop(ctx context.Context) {
	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			shared.Logger.Info("stopping health loop")
			return

		case <-ticker.C:
			cycleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

			c.checkChainHealth(cycleCtx)
			c.linkChain(cycleCtx)
			c.repairChain(cycleCtx)

			cancel()
		}
	}
}

func (c *ControlNode) checkChainHealth(ctx context.Context) {
	var wg sync.WaitGroup

	dead := make([]bool, len(c.healthyClients))
	for i, client := range c.healthyClients {
		wg.Add(1)

		go func() {
			defer wg.Done()
			_, err := client.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				dead[i] = true
			}
		}()
	}

	wg.Wait()

	var newHealthyClients []*chainClient
	var newDeadClients []*chainClient
	for i, client := range c.healthyClients {
		if dead[i] {
			newDeadClients = append(newDeadClients, client)
		} else {
			newHealthyClients = append(newHealthyClients, client)
		}
	}

	if c.pendingClient != nil {
		res, err := c.pendingClient.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			newDeadClients = append(newDeadClients, c.pendingClient)
			c.pendingClient = nil
		} else if res.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			newHealthyClients = append(newHealthyClients, c.pendingClient)
			c.pendingClient = nil
		}
	}

	c.healthyClients = newHealthyClients
	c.deadClients = newDeadClients
}

func (c *ControlNode) linkChain(ctx context.Context) {
	for i, client := range c.healthyClients {
		isHead := i == 0
		isTail := i == len(c.healthyClients)-1

		req := api.RewireRequest{}
		if isHead && isTail {
			req.DownstreamAddress = shared.AnyPtr("")
			req.UpstreamAddress = shared.AnyPtr("")
		} else if isHead {
			req.DownstreamAddress = shared.AnyPtr("")
			req.UpstreamAddress = shared.AnyPtr(c.healthyClients[i+1].conn.Target())
		} else if isTail {
			req.DownstreamAddress = shared.AnyPtr(c.healthyClients[i-1].conn.Target())
			req.UpstreamAddress = shared.AnyPtr("")
		} else {
			req.DownstreamAddress = shared.AnyPtr(c.healthyClients[i-1].conn.Target())
			req.UpstreamAddress = shared.AnyPtr(c.healthyClients[i+1].conn.Target())
		}

		go client.internal.Rewire(ctx, &req)
	}
}

func (c *ControlNode) repairChain(ctx context.Context) {
	if len(c.deadClients) == 0 || c.pendingClient != nil {
		return
	}

	c.pendingClient = c.deadClients[0]
	c.deadClients = c.deadClients[1:]

	// TODO actually create a new client
}
