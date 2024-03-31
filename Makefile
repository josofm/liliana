IMG = liliana
VERSION ?= latest
wd=Get-Location
# wd=$(shell pwd) use me in linux
appvol=$(wd):/app

.PHONY: image
image: ##@image create dev image
	docker build --progress=plain --target devimage . -t $(IMG)

.PHONY: run
run: image ##@run Run application on docker compose.
	docker compose run --rm -v $(appvol) --service-ports  --entrypoint "go run /app/cmd/liliana.go" liliana

.PHONY: unit
unit: image ##@unit Run unit tests
	docker run --rm $(IMG) go test -race -timeout 60s -tags unit ./...

.PHONY: start-compose
start-compose:
	docker compose -f docker-compose.yaml up -d

.PHONY: integration
integration: image start-compose ##@run integration tests
	docker compose run --rm -v $(appvol)/app --entrypoint "go test -race -timeout 60s -tags integration ./..." liliana
