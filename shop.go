package shop

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	config "github.com/comfforts/comff-config"
	api "github.com/comfforts/comff-shops/api/v1"
	"github.com/comfforts/logger"
)

const DEFAULT_SERVICE_PORT = "52051"
const DEFAULT_SERVICE_HOST = "127.0.0.1"

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

var (
	defaultDialTimeout      = 5 * time.Second
	defaultKeepAlive        = 30 * time.Second
	defaultKeepAliveTimeout = 10 * time.Second
)

const ShopClientContextKey = ContextKey("shop-client")

type ClientOption struct {
	DialTimeout      time.Duration
	KeepAlive        time.Duration
	KeepAliveTimeout time.Duration
	Caller           string
}

type Client interface {
	AddShop(ctx context.Context, req *api.AddShopRequest, opts ...grpc.CallOption) (*api.AddShopResponse, error)
	GetShop(ctx context.Context, req *api.GetShopRequest, opts ...grpc.CallOption) (*api.GetShopResponse, error)
	DeleteShop(ctx context.Context, req *api.DeleteShopRequest, opts ...grpc.CallOption) (*api.DeleteResponse, error)
	GetShops(ctx context.Context, req *api.SearchShopsRequest, opts ...grpc.CallOption) (*api.SearchShopsResponse, error)
	Close() error
}

func NewDefaultClientOption() *ClientOption {
	return &ClientOption{
		DialTimeout:      defaultDialTimeout,
		KeepAlive:        defaultKeepAlive,
		KeepAliveTimeout: defaultKeepAliveTimeout,
	}
}

type shopClient struct {
	logger logger.AppLogger
	client api.ShopsClient
	conn   *grpc.ClientConn
	opts   *ClientOption
}

func NewClient(logger logger.AppLogger, clientOpts *ClientOption) (*shopClient, error) {
	tlsConfig, err := config.SetupTLSConfig(&config.ConfigOpts{Target: config.SHOP_CLIENT})
	if err != nil {
		logger.Error("error setting shops client TLS", zap.Error(err))
		return nil, err
	}
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}

	servicePort := os.Getenv("SHOP_SERVICE_PORT")
	if servicePort == "" {
		servicePort = DEFAULT_SERVICE_PORT
	}
	serviceHost := os.Getenv("SHOP_SERVICE_HOST")
	if serviceHost == "" {
		serviceHost = DEFAULT_SERVICE_HOST
	}

	serviceAddr := fmt.Sprintf("%s:%s", serviceHost, servicePort)
	// with load balancer
	// serviceAddr = fmt.Sprintf("%s:///%s", loadbalance.ShopResolverName, serviceAddr)
	// serviceAddr = fmt.Sprintf("%s:///%s", "shops", serviceAddr)

	conn, err := grpc.Dial(serviceAddr, opts...)
	if err != nil {
		logger.Error("client failed to connect", zap.Error(err))
		return nil, err
	}

	client := api.NewShopsClient(conn)

	return &shopClient{
		client: client,
		logger: logger,
		conn:   conn,
		opts:   clientOpts,
	}, nil
}

func (sc *shopClient) AddShop(ctx context.Context, req *api.AddShopRequest, opts ...grpc.CallOption) (*api.AddShopResponse, error) {
	ctx, cancel := sc.contextWithOptions(ctx, sc.opts)
	defer cancel()

	resp, err := sc.client.AddShop(ctx, req)
	if err != nil {
		sc.logger.Error("error adding shop", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (sc *shopClient) GetShop(ctx context.Context, req *api.GetShopRequest, opts ...grpc.CallOption) (*api.GetShopResponse, error) {
	ctx, cancel := sc.contextWithOptions(ctx, sc.opts)
	defer cancel()

	resp, err := sc.client.GetShop(ctx, req)
	if err != nil {
		sc.logger.Error("error getting shop", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (sc *shopClient) DeleteShop(ctx context.Context, req *api.DeleteShopRequest, opts ...grpc.CallOption) (*api.DeleteResponse, error) {
	ctx, cancel := sc.contextWithOptions(ctx, sc.opts)
	defer cancel()

	resp, err := sc.client.DeleteShop(ctx, req)
	if err != nil {
		sc.logger.Error("error deleting shop", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (sc *shopClient) GetShops(ctx context.Context, req *api.SearchShopsRequest, opts ...grpc.CallOption) (*api.SearchShopsResponse, error) {
	ctx, cancel := sc.contextWithOptions(ctx, sc.opts)
	defer cancel()

	resp, err := sc.client.SearchShops(ctx, req)
	if err != nil {
		sc.logger.Error("error getting shops", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (sc *shopClient) Close() error {
	if err := sc.conn.Close(); err != nil {
		sc.logger.Error("error closing shop client connection", zap.Error(err))
		return err
	}
	return nil
}

func (sc *shopClient) contextWithOptions(ctx context.Context, opts *ClientOption) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, sc.opts.DialTimeout)
	if sc.opts.Caller != "" {
		md := metadata.New(map[string]string{"service-client": sc.opts.Caller})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx, cancel
}
