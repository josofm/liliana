IMG = liliana
VERSION ?= latest
GHCR_REGISTRY ?= ghcr.io
GHCR_OWNER ?= josofm
GHCR_IMAGE = $(GHCR_REGISTRY)/$(GHCR_OWNER)/$(IMG)
wd=$(shell cd)
appvol=$(wd):/app


.PHONY: image-dev
image-dev:
	docker build --target devimage -t liliana-dev .

.PHONY: image-prod
image-prod:
	docker build --target production -t $(IMG):$(VERSION) .

.PHONY: publish
publish: image-prod ##@publish Build and publish the production image to GHCR.
	docker tag $(IMG):$(VERSION) $(GHCR_IMAGE):$(VERSION)
	docker push $(GHCR_IMAGE):$(VERSION)

.PHONY: run
run: image-dev ##@run Run application on docker compose.
	docker compose up liliana

.PHONY: unit
unit: image-dev ##@unit Run unit tests
	docker run --rm liliana-dev go test -race -timeout 60s -tags unit ./...

.PHONY: start-compose
start-compose:
	docker compose -f docker-compose.yaml up -d postgres

.PHONY: stop-compose
stop-compose:
	docker compose -f docker-compose.yaml down --remove-orphans

.PHONY: migrate
migrate:
	go run ./cmd/migrate up

.PHONY: compose-migrate
compose-migrate:
	docker compose -f docker-compose.yaml run --rm migrate

.PHONY: integration
integration:
	go run ./hack/integration


