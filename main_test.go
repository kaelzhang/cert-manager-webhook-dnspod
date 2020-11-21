package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
	"github.com/kaelzhang/cert-manager-webhook-dnspod/webhook"
)

var (
	zone  = os.Getenv("TEST_ZONE_NAME")
	group = os.Getenv("GROUP_NAME")
)

func TestRunsSuite(t *testing.T) {
	t.Logf("TEST_ZONE_NAME %s", zone)
	t.Logf("GROUP_NAME %s", group)

	fixture := dns.NewFixture(webhook.Solver(),
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/my-custom-solver"),
		dns.SetBinariesPath("kubebuilder/bin"),
	)

	fixture.RunConformance(t)
}
