package relay

import (
	"bytes"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"

	pb "github.com/rusenask/webhookrelayd/grpc/webhook"

	log "github.com/Sirupsen/logrus"
)

// Relayer - relayer interface
type Relayer interface {
	Relay(wh *pb.WebhookRequest) error
}

// DefaultRelayer - default 'last mile' webhook relayer
type DefaultRelayer struct {
	// client *http.Client

	rClient *retryablehttp.Client

	// maximum number of retries that this relayer should try
	// before giving up.
	retries int

	// backoff strategy in seconds
	backoff int
}

// Opts - configuration
type Opts struct {
	Retries  int
	Insecure bool
}

// NewDefaultRelayer - create an instance of default relayer
func NewDefaultRelayer(opts *Opts) *DefaultRelayer {
	client := retryablehttp.NewClient()

	if opts.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		insecureClient := &http.Client{Transport: tr}
		client.HTTPClient = insecureClient
	}

	client.RetryMax = opts.Retries

	return &DefaultRelayer{rClient: client}
}

func getHeaders(wh *pb.WebhookRequest) http.Header {
	headers := make(map[string][]string)
	for k, v := range wh.Request.Header.Headers {
		headers[k] = v.Values
	}

	return headers
}

// Relay - relaying incomming webhook to original destination
func (r *DefaultRelayer) Relay(wh *pb.WebhookRequest) error {
	req, err := retryablehttp.NewRequest(wh.Request.Method, wh.Request.Destination, bytes.NewReader(wh.Request.Body))

	// adding headers
	req.Header = getHeaders(wh)

	resp, err := r.rClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode > 399 {
		log.WithFields(log.Fields{
			"status_code": resp.StatusCode,
			"destination": wh.Request.Destination,
			"method":      wh.Request.Method,
		}).Warn("relayer: unexpected status code")
	}

	return nil
}
