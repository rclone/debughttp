package debughttp_test

import (
	"net/http"

	"github.com/rclone/debughttp"
)

var (
	existingTransport *http.Transport
	transport         *debughttp.Transport
	client            *http.Client
)

func myLogf(format string, v ...interface{}) {}

func ExampleNewClient() {
	// Make a client with the defaults which dump headers to log.Printf
	client = debughttp.NewClient(nil)

	// Make a client which dumps headers and bodies to log.Printf
	client = debughttp.NewClient(&debughttp.DumpBodyOptions)

	// Make a client with full options
	// This dumps headers, request bodies and doesn't redact the auth
	client = debughttp.NewClient(&debughttp.Options{
		Flags: debughttp.DumpRequests | debughttp.DumpAuth,
		Logf:  myLogf,
	})
}

func ExampleNewDefault() {
	// Make a transport with the defaults which dump headers to log.Printf
	transport = debughttp.NewDefault(nil)

	// Make a transport which dumps headers and bodies to log.Printf
	transport = debughttp.NewDefault(&debughttp.DumpBodyOptions)

	// Make a transport with full options
	// This dumps headers, request bodies and doesn't redact the auth
	transport = debughttp.NewDefault(&debughttp.Options{
		Flags: debughttp.DumpRequests | debughttp.DumpAuth,
		Logf:  myLogf,
	})
}

func ExampleNew() {
	// Make a transport with the defaults which dump headers to log.Printf
	transport = debughttp.New(nil, existingTransport)

	// Make a transport which dumps headers and bodies to log.Printf
	transport = debughttp.New(&debughttp.DumpBodyOptions, existingTransport)

	// Make a transport with full options
	// This dumps headers, request bodies and doesn't redact the auth
	transport = debughttp.New(&debughttp.Options{
		Flags: debughttp.DumpRequests | debughttp.DumpAuth,
		Logf:  myLogf,
	}, existingTransport)
}
