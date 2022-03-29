.PHONY:

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

tool:
	$(GORUN) build/ci.go install ./cmd/tool
	@echo "Done building."
	@echo "Run \"$(GOBIN)/tool\" to launch tool."