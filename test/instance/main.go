package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dns3l/dns3l-go/api"
	"github.com/dns3l/dns3l-go/auth"
	"github.com/dns3l/dns3l-go/client"
	log "github.com/sirupsen/logrus"
)

func main() {

	instance := os.Getenv("DNS3L_INSTANCE")
	ca := os.Getenv("DNS3L_TESTCA")
	cert := os.Getenv("DNS3L_TESTCERT")
	username := os.Getenv("DNS3L_USER")
	password := os.Getenv("DNS3L_PASS")
	if instance == "" || cert == "" || ca == "" || username == "" || password == "" {
		log.Fatal("missing env parameter")
	}

	baseurl, err := url.Parse(instance)
	if err != nil {
		log.WithError(err).Fatalf("could not parse instance URL")
	}
	apiurl := baseurl.JoinPath("/api/v1")

	ctx := context.Background()
	defaultHTTPClient := &http.Client{Timeout: 10 * time.Second}
	longHTTPClient := &http.Client{Timeout: 5 * time.Minute}

	log.Info("Getting server info..")

	clNoAuth := client.NewUnauthenticated(apiurl, defaultHTTPClient)

	info, err := clNoAuth.GetServerInfo(ctx)
	if err != nil {
		log.WithError(err).Fatalf("could not get server info")
	}
	log.WithField("serverinfo", info).Info("got server info")

	actx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	log.Info("Authenticating...")
	token, err := auth.AuthAPIUserPass(actx, instance, username, password)
	if err != nil {
		log.WithError(err).Fatal("could not get token")
	}

	clAuth := client.NewByOauth2AccessToken(apiurl, defaultHTTPClient, token.AccessToken)
	clAuthLong := client.NewByOauth2AccessToken(apiurl, longHTTPClient, token.AccessToken)

create:
	log.Info("Claiming certificate (takes long)..")
	err = clAuthLong.ClaimCertificate(ctx, ca, &api.CertClaimInfo{
		Name: cert,
	})
	if err != nil {
		if errors.Is(err, client.ErrConflict) {
			log.Info("Deleting already existing cert...")
			err := clAuth.DeleteCertByCA(ctx, ca, cert)
			if err != nil {
				log.WithError(err).Fatalf("could not delete already existing cert.")
			}
			goto create
		}
		log.WithError(err).Fatalf("could not claim certificate")
	}
	defer deleteCertAgain(ctx, clAuth, ca, cert)

	log.Info("Getting cert's key...")
	key, err := clAuth.GetResourceKey(ctx, ca, cert)
	if err != nil {
		log.WithError(err).Fatalf("could not get cert's key")
	}
	log.WithField("publickey", key.PublicKey).Info("got private key (only showing pubkey in it)")

	log.Info("Getting cert...")
	certRes, err := clAuth.GetResourceCert(ctx, ca, cert)
	if err != nil {
		log.WithError(err).Fatalf("could not get cert's resource")
	}
	log.WithField("subject", certRes.Subject).Info("got cert's resource")

	log.Info("Getting fullchain...")
	certChain, err := clAuth.GetResourceFullChain(ctx, ca, cert)
	if err != nil {
		log.WithError(err).Fatalf("could not get cert's fullchain")
	}
	log.WithField("chainlen", len(certChain)).Info("got chain")

	log.Info("Listing all certs of the CA (limit 4)...")
	certs, _, err := clAuth.GetCertsByCA(ctx, ca, &client.CertQueryOptions{
		Limit: 4,
	})
	if err != nil {
		log.WithError(err).Fatalf("could not list certs")
	}
	certsRenewals := make([]string, len(certs))
	for i, cert := range certs {
		certsRenewals[i] = cert.NextRenewal
	}
	log.WithField("certsRenewals", certsRenewals).Info("here are the certs' renewal dates")

}

func deleteCertAgain(ctx context.Context, clAuth client.Client, ca, cert string) {
	log.Info("Deleting cert again...")
	err := clAuth.DeleteCertByCA(ctx, ca, cert)
	if err != nil {
		log.WithError(err).Fatalf("could not delete cert, you may have to delete it yourself")
	}
}
