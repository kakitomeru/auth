package api

import (
	"context"
	"errors"
	"regexp"

	"github.com/kakitomeru/auth/internal/interceptor"
	"github.com/kakitomeru/auth/internal/metric"
	"github.com/kakitomeru/auth/internal/repository"
	"github.com/kakitomeru/auth/internal/service"
	pb "github.com/kakitomeru/auth/pkg/pb/v1"
	"github.com/kakitomeru/shared/jwt"
	s "github.com/kakitomeru/shared/status"
	"github.com/kakitomeru/shared/telemetry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	emailRegex         = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	StatusInvalidEmail = status.Error(
		codes.InvalidArgument,
		"email is required and must be valid",
	)
)

type AuthServiceHandler struct {
	pb.UnimplementedAuthServiceServer
	service *service.Service
}

func NewAuthServiceHandler(service *service.Service) *AuthServiceHandler {
	return &AuthServiceHandler{
		service: service,
	}
}

func (h *AuthServiceHandler) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthHandler.GetUser")
	defer span.End()

	userID := interceptor.GetUserID(ctx)
	user, err := h.service.Auth.GetUser(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, jwt.ErrInvalidToken):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, jwt.ErrExpiredToken):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, s.StatusInternal
		}
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			UserId:   user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (h *AuthServiceHandler) Login(
	ctx context.Context,
	req *pb.LoginRequest,
) (*pb.LoginResponse, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthHandler.Login")
	defer span.End()

	if req.Email == "" || !emailRegex.MatchString(req.Email) || len(req.Email) > 254 {
		telemetry.RecordError(span, errors.New("email is required and must be valid"))
		return nil, StatusInvalidEmail
	}

	if req.Password == "" || len(req.Password) < 8 || len(req.Password) > 20 {
		errMsg := "password is required and must be between 8 and 20 characters long"
		telemetry.RecordError(span, errors.New(errMsg))
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	user, tokenPair, err := h.service.Auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		metric.FailedLoginsCounter.Add(ctx, 1)
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, service.ErrIncorrectPassword):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, s.StatusInternal
		}
	}

	metric.ActiveSessionsCounter.Add(ctx, 1)
	metric.SuccessfulLoginsCounter.Add(ctx, 1)
	return &pb.LoginResponse{
		User: &pb.User{
			UserId:   user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (h *AuthServiceHandler) RefreshToken(
	ctx context.Context,
	req *pb.RefreshTokenRequest,
) (*pb.RefreshTokenResponse, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthHandler.RefreshToken")
	defer span.End()

	if req.RefreshToken == "" {
		errMsg := "refresh token is required"
		telemetry.RecordError(span, errors.New(errMsg))
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	accessToken, err := h.service.Auth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, service.ErrSessionExpired):
			metric.ActiveSessionsCounter.Add(ctx, -1)
			return nil, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, repository.ErrSessionNotFound):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, s.StatusInternal
		}
	}

	return &pb.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (h *AuthServiceHandler) Register(
	ctx context.Context,
	req *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthHandler.Register")
	defer span.End()

	if req.Username == "" || len(req.Username) < 3 || len(req.Username) > 20 {
		errMsg := "username is required and must be between 3 and 20 characters long"
		telemetry.RecordError(span, errors.New(errMsg))
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	if req.Email == "" || !emailRegex.MatchString(req.Email) || len(req.Email) > 254 {
		telemetry.RecordError(span, errors.New("email is required and must be valid"))
		return nil, StatusInvalidEmail
	}

	if req.Password == "" || len(req.Password) < 8 || len(req.Password) > 20 {
		errMsg := "password is required and must be between 8 and 20 characters long"
		telemetry.RecordError(span, errors.New(errMsg))
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	id, err := h.service.Auth.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, service.ErrInvalidPassword):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, repository.ErrUserAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, s.StatusInternal
		}
	}

	metric.UserRegistrationsCounter.Add(ctx, 1)
	return &pb.RegisterResponse{
		UserId: id.String(),
	}, nil
}

func (h *AuthServiceHandler) Logout(
	ctx context.Context,
	req *pb.LogoutRequest,
) (*pb.LogoutResponse, error) {
	ctx, span := telemetry.StartSpan(ctx, "AuthHandler.Logout")
	defer span.End()

	userID := interceptor.GetUserID(ctx)

	err := h.service.Auth.Logout(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, jwt.ErrExpiredToken):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, s.StatusInternal
		}
	}

	metric.ActiveSessionsCounter.Add(ctx, -1)
	return &pb.LogoutResponse{}, nil
}
