package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ControlNode struct {
	api.UnimplementedControlPlaneServer
	id                 string
	grpcAddress        string
	raftAddress        string
	grpcServer         *grpc.Server
	raft               *raft.Raft
	state              *State
	leadershipTakeover sync.Mutex
	healthyClients     []*shared.ServerNodeClient
	pendingClient      *shared.ServerNodeClient
	deadClients        []*shared.ServerNodeClient
}

func NewControlNode(grpcAddress, raftAddress string, nServerNodes, serverNodeStartPort int, raftDir string, peers []string, bootstrap bool) (*ControlNode, error) {
	c := ControlNode{
		id:          raftAddress,
		grpcAddress: grpcAddress,
		raftAddress: raftAddress,
		grpcServer:  grpc.NewServer(),
	}

	// Prepare server clients
	serverClients := make([]*shared.ServerNodeClient, 0, nServerNodes)
	for i := range nServerNodes {
		clientAddr := fmt.Sprintf("localhost:%v", serverNodeStartPort+i)
		conn, err := grpc.NewClient(clientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			for _, client := range serverClients {
				client.Conn.Close()
			}
			return nil, status.Errorf(codes.Internal, "failed to open client connection to %q: %v", clientAddr, err)
		}

		serverClients = append(serverClients, &shared.ServerNodeClient{
			Id:       strconv.Itoa(i + 1),
			Conn:     conn,
			Health:   grpc_health_v1.NewHealthClient(conn),
			Internal: api.NewInternalMessageBoardServiceClient(conn),
		})
	}
	c.deadClients = serverClients

	// Initialize Raft
	if err := c.setupRaft(raftDir, peers, bootstrap); err != nil {
		return nil, fmt.Errorf("failed to setup raft: %w", err)
	}

	// Register gRPC service
	api.RegisterControlPlaneServer(c.grpcServer, &c)
	reflection.Register(c.grpcServer)

	return &c, nil
}

func (c *ControlNode) setupRaft(raftDir string, peers []string, bootstrap bool) error {
	c.state = NewState(StateLog{
		HealthyNodes: c.clientsToNodeInfo(c.healthyClients),
		PendingNode:  c.clientToNodeInfo(c.pendingClient),
		DeadNodes:    c.clientsToNodeInfo(c.deadClients),
	})

	if err := os.MkdirAll(raftDir, 0755); err != nil {
		return fmt.Errorf("failed to create raft directory: %w", err)
	}

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.id)
	config.Logger = hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Info,
		Output:     os.Stdout,
		JSONFormat: true,
	})

	notifyCh := make(chan bool, 1)
	config.NotifyCh = notifyCh

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.db"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.db"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", c.raftAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}
	transport, err := raft.NewTCPTransport(c.raftAddress, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	r, err := raft.NewRaft(config, c.state, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return fmt.Errorf("failed to create raft: %w", err)
	}
	c.raft = r

	if bootstrap {
		servers := []raft.Server{
			{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			},
		}
		for _, peer := range peers {
			servers = append(servers, raft.Server{
				ID:      raft.ServerID(peer),
				Address: raft.ServerAddress(peer),
			})
		}

		configuration := raft.Configuration{
			Servers: servers,
		}
		c.raft.BootstrapCluster(configuration)
	}

	go c.observeLeadership(notifyCh)

	return nil
}

func (c *ControlNode) observeLeadership(ch <-chan bool) {
	for isLeader := range ch {
		c.leadershipTakeover.Lock()
		if isLeader {
			shared.Logger.Info("became leader, syncing local state from the log")

			allClients := make(map[string]*shared.ServerNodeClient)
			for _, client := range slices.Concat(c.healthyClients, c.deadClients) {
				allClients[client.Id] = client
			}
			if c.pendingClient != nil {
				allClients[c.pendingClient.Id] = c.pendingClient
			}

			c.healthyClients = make([]*shared.ServerNodeClient, len(c.state.HealthyNodes))
			for i, client := range c.state.HealthyNodes {
				c.healthyClients[i] = allClients[client.NodeId]
			}

			c.deadClients = make([]*shared.ServerNodeClient, len(c.state.DeadNodes))
			for i, client := range c.state.DeadNodes {
				c.deadClients[i] = allClients[client.NodeId]
			}

			if c.state.PendingNode == nil {
				c.pendingClient = nil
			} else {
				c.pendingClient = allClients[c.state.PendingNode.NodeId]
			}

			shared.Logger.Info("local state synced from the log")
		}
		c.leadershipTakeover.Unlock()
	}
}

func (c *ControlNode) applyStateChange(stateLog StateLog) error {
	data, err := json.Marshal(stateLog)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	future := c.raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to apply state change: %w", err)
	}

	return nil
}

func (c *ControlNode) isLeader() bool {
	return c.raft.State() == raft.Leader
}

