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
	docker build --progress=plain --tag $(IMG) --target=test-unit .
