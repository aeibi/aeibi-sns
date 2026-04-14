package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"aeibi/api"
	"aeibi/internal/auth"
	"aeibi/internal/config"
	"aeibi/internal/controller"
	"aeibi/internal/repository/oss"
	searchrepo "aeibi/internal/repository/search"
	"aeibi/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"google.golang.org/grpc"
)

// StartGRPCServer starts the gRPC server and returns it plus an error channel.
func StartGRPCServer(ctx context.Context, cfg *config.Config, dbPool *pgxpool.Pool, ossClient *oss.OSS, searchRepo *searchrepo.Search, riverClient *river.Client[pgx.Tx]) (*grpc.Server, <-chan error, error) {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.NewAuthUnaryServerInterceptor(cfg.Auth.JWTSecret)))

	userSvc := service.NewUserService(dbPool, ossClient, searchRepo, cfg, riverClient)
	followSvc := service.NewFollowService(dbPool, riverClient)
	postSvc := service.NewPostService(dbPool, ossClient, searchRepo, riverClient)
	fileSvc := service.NewFileService(dbPool, ossClient, cfg.OSS.MaxUploadSizeKB)
	commentSvc := service.NewCommentService(dbPool, riverClient)
	messageSvc := service.NewMessageService(dbPool)
	reportSvc := service.NewReportService(dbPool)

	userHandler := controller.NewUserHandler(userSvc)
	followHandler := controller.NewFollowHandler(followSvc)
	postHandler := controller.NewPostHandler(postSvc)
	fileHandler := controller.NewFileHandler(fileSvc)
	commentHandler := controller.NewCommentHandler(commentSvc)
	messageHandler := controller.NewMessageHandler(messageSvc)
	reportHandler := controller.NewReportHandler(reportSvc)

	api.RegisterUserServiceServer(grpcServer, userHandler)
	api.RegisterFollowServiceServer(grpcServer, followHandler)
	api.RegisterPostServiceServer(grpcServer, postHandler)
	api.RegisterFileServiceServer(grpcServer, fileHandler)
	api.RegisterCommentServiceServer(grpcServer, commentHandler)
	api.RegisterMessageServiceServer(grpcServer, messageHandler)
	api.RegisterReportServiceServer(grpcServer, reportHandler)

	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("listen gRPC: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- err
		}
	}()

	return grpcServer, errCh, nil
}
