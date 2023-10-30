package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	cfg "github.com/PrahaTurbo/url-shortener/config"
	"github.com/PrahaTurbo/url-shortener/internal/app"
	"github.com/PrahaTurbo/url-shortener/internal/auth"
	"github.com/PrahaTurbo/url-shortener/internal/grpcapp"
	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	"github.com/PrahaTurbo/url-shortener/internal/storage/provider"
	pb "github.com/PrahaTurbo/url-shortener/proto"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	c := cfg.Load()
	lgr, err := logger.Initialize(c.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	store, err := provider.NewStorage(c.DatabaseDSN, c.StorageFilePath, lgr)
	if err != nil {
		log.Fatal(err)
	}

	srvc := service.NewService(c.BaseURL, store, lgr)
	auth := auth.NewAuth(c.JWTSecret, c.TrustedSubnet)

	application := app.NewApp(c.Addr, srvc, lgr, auth)
	httpServer := http.Server{
		Addr:    application.Addr(),
		Handler: application.Router(),
	}

	grpcApplication := grpcapp.NewgRPCShortener(srvc, lgr)
	listener, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(auth.UnaryServerInterceptor))
	pb.RegisterURLShortenerServer(grpcServer, grpcApplication)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigint

		grpcServer.GracefulStop()

		if err := httpServer.Shutdown(context.Background()); err != nil {
			lgr.Error("HTTP server shutdown error", zap.Error(err))
		}

		close(idleConnsClosed)
	}()

	go func() {
		lgr.Info("gRPC server is running")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	lgr.Info("HTTP server is running", zap.String("address", application.Addr()))

	if c.EnableHTTPS {
		if err := httpServer.ListenAndServeTLS(
			"cmd/shortener/cert.pem",
			"cmd/shortener/key.pem",
		); !errors.Is(err, http.ErrServerClosed) {
			lgr.Fatal("HTTP server error", zap.Error(err))
		}
	} else {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			lgr.Fatal("HTTP server error", zap.Error(err))
		}
	}

	<-idleConnsClosed
	lgr.Info("HTTP and gRPC servers shutdown gracefully")
}