func (c *ControlNode) Run() error {
	listener, err := net.Listen("tcp", c.grpcAddress)
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

		if c.raft != nil {
			if err := c.raft.Shutdown().Error(); err != nil {
				shared.Logger.Error("failed to shutdown raft", "error", err)
			}
		}

		c.grpcServer.GracefulStop()

		allClients := append(c.healthyClients, c.pendingClient)
		allClients = append(allClients, c.deadClients...)
		for _, client := range allClients {
			if client != nil {
				client.Conn.Close()
			}
		}
	}()

	shared.Logger.Info("control running", "address", c.grpcAddress)

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
			if !c.isLeader() {
				shared.Logger.InfoContext(ctx, "skipping health check - not leader")
				continue
			}

			c.leadershipTakeover.Lock()
			shared.Logger.InfoContext(ctx, "running health check loop")
			cycleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

			c.checkServersHealth(cycleCtx)
			c.linkServers(cycleCtx)
			c.repairServers(cycleCtx)

			stateLog := StateLog{
				HealthyNodes: c.clientsToNodeInfo(c.healthyClients),
				PendingNode:  c.clientToNodeInfo(c.pendingClient),
				DeadNodes:    c.clientsToNodeInfo(c.deadClients),
			}
			if err := c.applyStateChange(stateLog); err != nil {
				shared.Logger.ErrorContext(ctx, "failed to apply state change", "error", err)
			}
			c.leadershipTakeover.Unlock()

			cancel()
		}
	}
}

func (c *ControlNode) clientToNodeInfo(client *shared.ServerNodeClient) *api.NodeInfo {
	if client == nil {
		return nil
	}
	return &api.NodeInfo{
		NodeId:  client.Id,
		Address: client.Conn.Target(),
	}
}

func (c *ControlNode) clientsToNodeInfo(clients []*shared.ServerNodeClient) []*api.NodeInfo {
	nodes := make([]*api.NodeInfo, 0, len(clients))
	for _, client := range clients {
		node := c.clientToNodeInfo(client)
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (c *ControlNode) checkServersHealth(ctx context.Context) {
	var newHealthyClients []*shared.ServerNodeClient
	for _, client := range c.healthyClients {
		_, err := client.Health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})

		if err != nil {
			c.deadClients = append(c.deadClients, client)
			shared.Logger.WarnContext(ctx, "server node is dead", "address", client.Conn.Target())
		} else {
			newHealthyClients = append(newHealthyClients, client)
			shared.Logger.InfoContext(ctx, "server node is healthy", "address", client.Conn.Target())
		}
	}

	var newPendingClient *shared.ServerNodeClient
	if c.pendingClient != nil {
		res, err := c.pendingClient.Health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			c.deadClients = append(c.deadClients, c.pendingClient)
			shared.Logger.WarnContext(ctx, "pending server node is dead", "address", c.pendingClient.Conn.Target())
		} else if res.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			newHealthyClients = append(newHealthyClients, c.pendingClient)
			shared.Logger.InfoContext(ctx, "pending server node is ready", "address", c.pendingClient.Conn.Target())
		} else {
			newPendingClient = c.pendingClient
			shared.Logger.InfoContext(ctx, "pending server node is healthy", "address", c.pendingClient.Conn.Target())
		}
	}

	c.healthyClients = newHealthyClients
	c.pendingClient = newPendingClient
}

func (c *ControlNode) linkServers(ctx context.Context) {
	for i, client := range c.healthyClients {
		isHead := i == 0
		isTail := i == len(c.healthyClients)-1

		req := api.RewireRequest{
			UpstreamCount: int32(len(c.healthyClients) - i - 1),
		}
		if !isHead && !isTail {
			downstream := c.healthyClients[i-1]
			req.DownstreamId = &downstream.Id
			req.DownstreamAddress = shared.AnyPtr(downstream.Conn.Target())
			upstream := c.healthyClients[i+1]
			req.UpstreamId = &upstream.Id
			req.UpstreamAddress = shared.AnyPtr(upstream.Conn.Target())
		} else if isHead && !isTail {
			upstream := c.healthyClients[i+1]
			req.UpstreamId = &upstream.Id
			req.UpstreamAddress = shared.AnyPtr(upstream.Conn.Target())
		} else if !isHead && isTail {
			downstream := c.healthyClients[i-1]
			req.DownstreamId = &downstream.Id
			req.DownstreamAddress = shared.AnyPtr(downstream.Conn.Target())
		}

		shared.Logger.InfoContext(ctx, "rewiring server node", "address", client.Conn.Target())
		go client.Internal.Rewire(ctx, &req)
	}
}

func (c *ControlNode) repairServers(ctx context.Context) {
	if len(c.deadClients) > 0 && c.pendingClient == nil {
		c.pendingClient = c.deadClients[0]
		c.deadClients = c.deadClients[1:]

		go func() {
			args := []string{"--address", c.pendingClient.Conn.Target()}
			if len(c.healthyClients) > 0 {
				downstream := c.healthyClients[len(c.healthyClients)-1]
				args = append(args, "--downstream-id", downstream.Id, "--downstream-address", downstream.Conn.Target())
			}

			cmdStr := fmt.Sprintf("%s %s > %s 2>&1",
				filepath.Join("build", "server"),
				strings.Join(args, " "),
				fmt.Sprintf("./logs/%s.jsonl", c.pendingClient.Id),
			)

			cmd := exec.Command("sh", "-c", cmdStr)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			if err := cmd.Start(); err != nil {
				shared.Logger.ErrorContext(ctx, "failed to start server node", "address", c.pendingClient.Conn.Target(), "error", err)
			} else {
				shared.Logger.InfoContext(ctx, "starting server node", "address", c.pendingClient.Conn.Target())
			}
		}()
	}
}
