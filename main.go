package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rusenask/webhookrelayd/grpc/client"
	"github.com/rusenask/webhookrelayd/relay"
)

// Version - client version
const Version = "0.1.0"

// defaults
const (
	DefaultServerAddress = "api.webhookrelay.com"
	DefaultServerPort    = 40000
)

var usageStr = `
Usage: webhookrelayd [options]
Server Options:
    -k, --key <key>                  Bind to host address (default: 0.0.0.0)
    -s, --secret <secret>            Access secret to use
    -a, --address <address>          Address of the webhook-relay server to connect (default: api.webhookrelay.com)

Common Options:
    -h, --help                       Show this message
    -v, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func main() {
	// Server Options
	opts := client.Opts{}

	var showVersion bool

	// Parse flags
	flag.StringVar(&opts.AccessKey, "k", "", "Access key to use")
	flag.StringVar(&opts.AccessKey, "key", "", "Access key to use")

	flag.StringVar(&opts.AccessSecret, "s", "", "Access secret to use")
	flag.StringVar(&opts.AccessSecret, "secret", "", "Access secret to use")

	flag.BoolVar(&opts.Debug, "D", false, "Enable Debug logging.")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable Debug logging.")

	flag.StringVar(&opts.Address, "a", "", "Server address to connect to")
	flag.StringVar(&opts.Address, "address", "", "Server address to connect to")

	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.BoolVar(&showVersion, "v", false, "Print version information.")

	flag.Usage = usage

	flag.Parse()

	if showVersion {
		fmt.Printf("Webhookrelayd version: %s", Version)
		os.Exit(0)
	}

	if opts.Address == "" {
		opts.Address = fmt.Sprintf("%s:%d", DefaultServerAddress, DefaultServerPort)
	}

	// getting relayer
	rOpts := &relay.Opts{Retries: 5}
	relayer := relay.NewDefaultRelayer(rOpts)

	c := client.NewDefaultClient(&opts, relayer)

	c.StartRelay(&client.Filter{})
}
