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

type URLShortener struct {
	pb.UnimplementedURLShortenerServer

	srvc service.Service
	log  *logger.Logger
}

func NewgRPCShortener(srvc service.Service, logger *logger.Logger) *URLShortener {
	return &URLShortener{
		srvc: srvc,
		log:  logger,
	}
}

func (u *URLShortener) MakeURL(ctx context.Context, in *pb.MakeURLRequest) (*pb.MakeURLResponse, error) {
	url, err := u.srvc.SaveURL(ctx, in.Url)
	if err != nil {
		u.log.Error("error while saving url", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.MakeURLResponse{Result: url}

	return &response, nil
}

func (u *URLShortener) GetOriginalURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	originalURL, err := u.srvc.GetURL(ctx, in.ShortUrl)
	if err != nil {
		u.log.Error("error while getting original url", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.GetURLResponse{OriginalUrl: originalURL}

	return &response, nil
}

func (u *URLShortener) GetUserURLs(ctx context.Context, in *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	urls, err := u.srvc.GetURLsByUserID(ctx)
	if err != nil {
		u.log.Error("error getting user urls", zap.Error(err))

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

func (u *URLShortener) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	if err := u.srvc.DeleteURLs(ctx, in.Urls); err != nil {
		u.log.Error("error accepting urls for deletion", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.DeleteURLsResponse{Status: pb.DeleteURLsResponse_ACCEPTED}

	return &response, nil
}

func (u *URLShortener) PingDB(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	var response pb.PingResponse

	err := u.srvc.PingDB()
	switch {
	case err != nil:
		response.Status = pb.PingResponse_INACTIVE
	default:
		response.Status = pb.PingResponse_ACTIVE
	}

	return &response, nil
}

func (u *URLShortener) GetStats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats, err := u.srvc.GetStats(ctx)
	if err != nil {
		u.log.Error("error getting stats", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response := pb.StatsResponse{
		Urls:  int64(stats.URLs),
		Users: int64(stats.Users),
	}

	return &response, nil
}
