export BUILD_DATE         := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
export VCS_REF            := $(shell git rev-parse HEAD)
export IMAGE_TAG 		  := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)
GOBIN ?= $$(go env GOPATH)/bin

run-help:
	go run main.go

run-serve:
	go run main.go serve

.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 \
	docker build -f .docker/Dockerfile.distroless \
	--build-arg=COMMIT=$(VCS_REF) \
	--build-arg=BUILD_DATE=$(BUILD_DATE) \
	-t arwoosa/notifaction:${IMAGE_TAG} .


.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml

.PHONY: check-golangci
check-golangci:
	golangci-lint run router/... service/... cmd/...
