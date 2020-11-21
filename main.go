package main

import (
	"os"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"github.com/kaelzhang/cert-manager-webhook-dnspod/webhook"
	"k8s.io/klog"
)

func main() {
	groupName := os.Getenv("GROUP_NAME")

	if groupName == "" {
		klog.Fatal("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided groupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(groupName, webhook.Solver())
}
