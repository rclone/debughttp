package debughttp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Check transport implements the interface
var _ http.RoundTripper = (*Transport)(nil)

// returns the "%p" representation of the thing passed in
func ptr(p interface{}) string {
	return fmt.Sprintf("%p", p)
}

func TestSetDefaults(t *testing.T) {
	old := http.DefaultTransport.(*http.Transport)
	newT := new(http.Transport)
	setDefaults(newT, old)
	// Can't use assert.Equal or reflect.DeepEqual for this as it has functions in
	// Check functions by comparing the "%p" representations of them
	assert.Equal(t, ptr(old.Proxy), ptr(newT.Proxy), "when checking .Proxy")
	assert.Equal(t, ptr(old.DialContext), ptr(newT.DialContext), "when checking .DialContext")
	// Check the other public fields
	assert.Equal(t, ptr(old.Dial), ptr(newT.Dial), "when checking .Dial")
	assert.Equal(t, ptr(old.DialTLS), ptr(newT.DialTLS), "when checking .DialTLS")
	assert.Equal(t, old.TLSClientConfig, newT.TLSClientConfig, "when checking .TLSClientConfig")
	assert.Equal(t, old.TLSHandshakeTimeout, newT.TLSHandshakeTimeout, "when checking .TLSHandshakeTimeout")
	assert.Equal(t, old.DisableKeepAlives, newT.DisableKeepAlives, "when checking .DisableKeepAlives")
	assert.Equal(t, old.DisableCompression, newT.DisableCompression, "when checking .DisableCompression")
	assert.Equal(t, old.MaxIdleConns, newT.MaxIdleConns, "when checking .MaxIdleConns")
	assert.Equal(t, old.MaxIdleConnsPerHost, newT.MaxIdleConnsPerHost, "when checking .MaxIdleConnsPerHost")
	assert.Equal(t, old.IdleConnTimeout, newT.IdleConnTimeout, "when checking .IdleConnTimeout")
	assert.Equal(t, old.ResponseHeaderTimeout, newT.ResponseHeaderTimeout, "when checking .ResponseHeaderTimeout")
	assert.Equal(t, old.ExpectContinueTimeout, newT.ExpectContinueTimeout, "when checking .ExpectContinueTimeout")
	assert.Equal(t, old.TLSNextProto, newT.TLSNextProto, "when checking .TLSNextProto")
	assert.Equal(t, old.MaxResponseHeaderBytes, newT.MaxResponseHeaderBytes, "when checking .MaxResponseHeaderBytes")
}

func TestCleanAuth(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{"", ""},
		{"floo", "floo"},
		{"Authorization: ", "Authorization: "},
		{"Authorization: \n", "Authorization: \n"},
		{"Authorization: A", "Authorization: X"},
		{"Authorization: A\n", "Authorization: X\n"},
		{"Authorization: AAAA", "Authorization: XXXX"},
		{"Authorization: AAAA\n", "Authorization: XXXX\n"},
		{"Authorization: AAAAA", "Authorization: XXXX"},
		{"Authorization: AAAAA\n", "Authorization: XXXX\n"},
		{"Authorization: AAAA\n", "Authorization: XXXX\n"},
		{"Authorization: AAAAAAAAA\nPotato: Help\n", "Authorization: XXXX\nPotato: Help\n"},
		{"Sausage: 1\nAuthorization: AAAAAAAAA\nPotato: Help\n", "Sausage: 1\nAuthorization: XXXX\nPotato: Help\n"},
	} {
		got := string(cleanAuth([]byte(test.in), Auth[0]))
		assert.Equal(t, test.want, got, test.in)
	}
}

