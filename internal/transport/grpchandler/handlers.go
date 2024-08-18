package grpchandler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/lks-go/url-shortener/internal/entity"
	"github.com/lks-go/url-shortener/internal/service"
	"github.com/lks-go/url-shortener/pkg/proto"
)

// Service это интерфейс сервиса отвечающего за обратоку входящих http запросов
type Service interface {
	MakeBatchShortURL(ctx context.Context, userID string, urls []service.URL) ([]service.URL, error)
	MakeShortURL(ctx context.Context, userID, url string) (string, error)
	URL(ctx context.Context, id string) (string, error)
	UsersURLs(ctx context.Context, userID string) ([]service.UsersURL, error)
	Stats(ctx context.Context) (*service.StatsInfo, error)
}

// Deleter это интерфейс сервиса отвечающего за получение запроса на удаление
type Deleter interface {
	Delete(ctx context.Context, userID string, codes []string) error
}

type Config struct {
	RedirectBasePath string
	TrustedSubnet    string
}

type Deps struct {
	Service Service
	Deleter Deleter
}

func New(cfg Config, d *Deps) (*Handler, error) {
	var ipNet *net.IPNet
	var err error

	if cfg.TrustedSubnet != "" {
		_, ipNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
		}
	}

	return &Handler{
		redirectBasePath: cfg.RedirectBasePath,
		service:          d.Service,
		deleter:          d.Deleter,
		ipNet:            ipNet,
	}, nil
}

type Handler struct {
	redirectBasePath string
	service          Service
	deleter          Deleter
	ipNet            *net.IPNet

	proto.UnimplementedURLShortenerServer
}

func (h *Handler) ShortURL(ctx context.Context, request *proto.ShortURLRequest) (*proto.ShortURLResponse, error) {
	userID, err := outgoingMetaData(ctx, entity.UserIDHeaderName)
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	id, err := h.service.MakeShortURL(ctx, userID[0], request.Url)
	if err != nil && !errors.Is(err, service.ErrURLAlreadyExists) {
		logrus.Errorf("failed to make short url: %s", err)
		return nil, status.Error(codes.Internal, (codes.Internal).String())
	}

	if errors.Is(err, service.ErrURLAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, (codes.AlreadyExists).String())
	}

	return &proto.ShortURLResponse{ShortenUrl: fmt.Sprintf("%s/%s", h.redirectBasePath, id)}, nil
}

func (h *Handler) Redirect(ctx context.Context, request *proto.RedirectRequest) (*proto.RedirectResponse, error) {
	parsedURL, err := url.Parse(request.ShortenUrl)
	if err != nil {
		logrus.Errorf("failed to parse url: %s", request.ShortenUrl)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	matches := regexp.MustCompile(`/(\w+)`).FindStringSubmatch(parsedURL.Path)
	if len(matches) < 1 {
		logrus.Errorf("failed to compile shorten url: %s", request.ShortenUrl)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	code := matches[1]
	url, err := h.service.URL(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			return nil, status.Error(codes.NotFound, (codes.NotFound).String())
		case errors.Is(err, service.ErrDeleted):
			return nil, status.Error(codes.NotFound, (codes.NotFound).String())
		default:
			logrus.Errorf("failed to get url by code [%s]: %s", code, err)
			return nil, status.Error(codes.Internal, (codes.Internal).String())
		}
	}

	return &proto.RedirectResponse{Url: url}, nil
}

func (h *Handler) ShortenURL(ctx context.Context, request *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID, err := outgoingMetaData(ctx, entity.UserIDHeaderName)
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	id, err := h.service.MakeShortURL(ctx, userID[0], request.Url)
	if err != nil && !errors.Is(err, service.ErrURLAlreadyExists) {
		logrus.Errorf("failed to make short url: %s", err)
		return nil, status.Error(codes.Internal, (codes.Internal).String())
	}

	if errors.Is(err, service.ErrURLAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, (codes.AlreadyExists).String())
	}

	return &proto.ShortenURLResponse{Result: fmt.Sprintf("%s/%s", h.redirectBasePath, id)}, nil
}

