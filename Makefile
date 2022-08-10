OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)
VERSION ?= 1.0.0

install:
	go build -o $(HOME)/.terraform.d/plugins/terraform.local/local/storj/$(VERSION)/$(OS)_$(ARCH)/terraform-provider-storj_v$(VERSION) .
