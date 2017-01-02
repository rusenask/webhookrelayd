package client

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/rusenask/webhookrelayd/grpc/webhook"
	"github.com/rusenask/webhookrelayd/relay"

	log "github.com/Sirupsen/logrus"
)

// Filter - optional filter that can be passed to the server
type Filter struct {
	Bucket, Destination string
}

// WebhookRelayClient - default webhookrelay interface
type WebhookRelayClient interface {
	StartRelay(filter *Filter) error
	Stop() error
}

const (
	address = "api.webhookrelay.com:40000"
)

type loginCreds struct {
	AccessKey, AccessSecret string
}

func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"access_key": c.AccessKey,
		"secret_key": c.AccessSecret,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return true
}

// Opts - client configuration
type Opts struct {
	Address, AccessKey, AccessSecret string
	Debug                            bool
}

// DefaultClient - default client that connects to webhookrelay service via gRPC protocol
type DefaultClient struct {
	conn *grpc.ClientConn

	opts *Opts

	relayer relay.Relayer
}

// NewDefaultClient - create new default client with given options
func NewDefaultClient(opts *Opts, relayer relay.Relayer) *DefaultClient {
	return &DefaultClient{opts: opts, relayer: relayer}
}

// StartRelay - start relaying
func (c *DefaultClient) StartRelay(filter *Filter) error {
	// Set up a connection to the gRPC server.
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(&loginCreds{
			AccessKey:    c.opts.AccessKey,
			AccessSecret: c.opts.AccessSecret,
		}),
		grpc.WithTimeout(5 * time.Second),
		grpc.WithBackoffMaxDelay(5 * time.Second),
	}

	conn, err := grpc.Dial(c.opts.Address, opts...)

	if err != nil {
		return err
	}

	c.conn = conn

	client := pb.NewWebhookClient(conn)
	log.WithFields(log.Fields{
		"host": c.opts.Address,
	}).Info("webhookrelayd: connected...")

	fl := &pb.WebhookFilter{Bucket: filter.Bucket, Destination: filter.Destination}
	return c.getWebhooks(client, fl)
}

func (c *DefaultClient) getWebhooks(client pb.WebhookClient, filter *pb.WebhookFilter) error {
	// calling the streaming API
	stream, err := client.GetWebhooks(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("error while getting webhooks: %s", err)
	}
	for {
		whRequest, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("failed to get stream from server")
			return err
		}

		err = c.relayer.Relay(whRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"destination": whRequest.Request.Destination,
				"method":      whRequest.Request.Method,
			}).Error("failed to relay webhook request")
			continue
		}

		log.WithFields(log.Fields{
			"bucket":      whRequest.Bucket.Name,
			"destination": whRequest.Request.Destination,
		}).Info("webhook request relayed")
	}

	return nil
}
