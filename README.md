[<img src="https://rclone.org/img/logo_on_light__horizontal_color.svg" width="50%" alt="rclone logo">](https://rclone.org/)

[![Go Docs](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/rclone/debughttp)
[![Build Status](https://github.com/rclone/debughttp/workflows/build/badge.svg)](https://github.com/rclone/debughttp/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/rclone/debughttp)](https://goreportcard.com/report/github.com/rclone/debughttp)

# debughttp

This is a go package to dump HTTP requests and responses

## Install

Install like this

    go get github.com/rclone/debughttp

and this will build the binary in `$GOPATH/bin`.

## Usage

See the [full docs](https://pkg.go.dev/github.com/rclone/debughttp)
or read on for a quickstart
Instead of using http.Get or client.Get, use this

```go
import github.com/rclone/debughttp

/ Make a client with the defaults which dump headers to log.Printf
client := debughttp.NewClient(nil)

// Now use the client, eg
resp, err := client.Get("http://example.com")
```

This will log something like this

```
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
```

If you want to see the bodies of the transactions use this

```go
// Make a client with the defaults which dump headers and bodies to log.Printf
client := debughttp.NewClient(debughttp.DumpBodyOptions)
```

Note that this redacts authorization headers by default.

For more info see [the full documentation](https://pkg.go.dev/github.com/rclone/debughttp).

## License

This is free software under the terms of the MIT license (check the
LICENSE file included in this package).

This code was originally part of the rclone binary but was factored out to be of wider use.

## Contact and support

The project website is at:

- https://github.com/rclone/debughttp

There you can file bug reports, ask for help or contribute patches.
