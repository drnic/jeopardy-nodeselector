
IMAGE?=docker.io/drnic/jeopardy-nodeselector
TAG?=latest

.PHONY: build push manifest test verify-codegen charts

# docker manifest command will work with Docker CLI 18.03 or newer
# but for now it's still experimental feature so we need to enable that
export DOCKER_CLI_EXPERIMENTAL=enabled

build:
	docker build -t $(IMAGE):$(TAG)-amd64 . -f Dockerfile
	docker build --build-arg OPTS="GOARCH=arm64" -t $(IMAGE):$(TAG)-arm64 . -f Dockerfile
	docker build --build-arg OPTS="GOARCH=arm GOARM=7" -t $(IMAGE):$(TAG)-armhf . -f Dockerfile

push:
	docker push $(IMAGE):$(TAG)-amd64
	docker push $(IMAGE):$(TAG)-arm64
	docker push $(IMAGE):$(TAG)-armhf

manifest:
	docker manifest create --amend $(IMAGE):$(TAG) \
		$(IMAGE):$(TAG)-amd64 \
		$(IMAGE):$(TAG)-arm64 \
		$(IMAGE):$(TAG)-armhf
	docker manifest annotate $(IMAGE):$(TAG) $(IMAGE):$(TAG)-arm64 --os linux --arch arm64
	docker manifest annotate $(IMAGE):$(TAG) $(IMAGE):$(TAG)-armhf --os linux --arch arm --variant v7
	docker manifest push -p $(IMAGE):$(TAG)

test:
	go test ./...
