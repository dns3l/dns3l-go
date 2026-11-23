// Package client contains a client library to access a DNS3L instance's API.
// If you don't want to mess with raw PEM and JSON, you can easily access certificates
// and their resources using the Go `crypto/x509` and `crypto/rsa` structs, and obtain cert lists
// in API Go structs format. But if you want to obtain raw formats, you can access a raw API that
// return streams of JSON or PEM data as sent by the server.
package client
