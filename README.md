# Go API structs and client library for DNS3L access

This project carries the Go API structs matching the [official
DNS3L API](https://github.com/dns3l/dns3l/blob/master/openapi.yaml).
Furthermore, it contains a client library to access a DNS3L instance's API.


## Client library

If you don't want to mess with raw PEM and JSON, you can easily access certificates
and their resources using the Go `crypto/x509` and `crypto/rsa` structs, and obtain cert lists
in API Go structs format. But if you want to obtain raw formats, you can access a raw API that
return streams of JSON or PEM data as sent by the server.

### Usage example (unauthenticated complex and raw access)

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/url"

	dns3l "github.com/dns3l/dns3l-go/client"
)

func main() {

	ctx := context.Background()

	url, err := url.Parse("https://dns3l.example.com/api/v1")
	if err != nil {
		panic(err)
	}

	// Create an unauthenticated client
	c := dns3l.NewUnauthed(url, nil)

	// Returns the cert
	crt, err := c.GetResourceCert(ctx, "les", "foo-cert.example.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(crt.Subject.CommonName, crt.DNSNames)

	// Returns the cert as PEM data
	crtrdr, err := c.Raw().GetResource(ctx, "les", "foo-cert.example.com",
		dns3l.ResourceTypeCert)
	if err != nil {
		panic(err)
	}

	// read the entire stream
	defer crtrdr.Close()
	crtraw, err := io.ReadAll(crtrdr)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(crtraw))

}

```

### Usage example (authenticated complex access)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	dns3l "github.com/dns3l/dns3l-go/client"
	"golang.org/x/oauth2"
)

func main() {

	ctx := context.Background()

	instance := "https://dns3l.example.com"

	// First, we need to auth
	token, err := auth(ctx, instance)
	if err != nil {
		panic(err)
	}

	url, err := url.Parse(instance + "/api/v1")
	if err != nil {
		panic(err)
	}

	// Now we have the token, let's create an authenticated client
	c := dns3l.NewByOauth2AccessToken(url, nil, token.AccessToken)

	// Returns the private key
	k, err := c.GetResourceKey(ctx, "les", "foo-cert.example.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(k)

}

func auth(ctx context.Context, instance string) (*oauth2.Token, error) {
	cfg := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: instance + "/auth/token",
		},
		ClientID: "dns3l-api-pub",
		Scopes: []string{
			"openid",
			"profile",
			"groups",
			"email",
			"audience:server:client_id:dns3ld",
		},
	}

	username := os.Getenv("DNS3L_USER")
	password := os.Getenv("DNS3L_PASS")
	if username == "" || password == "" {
		return nil, errors.New("no username or password set in env")
	}

	token, err := cfg.PasswordCredentialsToken(
		ctx,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}

	return token, err
}

```

The client is extensively documented, so refer to the interface documentation
for more details.

# Instance test

You can run an [instance test](./test/instance) against your dns3l stack.

## API structs

The API structs are located in the [api](./api) subfolder.

## Versioning

This module is versioned according to the *major* and *minor* revisions of the [OpenAPI
description](https://github.com/dns3l/dns3l/blob/master/openapi.yaml).
The patch revision are used for client code patches. This means that if
you pick a version number of this module, you can expect it to adhere to the respective
OpenAPI specification with this version.
