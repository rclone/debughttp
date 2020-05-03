/*
Package debughttp provides an http Transport or Client which can be
used for tracing HTTP requests for debugging purposes.

Quickstart

This can be used for a quick bit of debugging.

Instead of using http.Get or client.Get, use this

	// Make a client with the defaults which dump headers to log.Printf
	client := debughttp.NewClient(nil)

	// Now use the client, eg
	resp, err := client.Get("http://example.com")

This will log something like this

	2020/05/03 16:06:03 >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	2020/05/03 16:06:03 HTTP REQUEST (req 0xc00022a300)
	2020/05/03 16:06:03 GET / HTTP/1.1
	Host: example.com
	User-Agent: Go-http-client/1.1
	Accept-Encoding: gzip

	2020/05/03 16:06:03 >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	2020/05/03 16:06:03 <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	2020/05/03 16:06:03 HTTP RESPONSE (req 0xc00022a300)
	2020/05/03 16:06:03 HTTP/1.1 200 OK
	Accept-Ranges: bytes
	Age: 518408
	Cache-Control: max-age=604800
	Content-Type: text/html; charset=UTF-8
	Date: Sun, 03 May 2020 15:06:03 GMT
	Etag: "3147526947"
	Expires: Sun, 10 May 2020 15:06:03 GMT
	Last-Modified: Thu, 17 Oct 2019 07:18:26 GMT
	Server: ECS (nyb/1D2A)
	Vary: Accept-Encoding
	X-Cache: HIT

	2020/05/03 16:06:03 <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

If you want to see the bodies of the transactions use this

	// Make a client with the defaults which dump headers and bodies to log.Printf
	client := debughttp.NewClient(debughttp.DumpBodyOptions)

Note that this redacts authorization headers by default.

Fuller integration

If you want more control over what is logged and what isn't logged
then you can use the Options struct, eg

	client := debughttp.NewClient(&debughttp.Options{
		Flags: debughttp.DumpRequests|debughttp.DumpAuth,
	}

If you are integrating this with code which has its own logging system
then you will want to pass in the Logf parameter to control where
the logs are sent.

Every Go library which does HTTP transactions on your behalf should
take an http.Client or allow the setting of an http.Transport
replacement. (If you find one which doesn't, then report an issue!)

To create a new Transport use the NewDefault function to base one
off the default transport or the New function to base one off an
existing transport.

This means that you can use this library for debugging other people's
code. For example this is how you add this library to the AWS SDK

	client := debughttp.NewClient(nil)
        awsConfig := aws.NewConfig().
                WithCredentials(cred).
                WithHTTPClient(fshttp.NewClient(fs.Config))

If you do this you can see exactly what requests are sent to and from
AWS.

Warnings

If dumping bodies is enabled the bodies are held in memory so large
requests and responses can use a lot of memory.

The Accept-Encoding as shown may not be correct in the Request and
the Response may not show Content-Encoding if the Go standard
libraries auto gzip encoding was in effect. In this case the body of
the request will be gunzipped before showing it.
*/
package debughttp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"reflect"
)

var (
	// Default separators for request and responses
	SeparatorReq  = ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
	SeparatorResp = "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
)

// DumpFlags describes the Dump options in force
type DumpFlags int

// DumpFlags definitions
const (
	DumpHeaders   DumpFlags = 1 << iota // dump just the http headers
	DumpBodies                          // dump the bodies also
	DumpRequests                        // dump all the headers and the request bodies but not the response bodies
	DumpResponses                       // dump all the headers and the response bodies but not the request bodies
	DumpAuth                            // dump the auth instead of redacting it
)

// Options controls the configuration of the HTTP debugging
type Options struct {
	Flags DumpFlags                             // Which parts of the HTTP transaction we are dumping
	Logf  func(format string, v ...interface{}) // Where to log the dumped transactions - defaults to log.Printf if not set
	Auth  [][]byte                              // which headers we are treating as Auth to redact - defaults to Auth if not set
}

// Default options if nil is passed in to New or NewDefault or NewClient
var DefaultOptions = Options{
	Flags: DumpHeaders,
	Logf:  log.Printf,
	Auth:  Auth,
}

// DumpBodyOptions is an easy set of options for dumping bodies
var DumpBodyOptions = Options{
	Flags: DumpBodies,
	Logf:  log.Printf,
	Auth:  Auth,
}

// Auth is the headers which we redact if DumpAuth is not set in Options
var Auth = [][]byte{
	[]byte("Authorization: "),
	[]byte("X-Auth-Token: "),
}

