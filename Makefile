.PHONY: test e2e-test run dev build push

IMAGE ?= rrrrdockerrrr/gmb-bot:latest
LOCAL_IMAGE ?= gmb-bot:local
PLATFORMS ?= linux/amd64,linux/arm64
DATA_DIR ?= $(PWD)/gmb_data

test:
	go test ./...
	python3 -m unittest -v scripts.instagram.test_fetch_posts_unit

e2e-test:
	python3 -m unittest -v scripts.instagram.test_fetch_posts_e2e

run:
	mkdir -p $(DATA_DIR)
	docker build -t $(LOCAL_IMAGE) .
	docker run --rm --name gmb-bot \
		-v $(DATA_DIR):/data \
		--env-file .env \
		-e ONE_SHOT=false \
		-e LOG_LEVEL=debug \
		$(LOCAL_IMAGE)

dev:
	mkdir -p $(DATA_DIR)
	docker build -t $(LOCAL_IMAGE) .
	docker run --rm --name gmb-bot-dev \
		-v $(DATA_DIR):/data \
		--env-file .env \
		-e ONE_SHOT=true \
		-e LOG_LEVEL=debug \
		$(LOCAL_IMAGE)

build:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE) .

push:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE) --push .

# use below if say, instaloader repo or version has changed (RUN line does not change, docker will cache the install):
rebuild-push:
	docker buildx build --no-cache --platform $(PLATFORMS) -t $(IMAGE) --push .
