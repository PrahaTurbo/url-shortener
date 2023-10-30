package grpcapp

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/PrahaTurbo/url-shortener/internal/logger"
	"github.com/PrahaTurbo/url-shortener/internal/service"
	pb "github.com/PrahaTurbo/url-shortener/proto"
)

type Application struct {
	pb.UnimplementedURLShortenerServer

	srvc service.Service
	log  *logger.Logger
}

func NewGRPCApp(srvc service.Service, logger *logger.Logger) *Application {
	return &Application{
		srvc: srvc,
		log:  logger,
	}
}

func (a *Application) MakeURL(ctx context.Context, in *pb.MakeURLRequest) (*pb.MakeURLResponse, error) {
	url, err := a.srvc.SaveURL(ctx, in.Url)
	if err != nil {
		a.log.Error("error while saving url", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.MakeURLResponse{Result: url}

	return &response, nil
}

func (a *Application) GetOriginalURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	originalURL, err := a.srvc.GetURL(ctx, in.ShortUrl)
	if err != nil {
		a.log.Error("error while getting original url", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.GetURLResponse{OriginalUrl: originalURL}

	return &response, nil
}

func (a *Application) GetUserURLs(ctx context.Context, in *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	urls, err := a.srvc.GetURLsByUserID(ctx)
	if err != nil {
		a.log.Error("error getting user urls", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	pbURLs := make([]*pb.UserURLsResponse_UserURLs, len(urls))

	for i := range urls {
		url := &pb.UserURLsResponse_UserURLs{
			ShortUrl:    urls[i].ShortURL,
			OriginalUrl: urls[i].OriginalURL,
		}

		pbURLs[i] = url
	}

	response := pb.UserURLsResponse{UserUrls: pbURLs}

	return &response, nil
}

func (a *Application) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	if err := a.srvc.DeleteURLs(ctx, in.Urls); err != nil {
		a.log.Error("error accepting urls for deletion", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.DeleteURLsResponse{Status: pb.DeleteURLsResponse_ACCEPTED}

	return &response, nil
}

func (a *Application) PingDB(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	var response pb.PingResponse

	err := a.srvc.PingDB()
	switch {
	case err != nil:
		response.Status = pb.PingResponse_INACTIVE
	default:
		response.Status = pb.PingResponse_ACTIVE
	}

	return &response, nil
}

func (a *Application) GetStats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats, err := a.srvc.GetStats(ctx)
	if err != nil {
		a.log.Error("error getting stats", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.StatsResponse{
		Urls:  int64(stats.URLs),
		Users: int64(stats.Users),
	}

	return &response, nil
}
