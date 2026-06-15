package grpcserver

import (
	"context"

	shortifyv1 "shortify/server/api/gen/shortify/v1"
	"shortify/server/internal/domain"
	"shortify/server/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	shortifyv1.UnimplementedAuthServiceServer
	auth *service.AuthService
}

func NewAuthServer(auth *service.AuthService) *AuthServer {
	return &AuthServer{auth: auth}
}

func (s *AuthServer) Register(ctx context.Context, req *shortifyv1.RegisterRequest) (*shortifyv1.AuthResponse, error) {
	token, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &shortifyv1.AuthResponse{Token: token}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *shortifyv1.LoginRequest) (*shortifyv1.AuthResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &shortifyv1.AuthResponse{Token: token}, nil
}

func mapError(err error) error {
	switch err {
	case domain.ErrInvalidInput:
		return status.Error(codes.InvalidArgument, "invalid input")
	case domain.ErrInvalidEmail:
		return status.Error(codes.InvalidArgument, "invalid email format")
	case domain.ErrPasswordTooShort:
		return status.Error(codes.InvalidArgument, "А что ещё у тебя такое же короткое как этот пароль?")
	case domain.ErrProfanity:
		return status.Error(codes.InvalidArgument, "Фу как некультурно")
	case domain.ErrInvalidURL:
		return status.Error(codes.InvalidArgument, "invalid url")
	case domain.ErrInvalidCreds:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case domain.ErrForbidden:
		return status.Error(codes.PermissionDenied, "forbidden")
	case domain.ErrNotFound:
		return status.Error(codes.NotFound, "not found")
	case domain.ErrAlreadyExists:
		return status.Error(codes.AlreadyExists, "already exists")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
