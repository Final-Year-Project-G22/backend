package coregrpc

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	coreuserv1 "github.com/Final-Year-Project-G22/backend/core/pb/core/user/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Module("core-grpc",
	fx.Provide(NewUserProfileService),
	fx.Invoke(registerGrpcServer),
)

func registerGrpcServer(
	lc fx.Lifecycle,
	cfg *core.Config,
	log core.Logger,
	service *UserProfileService,
) {
	var (
		server   *grpc.Server
		listener net.Listener
	)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			addr := net.JoinHostPort("", strconv.Itoa(cfg.App.GRPCPort))

			l, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			listener = l

			server = grpc.NewServer()
			coreuserv1.RegisterCoreUserServiceServer(server, service)

			go func() {
				if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
					log.Error("gRPC server stopped with error", core.Error(serveErr))
				}
			}()

			log.Info("gRPC server started", core.String("addr", addr))

			return nil
		},
		OnStop: func(_ context.Context) error {
			if server != nil {
				server.GracefulStop()
			}
			if listener != nil {
				_ = listener.Close()
			}
			log.Info("gRPC server stopped")
			return nil
		},
	})
}
