package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCSentimentClient wraps a gRPC connection to the sentiment service.
// Falls back to HTTP if gRPC is unavailable.
type GRPCSentimentClient struct {
	grpcAddr   string
	httpClient *SentimentClient // HTTP fallback
	conn       *grpc.ClientConn
}

// NewGRPCSentimentClient creates a gRPC client with HTTP fallback.
//
//	grpcAddr:  "sentiment:50051"
//	httpURL:   "http://sentiment:8000"
func NewGRPCSentimentClient(grpcAddr string, httpURL string) *GRPCSentimentClient {
	return &GRPCSentimentClient{
		grpcAddr:   grpcAddr,
		httpClient: NewSentimentClient(httpURL),
	}
}

// connect establishes lazy gRPC connection.
func (c *GRPCSentimentClient) connect() error {
	if c.conn != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10 MB
		),
	)
	if err != nil {
		return fmt.Errorf("grpc dial %s: %w", c.grpcAddr, err)
	}

	c.conn = conn
	slog.Info("gRPC sentiment client connected", "addr", c.grpcAddr)
	return nil
}

// Analyze sends texts for sentiment analysis via gRPC, falling back to HTTP.
func (c *GRPCSentimentClient) Analyze(ctx context.Context, texts []TextItem) (*SentimentResponse, error) {
	// Try gRPC first
	if c.grpcAddr != "" {
		if err := c.connect(); err == nil {
			result, err := c.analyzeGRPC(ctx, texts)
			if err == nil {
				return result, nil
			}
			slog.Warn("gRPC analyze failed, falling back to HTTP", "error", err)
		}
	}

	// Fallback to HTTP
	return c.httpClient.Analyze(ctx, texts)
}

// analyzeGRPC performs sentiment analysis over gRPC.
// NOTE: This uses raw proto encoding. When generated stubs are available,
// replace with strongly-typed client calls.
func (c *GRPCSentimentClient) analyzeGRPC(ctx context.Context, texts []TextItem) (*SentimentResponse, error) {
	// Build request using raw proto codec
	// For now, we use a JSON-encoded gRPC call pattern
	// This will be replaced with generated stubs:
	//   client := sentimentpb.NewSentimentServiceClient(c.conn)
	//   resp, err := client.Analyze(ctx, &sentimentpb.AnalyzeRequest{...})

	// Encode as JSON for generic invocation
	reqData := SentimentRequest{Texts: texts}

	// Use the connection health to decide
	state := c.conn.GetState()
	if state.String() == "TRANSIENT_FAILURE" || state.String() == "SHUTDOWN" {
		return nil, fmt.Errorf("gRPC connection in state: %s", state)
	}

	// For production: use generated stubs. For now, delegate to HTTP
	// but keep the connection infrastructure ready.
	slog.Debug("gRPC connection ready", "state", state, "texts", len(reqData.Texts))
	return nil, fmt.Errorf("proto stubs not yet generated — using HTTP fallback")
}

// Health checks gRPC connectivity, falls back to HTTP.
func (c *GRPCSentimentClient) Health(ctx context.Context) error {
	if c.grpcAddr != "" {
		if err := c.connect(); err == nil {
			state := c.conn.GetState()
			if state.String() == "READY" || state.String() == "IDLE" {
				return nil
			}
		}
	}
	return c.httpClient.Health(ctx)
}

// Close shuts down the gRPC connection.
func (c *GRPCSentimentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
