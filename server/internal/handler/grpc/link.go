package grpcserver

import (
	"context"

	shortifyv1 "shortify/server/api/gen/shortify/v1"
	"shortify/server/internal/domain"
	"shortify/server/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LinkServer struct {
	shortifyv1.UnimplementedLinkServiceServer
	links *service.LinkService
}

func NewLinkServer(links *service.LinkService) *LinkServer {
	return &LinkServer{links: links}
}

func (s *LinkServer) CreateLink(ctx context.Context, req *shortifyv1.CreateLinkRequest) (*shortifyv1.Link, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	link, err := s.links.Create(ctx, userID, req.GetUrl())
	if err != nil {
		return nil, mapError(err)
	}

	return toProtoLink(link), nil
}

func (s *LinkServer) ListLinks(ctx context.Context, _ *shortifyv1.ListLinksRequest) (*shortifyv1.ListLinksResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	links, err := s.links.List(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	result := make([]*shortifyv1.Link, 0, len(links))
	for i := range links {
		result = append(result, toProtoLink(&links[i]))
	}

	return &shortifyv1.ListLinksResponse{Links: result}, nil
}

func (s *LinkServer) DeleteLink(ctx context.Context, req *shortifyv1.DeleteLinkRequest) (*shortifyv1.DeleteLinkResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	linkID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.links.Delete(ctx, userID, linkID); err != nil {
		return nil, mapError(err)
	}

	return &shortifyv1.DeleteLinkResponse{}, nil
}

func (s *LinkServer) GetLinkStats(ctx context.Context, req *shortifyv1.GetLinkStatsRequest) (*shortifyv1.LinkStats, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	linkID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	total, recent, err := s.links.GetStats(ctx, userID, linkID)
	if err != nil {
		return nil, mapError(err)
	}

	return &shortifyv1.LinkStats{
		TotalClicks: total,
		Recent:      toProtoClicks(recent),
	}, nil
}

func (s *LinkServer) ResolveLink(ctx context.Context, req *shortifyv1.ResolveLinkRequest) (*shortifyv1.ResolveLinkResponse, error) {
	originalURL, err := s.links.Resolve(ctx, req.GetCode(), "grpc", "grpc-client")
	if err != nil {
		return nil, mapError(err)
	}

	return &shortifyv1.ResolveLinkResponse{
		OriginalUrl: originalURL,
	}, nil
}

func toProtoLink(link *service.LinkView) *shortifyv1.Link {
	return &shortifyv1.Link{
		Id:          link.ID.String(),
		OriginalUrl: link.OriginalURL,
		ShortCode:   link.ShortCode,
		ShortUrl:    link.ShortURL,
		ClickCount:  link.ClicksCount,
		CreatedAt:   link.CreatedAt,
	}
}

func toProtoClicks(clicks []domain.Click) []*shortifyv1.Click {
	result := make([]*shortifyv1.Click, 0, len(clicks))
	for _, click := range clicks {
		result = append(result, &shortifyv1.Click{
			Id:        click.ID.String(),
			LinkId:    click.LinkID.String(),
			Ip:        click.IP,
			UserAgent: click.UserAgent,
			CreatedAt: click.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return result
}
