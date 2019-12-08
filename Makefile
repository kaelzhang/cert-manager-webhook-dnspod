IMAGE_NAME := "kaelz/cert-manager-webhook-dnspod"
IMAGE_TAG := "1.0.0"

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

test:
	go test -v .

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push:
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template cert-manager-webhook-dnspod \
        --set image.repository=$(IMAGE_NAME) \
        --set image.tag=$(IMAGE_TAG) \
        charts > "$(OUT)/rendered-manifest.yaml"