// Transport wraps an *http.Transport and logs requests and responses
//
// Create one with New, NewDefault or NewClient - don't use directly
type Transport struct {
	*http.Transport
	opt Options
}

// New wraps the http.Transport passed in and logs all
// round trips according to the Flags in opt
func New(opt *Options, transport *http.Transport) *Transport {
	if opt == nil {
		opt = &DefaultOptions
	}
	t := &Transport{
		Transport: transport,
		opt:       *opt,
	}
	if t.opt.Logf == nil {
		t.opt.Logf = log.Printf
	}
	if t.opt.Auth == nil {
		t.opt.Auth = Auth
	}
	return t
}

// setDefaults for a from b
//
// Copy the public members from b to a.  We can't just use a struct
// copy as Transport contains a private mutex.
func setDefaults(a, b interface{}) {
	pt := reflect.TypeOf(a)
	t := pt.Elem()
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()
	for i := 0; i < t.NumField(); i++ {
		aField := va.Field(i)
		// Set a from b if it is public
		if aField.CanSet() {
			bField := vb.Field(i)
			aField.Set(bField)
		}
	}
}

// NewDefault returns an http.RoundTripper based off
// http.DefaultTransport which will log the HTTP transactions as
// directed in opt
//
// If opt is nil then DefaultOptions is used
func NewDefault(opt *Options) *Transport {
	// Start with a sensible set of defaults then override.
	// This also means we get new stuff when it gets added to go
	t := new(http.Transport)
	setDefaults(t, http.DefaultTransport.(*http.Transport))

	// Wrap that http.Transport in our own transport
	return New(opt, t)
}

// NewClient returns an http.Client based off a transport which will
// log the HTTP transactions as directed in opt
func NewClient(opt *Options) *http.Client {
	client := &http.Client{
		Transport: NewDefault(opt),
	}
	return client
}

// cleanAuth gets rid of one authBuf header within the first 4k
func cleanAuth(buf, authBuf []byte) []byte {
	// Find how much buffer to check
	n := 4096
	if len(buf) < n {
		n = len(buf)
	}
	// See if there is an Authorization: header
	i := bytes.Index(buf[:n], authBuf)
	if i < 0 {
		return buf
	}
	i += len(authBuf)
	// Overwrite the next 4 chars with 'X'
	for j := 0; i < len(buf) && j < 4; j++ {
		if buf[i] == '\n' {
			break
		}
		buf[i] = 'X'
		i++
	}
	// Snip out to the next '\n'
	j := bytes.IndexByte(buf[i:], '\n')
	if j < 0 {
		return buf[:i]
	}
	n = copy(buf[i:], buf[i+j:])
	return buf[:i+n]
}

// cleanAuths gets rid of all the possible Auth headers
func (t *Transport) cleanAuths(buf []byte) []byte {
	for _, authBuf := range t.opt.Auth {
		buf = cleanAuth(buf, authBuf)
	}
	return buf
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Logf request
	if t.opt.Flags&(DumpHeaders|DumpBodies|DumpAuth|DumpRequests|DumpResponses) != 0 {
		t.opt.Logf("%s", SeparatorReq)
		t.opt.Logf("%s (req %p)", "HTTP REQUEST", req)
		buf, derr := httputil.DumpRequestOut(req, t.opt.Flags&(DumpBodies|DumpRequests) != 0)
		if derr != nil {
			t.opt.Logf("Dump request failed: %v", derr)
		} else {
			if t.opt.Flags&DumpAuth == 0 {
				buf = t.cleanAuths(buf)
			}
			t.opt.Logf("%s", string(buf))
		}
		t.opt.Logf("%s", SeparatorReq)
	}
	// Do round trip
	resp, err = t.Transport.RoundTrip(req)
	// Logf response
	if t.opt.Flags&(DumpHeaders|DumpBodies|DumpAuth|DumpRequests|DumpResponses) != 0 {
		t.opt.Logf("%s", SeparatorResp)
		t.opt.Logf("%s (req %p)", "HTTP RESPONSE", req)
		if err != nil {
			t.opt.Logf("HTTP request failed: %v", err)
		} else {
			buf, derr := httputil.DumpResponse(resp, t.opt.Flags&(DumpBodies|DumpResponses) != 0)
			if derr != nil {
				t.opt.Logf("Dump response failed: %v", derr)
			} else {
				t.opt.Logf("%s", string(buf))
			}
		}
		t.opt.Logf("%s", SeparatorResp)
	}
	return resp, err
}
