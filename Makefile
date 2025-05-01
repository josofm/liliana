IMG = liliana
VERSION ?= latest
# wd=Get-Location
wd=$(shell pwd)
appvol=$(wd):/app


.PHONY: image-dev
image-dev:
	docker build --target devimage -t liliana-dev .

.PHONY: image-prod
image-prod:
	docker build --target production -t liliana:latest .

.PHONY: run
run: image-dev ##@run Run application on docker compose.
	docker compose up liliana

.PHONY: unit
unit: image-dev ##@unit Run unit tests
	docker run --rm $(IMG) go test -race -timeout 60s -tags unit ./...

.PHONY: start-compose
start-compose:
	docker compose -f docker-compose.yaml up -d

.PHONY: integration
integration: image-dev start-compose ##@run integration tests
	docker compose run --rm -v $(appvol)/app --entrypoint "go test -race -timeout 60s -tags integration ./..." liliana
