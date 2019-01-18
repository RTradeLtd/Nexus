package client

import (
	"fmt"

	"github.com/RTradeLtd/grpc/dialer"
	"github.com/RTradeLtd/grpc/nexus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/RTradeLtd/Nexus/config"
)

// IPFSOrchestratorClient is a lighweight container for the orchestrator's
// gRPC API client
type IPFSOrchestratorClient struct {
	nexus.ServiceClient
	grpc *grpc.ClientConn
}

// New instantiates a new orchestrator API client
func New(opts config.API, devMode bool) (*IPFSOrchestratorClient, error) {
	var (
		c        = &IPFSOrchestratorClient{}
		dialOpts []grpc.DialOption
	)

	if opts.Key != "" {
		dialOpts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Key, !devMode)),
		}
	}

	if opts.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	// connect to orchestrator
	var err error
	c.grpc, err = grpc.Dial(opts.Host+":"+opts.Port, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core service: %s", err.Error())
	}
	c.ServiceClient = nexus.NewServiceClient(c.grpc)
	return c, nil
}

// Close shuts down the client's gRPC connection
func (i *IPFSOrchestratorClient) Close() { i.grpc.Close() }
