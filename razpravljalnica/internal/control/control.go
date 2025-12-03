package control

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Control struct {
	api.UnimplementedControlPlaneServer
	address      string
	headAddress  string
	headID       string
	tailAddress  string
	tailID       string
	grpcServer   *grpc.Server
	healthServer *health.Server
}

var _ api.ControlPlaneServer = (*Control)(nil)

func NewControl(address, headAddress, headID, tailAddress, tailID string) *Control {
	control := &Control{
		address:      address,
		headAddress:  headAddress,
		headID:       headID,
		tailAddress:  tailAddress,
		tailID:       tailID,
		grpcServer:   grpc.NewServer(),
		healthServer: health.NewServer(),
	}

	api.RegisterControlPlaneServer(control.grpcServer, control)
	grpc_health_v1.RegisterHealthServer(control.grpcServer, control.healthServer)
	reflection.Register(control.grpcServer)

	return control
}

func (c *Control) Run() error {
	listener, err := net.Listen("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		shared.Logger.Info("received shutdown signal", "signal", sig.String())
		c.grpcServer.GracefulStop()
	}()

	shared.Logger.Info("control running", "address", c.address)

	if err := c.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("control stopped gracefully")
	return nil
}