func (h *Handler) ShortenBatchURL(ctx context.Context, request *proto.ShortenBatchURLRequest) (*proto.ShortenBatchURLResponse, error) {
	userID, err := outgoingMetaData(ctx, entity.UserIDHeaderName)
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	urlList := make([]service.URL, 0, len(request.Urls))
	for _, u := range request.Urls {
		urlList = append(urlList, service.URL{
			СorrelationID: u.CorrelationId,
			OriginalURL:   u.OriginalUrl,
		})
	}

	shortURLList, err := h.service.MakeBatchShortURL(ctx, userID[0], urlList)
	if err != nil {
		logrus.Errorf("failed to make batch short urls: %s", err)
		return nil, status.Error(codes.Internal, (codes.Internal).String())
	}

	urls := make([]*proto.ShortenBatchURLResponse_URL, 0, len(shortURLList))
	for _, u := range shortURLList {
		urls = append(urls, &proto.ShortenBatchURLResponse_URL{
			CorrelationId: u.СorrelationID,
			ShortUrl:      u.OriginalURL,
		})
	}

	return &proto.ShortenBatchURLResponse{Urls: urls}, nil
}

func (h *Handler) UsersURLs(ctx context.Context, request *proto.UsersURLsRequest) (*proto.UsersURLsResponse, error) {
	userID, err := outgoingMetaData(ctx, entity.UserIDHeaderName)
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	usersUrls, err := h.service.UsersURLs(ctx, userID[0])
	if err != nil {
		logrus.Errorf("failed to get users urls: %s", err)
		return nil, status.Error(codes.Internal, (codes.Internal).String())
	}

	urls := make([]*proto.UsersURLsResponse_URL, 0, len(usersUrls))

	return &proto.UsersURLsResponse{Urls: urls}, nil
}

func (h *Handler) Delete(ctx context.Context, request *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	userID, err := outgoingMetaData(ctx, entity.UserIDHeaderName)
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	go func() {
		if err := h.deleter.Delete(ctx, userID[0], request.Codes); err != nil {
			logrus.Errorf("failed to delete urls (userId = %s, codes = [%v]): %s", userID, request.Codes, err)
		}
	}()

	return &proto.DeleteResponse{}, nil
}

func (h *Handler) Stats(ctx context.Context, _ *proto.StatsRequest) (*proto.StatsResponse, error) {
	ips, err := incomingMetaData(ctx, "X-Real-IP")
	if err != nil {
		logrus.Errorf("failed to get metadata: %s", err)
		return nil, status.Error(codes.InvalidArgument, (codes.InvalidArgument).String())
	}

	ip := ips[0]
	if h.ipNet != nil && !h.ipNet.Contains(net.ParseIP(ip)) {
		logrus.Errorf("ip %s is not in trusted subnet", ip)
		return nil, status.Error(codes.PermissionDenied, (codes.PermissionDenied).String())

	}

	statsInfo, err := h.service.Stats(ctx)
	if err != nil {
		logrus.Errorf("failed to get stats in handler: %s", err)
		return nil, status.Error(codes.Internal, (codes.Internal).String())
	}

	return &proto.StatsResponse{Urls: int64(statsInfo.URLCount), Users: int64(statsInfo.UserCount)}, nil
}

func outgoingMetaData(ctx context.Context, key string) ([]string, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	value := md.Get(key)
	if len(value) == 0 {
		return nil, fmt.Errorf("%s not supplied", key)
	}

	return value, nil
}

func incomingMetaData(ctx context.Context, key string) ([]string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	value := md.Get(key)
	if len(value) == 0 {
		return nil, fmt.Errorf("%s not supplied", key)
	}

	return value, nil
}
