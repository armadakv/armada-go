# Armada Go client
[![GoDoc](https://pkg.go.dev/badge/github.com/armadakv/armada-go.svg)](https://pkg.go.dev/github.com/armadakv/armada-go)
[![tag](https://img.shields.io/github/v/tag/armadakv/armada-go)](https://github.com/armadakv/armada-go/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20-%23007d9c)
![Build Status](https://github.com/armadakv/armada-go/actions/workflows/test.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/armadakv/armada-go/badge.svg?branch=main)](https://coveralls.io/github/armadakv/armada-go?branch=main)
[![Go report](https://goreportcard.com/badge/github.com/armadakv/armada-go)](https://goreportcard.com/report/github.com/armadakv/armada-go)
[![Contributors](https://img.shields.io/github/contributors/armadakv/armada-go)](https://github.com/armadakv/armada-go/graphs/contributors)
[![License](https://img.shields.io/github/license/armadakv/armada-go)](LICENSE)

This repository hosts the Go client for [**Armada**](https://github.com/armadakv/armada). For documentation and examples check the [package documentation](https://pkg.go.dev/github.com/armadakv/armada-go).
Additional functionality like Prometheus metrics and OpenTelemetry tracing is provided using [plugins](https://github.com/armadakv/armada-go/tree/main/plugin).

## Example use

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	client "github.com/armadakv/armada-go"
)

func main() {
	// Create Armada client
	c, err := client.New(
		client.WithEndpoints("127.0.0.1:8443"),
		client.WithLogger(client.PrintLogger{}),
		client.WithBearerToken(os.Getenv("ARMADA_TOKEN")),
		client.WithSecureConfig(&client.SecureConfig{
			InsecureSkipVerify: true, // Skip verification of self-signed certificate
		}),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Provide operation timeout
	defer cancel()
	
	put, err := c.Table("armada-test").Put(ctx, "foo", "bar")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", put)
}
```

## Authentication

The client supports Armada identity tokens via `Authorization: Bearer` metadata:

```go
c, err := client.New(
	client.WithEndpoints("127.0.0.1:8443"),
	client.WithBearerToken(os.Getenv("ARMADA_TOKEN")),
	client.WithSecureConfig(&client.SecureConfig{
		InsecureSkipVerify: true, // Skip verification of self-signed certificate
	}),
)
```

For tokens that need refresh, use `client.WithBearerTokenProvider`.

## Migration

- module path moved from `github.com/jamf/regatta-go` to `github.com/armadakv/armada-go`
- plugin module paths now live under `github.com/armadakv/armada-go/plugin/...`
- bearer-token auth is now a first-class client option through `WithBearerToken` and `WithBearerTokenProvider`

## Contributing

Armada is in active development and contributors are welcome! Feel free to ask questions and engage in [GitHub Discussions](https://github.com/armadakv/armada-go/discussions)!
