OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)
VERSION ?= 1.0.0

# install - compiles the terraform provider to the appropriate plugin location until we can upload and release to the
#           Terraform registry.
install:
	go build -o $(HOME)/.terraform.d/plugins/terraform.local/local/storj/$(VERSION)/$(OS)_$(ARCH)/terraform-provider-storj_v$(VERSION) .

# docs - generates documentation for the Terraform provider using the terraform-plugin-docs project.
#        https://github.com/hashicorp/terraform-plugin-docs
docs: .docs
.docs:
	tfplugindocs generate
