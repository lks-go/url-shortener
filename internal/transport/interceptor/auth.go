package interceptor

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/lks-go/url-shortener/internal/entity"
	"github.com/lks-go/url-shortener/internal/lib/jwt"
)

func Auth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	var userID string
	var claims *jwt.Claims
	var err error

	token, ok := md[entity.AuthTokenHeader]
	if ok {
		claims, err = jwt.ParseJWTToken(token[0])
		if err != nil && !errors.Is(err, jwt.ErrInvalidToken) && !errors.Is(err, jwt.ErrTokenExpired) {
			log.Println("failed to parse jwt:", err)
			return nil, status.Error(codes.InvalidArgument, "failed to parse jwt")
		}

		if claims != nil && claims.UserID == "" {
			return nil, status.Error(codes.Unauthenticated, (codes.Unauthenticated).String())
		}

		if claims != nil {
			userID = claims.UserID
		}
	}

	if !ok || errors.Is(err, jwt.ErrInvalidToken) || errors.Is(err, jwt.ErrTokenExpired) {
		userID = uuid.NewString()

		token, err := jwt.BuildNewJWTToken(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, (codes.Internal).String())
		}

		// TODO проверить работает ли это
		header := metadata.Pairs(entity.AuthTokenHeader, token)
		grpc.SendHeader(ctx, header)
	}

	ctx = metadata.AppendToOutgoingContext(ctx, entity.UserIDHeaderName, userID)
	return handler(ctx, req)
}