func TestCleanAuths(t *testing.T) {
	transport := NewDefault(nil)
	for _, test := range []struct {
		in   string
		want string
	}{
		{"", ""},
		{"floo", "floo"},
		{"Authorization: AAAAAAAAA\nPotato: Help\n", "Authorization: XXXX\nPotato: Help\n"},
		{"X-Auth-Token: AAAAAAAAA\nPotato: Help\n", "X-Auth-Token: XXXX\nPotato: Help\n"},
		{"X-Auth-Token: AAAAAAAAA\nAuthorization: AAAAAAAAA\nPotato: Help\n", "X-Auth-Token: XXXX\nAuthorization: XXXX\nPotato: Help\n"},
	} {
		got := string(transport.cleanAuths([]byte(test.in)))
		assert.Equal(t, test.want, got, test.in)
	}
}

func TestTransport(t *testing.T) {
	const (
		requestBody  = "Request text"
		responseBody = "Response body"
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, responseBody)
		if r.Body != nil {
			buf, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			if len(buf) > 0 {
				fmt.Fprintln(w, strings.ToUpper(string(buf)))
			}
		}
	}))
	defer ts.Close()

	var lines []string
	logf := func(format string, v ...interface{}) {
		line := fmt.Sprintf(format, v...)
		line = strings.Replace(line, "\r", "", -1)
		lines = append(lines, line)
	}

	for _, test := range []struct {
		name         string
		flags        DumpFlags
		wantHeaders  bool
		wantReqBody  bool
		wantRespBody bool
		wantAuth     bool
	}{
		{
			name:  "NoDump",
			flags: 0,
		},
		{
			name:        "DumpHeaders",
			flags:       DumpHeaders,
			wantHeaders: true,
		},
		{
			name:         "DumpBodies",
			flags:        DumpBodies,
			wantHeaders:  true,
			wantReqBody:  true,
			wantRespBody: true,
		},
		{
			name:         "DumpRequests",
			flags:        DumpRequests,
			wantHeaders:  true,
			wantReqBody:  true,
			wantRespBody: false,
		},
		{
			name:         "DumpResponses",
			flags:        DumpResponses,
			wantHeaders:  true,
			wantReqBody:  false,
			wantRespBody: true,
		},
		{
			name:         "DumpResponsesWithAuth",
			flags:        DumpResponses | DumpAuth,
			wantHeaders:  true,
			wantReqBody:  false,
			wantRespBody: true,
			wantAuth:     true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(&Options{
				Flags: test.flags,
				Logf:  logf,
			})
			lines = nil

			// Do the test request
			req, err := http.NewRequest("PUT", ts.URL, bytes.NewBufferString(requestBody))
			require.NoError(t, err)
			req.Header.Set("Authorization", "POTATO")
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
			expectedResponse := responseBody + "\n" + strings.ToUpper(requestBody) + "\n"
			assert.Equal(t, expectedResponse, string(body))

			if !test.wantHeaders {
				assert.Equal(t, 0, len(lines))
				return
			}

			// Check what we expect was logged
			require.Equal(t, 8, len(lines))
			assert.Equal(t, SeparatorReq, lines[0])
			assert.Contains(t, lines[1], "HTTP REQUEST")
			assert.Contains(t, lines[2], "PUT / HTTP")
			if test.wantAuth {
				assert.Contains(t, lines[2], "\nAuthorization: POTATO\n")
			} else {
				assert.Contains(t, lines[2], "\nAuthorization: XXXX\n")
			}
			if test.wantReqBody {
				assert.Contains(t, lines[2], requestBody)
			} else {
				assert.NotContains(t, lines[2], requestBody)
			}
			assert.Equal(t, SeparatorReq, lines[3])
			assert.Equal(t, SeparatorResp, lines[4])
			assert.Contains(t, lines[5], "HTTP RESPONSE")
			assert.Contains(t, lines[6], "200 OK\n")
			if test.wantRespBody {
				assert.Contains(t, lines[6], expectedResponse)
			} else {
				assert.NotContains(t, lines[6], expectedResponse)
			}
			assert.Equal(t, SeparatorResp, lines[7])
		})
	}
}
