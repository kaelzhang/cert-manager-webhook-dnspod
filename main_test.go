package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
)

var (
	zone  = os.Getenv("TEST_ZONE_NAME")
	group = os.Getenv("GROUP_NAME")
)

func TestRunsSuite(t *testing.T) {
	t.Logf("TEST_ZONE_NAME %s", zone)
	t.Logf("GROUP_NAME %s", group)

	fixture := dns.NewFixture(&solver{},
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/my-custom-solver"),
		dns.SetBinariesPath("kubebuilder/bin"),
	)

	fixture.RunConformance(t)
}
