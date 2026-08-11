package grpcclients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "github.com/Yuvraj02/md-protos/proto/auth/v1"
	blogv1 "github.com/Yuvraj02/md-protos/proto/blog/v1"
)

// Clients holds outbound gRPC clients used by the gateway.
type Clients struct {
	Auth       authv1.AuthServiceClient
	Blog       blogv1.BlogServiceClient
	authConn   *grpc.ClientConn
	blogConn   *grpc.ClientConn
}

// Dial connects to Auth and Blog gRPC services.
func Dial(authAddr, blogAddr string) (*Clients, error) {
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial auth: %w", err)
	}
	blogConn, err := grpc.NewClient(blogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("dial blog: %w", err)
	}

	return &Clients{
		Auth:     authv1.NewAuthServiceClient(authConn),
		Blog:     blogv1.NewBlogServiceClient(blogConn),
		authConn: authConn,
		blogConn: blogConn,
	}, nil
}

// Close closes client connections.
func (c *Clients) Close() error {
	var first error
	if c.authConn != nil {
		if err := c.authConn.Close(); err != nil && first == nil {
			first = err
		}
	}
	if c.blogConn != nil {
		if err := c.blogConn.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
