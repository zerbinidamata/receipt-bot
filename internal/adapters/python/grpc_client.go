package python

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"receipt-bot/internal/adapters/python/pb"
)

// GRPCClient manages the connection to the Python gRPC service
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.ScraperServiceClient
}

// NewGRPCClient creates a new gRPC client connection to the Python service
func NewGRPCClient(address string, timeout time.Duration) (*GRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Dial the server
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Python service at %s: %w", address, err)
	}

	// Use the generated client
	client := pb.NewScraperServiceClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ScrapeContent calls the Python service to scrape content
func (c *GRPCClient) ScrapeContent(ctx context.Context, req *pb.ScrapeRequest) (*pb.ScrapeResponse, error) {
	return c.client.ScrapeContent(ctx, req)
}

// Re-export types from the generated package for convenience
type ScrapeRequest = pb.ScrapeRequest
type ScrapeResponse = pb.ScrapeResponse
