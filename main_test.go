package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"

	// "bytes"
	// "log"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
	group = os.Getenv("GROUP_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	// var buf bytes.Buffer
	// log.SetOutput(&buf)
	// defer func () {
	// 	log.SetOutput(os.Stderr)
	// }()

	t.Logf("TEST_ZONE_NAME %s", zone)
	t.Logf("GROUP_NAME %s", group)

	fixture := dns.NewFixture(&customDNSProviderSolver{},
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/my-custom-solver"),
		dns.SetBinariesPath("__main__/hack/bin"),
	)

	fixture.RunConformance(t)

	// t.Log(buf.String())
	// t.Log("END")
}
