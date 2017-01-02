package client

import (
	"fmt"
	"net"
	"time"

	pb "github.com/rusenask/webhookrelayd/grpc/webhook"
	"google.golang.org/grpc"

	"testing"

	log "github.com/Sirupsen/logrus"
)

type SrvOpts struct {
	Webhooks []*pb.WebhookRequest

	// meant for testing, for example reject new connection due to auth failure
	ConnectionError error

	Port int
}

type server struct {
	opts *SrvOpts
}

// DummyRelayer - dummy relayer to capture "relayed" requests
type DummyRelayer struct {
	Relayed []*pb.WebhookRequest

	Error error
}

func (d *DummyRelayer) Relay(wh *pb.WebhookRequest) error {
	if d.Error != nil {
		return d.Error
	}
	d.Relayed = append(d.Relayed, wh)
	return nil
}

func NewTestingServer(testingOpts *SrvOpts) func() {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", testingOpts.Port))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"port":  testingOpts.Port,
		}).Fatal("failed to create TCP listener")
	}
	var opts []grpc.ServerOption
	s := &server{opts: testingOpts}

	grpcSrv := grpc.NewServer(opts...)
	pb.RegisterWebhookServer(grpcSrv, s)

	go grpcSrv.Serve(listener)

	time.Sleep(60 * time.Millisecond)

	return grpcSrv.Stop
}

// GetWebhooks - stream webhooks
func (s *server) GetWebhooks(filter *pb.WebhookFilter, stream pb.Webhook_GetWebhooksServer) error {

	if s.opts.ConnectionError != nil {
		return s.opts.ConnectionError
	}

	for _, req := range s.opts.Webhooks {
		err := stream.Send(req)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestDefaultClient(t *testing.T) {
	req := &pb.WebhookRequest{
		Bucket: &pb.Bucket{
			Id:   "xx",
			Name: "bucket_name",
		},
		Request: &pb.Request{
			Destination: "http://localhost:3000",
			Method:      "GET",
		},
	}
	reqs := []*pb.WebhookRequest{req}

	port := 34444
	teardown := NewTestingServer(&SrvOpts{Port: port, Webhooks: reqs})
	defer teardown()
	dr := &DummyRelayer{}

	// getting client
	clientOpts := &Opts{
		Address:      fmt.Sprintf("localhost:%d", port),
		AccessKey:    "dummy",
		AccessSecret: "dummy",
	}
	client := NewDefaultClient(clientOpts, dr)
	err := client.StartRelay(&Filter{})
	if err != nil {
		t.Errorf("failed to start client relay")
	}

	if len(dr.Relayed) == 0 {
		t.Errorf("expected to find wh request in dummy relayer")
	} else {
		// checking relayed details
		if dr.Relayed[0].Bucket.Id != req.Bucket.Id {
			t.Errorf("expected bucket ID: %s, got: %s", req.Bucket.Id, dr.Relayed[0].Bucket.Id)
		}

		if dr.Relayed[0].Bucket.Name != req.Bucket.Name {
			t.Errorf("expected bucket name: %s, got: %s", req.Bucket.Name, dr.Relayed[0].Bucket.Name)
		}

		if dr.Relayed[0].Request.Destination != req.Request.Destination {
			t.Errorf("expected bucket destination: %s, got: %s", req.Request.Destination, dr.Relayed[0].Request.Destination)
		}
	}
}

func TestDefaultClientWithError(t *testing.T) {

	port := 34445
	teardown := NewTestingServer(&SrvOpts{Port: port, ConnectionError: fmt.Errorf("dummy testing error")})
	defer teardown()
	dr := &DummyRelayer{}

	// getting client
	clientOpts := &Opts{
		Address:      fmt.Sprintf("localhost:%d", port),
		AccessKey:    "dummy",
		AccessSecret: "dummy",
	}
	client := NewDefaultClient(clientOpts, dr)
	err := client.StartRelay(&Filter{})
	if err.Error() != "rpc error: code = 2 desc = dummy testing error" {
		t.Errorf("expected dummy testing error, got: %s", err.Error())
	}
}

func TestDefaultClientRelayerError(t *testing.T) {
	req := &pb.WebhookRequest{
		Bucket: &pb.Bucket{
			Id:   "xx",
			Name: "bucket_name",
		},
		Request: &pb.Request{
			Destination: "http://localhost:3000",
			Method:      "GET",
		},
	}
	reqs := []*pb.WebhookRequest{req}

	port := 34446
	teardown := NewTestingServer(&SrvOpts{Port: port, Webhooks: reqs})
	defer teardown()
	dr := &DummyRelayer{Error: fmt.Errorf("relayer error")}

	// getting client
	clientOpts := &Opts{
		Address:      fmt.Sprintf("localhost:%d", port),
		AccessKey:    "dummy",
		AccessSecret: "dummy",
	}
	client := NewDefaultClient(clientOpts, dr)
	err := client.StartRelay(&Filter{})
	if err != nil {
		t.Errorf("failed to start client relay")
	}

	if len(dr.Relayed) != 0 {
		t.Errorf("expected to find no wh requests in dummy relayer")
	}
}
