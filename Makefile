export BUILD_DATE         := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
export VCS_REF            := $(shell git rev-parse HEAD)
export IMAGE_TAG 		  := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)

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