SHELL         := /usr/bin/env bash
.DEFAULT_GOAL := generate/mocks
.SUFFIXES:

-include .env
export

export COCKROACHDB_URL="cockroachdb://$$DB_USERNAME:$$DB_PASSWORD@$$DB_URL:$$DB_PORT/$$DB_NAME?sslmode=verify-full"
export POSTGRES_URL="postgresql://$$DB_USERNAME:$$DB_PASSWORD@$$DB_URL:$$DB_PORT/$$DB_NAME?sslmode=$$SSL_MODE&sslrootcert=/Users/ricardolopes/.postgresql/global-bundle.pem"

MOCKS_FOLDER=mocks

.PHONY: generate/mocks
generate/mocks: generator/faker.go
	@echo "Generating mocks"
	@rm -rf $(MOCKS_FOLDER)
	@for file in $^; do mockgen -source=$$file -destination=$(MOCKS_FOLDER)/$$file; done

.PHONY: hook/setup
hook/setup:
	pre-commit install

.PHONY: test
test: test/go test/http-requests

.PHONY: test/cover
test/cover:
	TEST_MODE=full go test -coverprofile=c.out -v ./...
	go tool cover -html=c.out -o coverage.html
	go vet -printf=false ./...

.PHONY: test/go
test/go:
	TEST_MODE=full go test -v ./...
	go vet -printf=false ./...

.PHONY: test/http-requests
test/http-requests:
	docker run --rm -i -t -v ./http-requests:/workdir jetbrains/intellij-http-client --env development --env-file http-client.env.json --private-env-file http-client.private.env.json -D rest-api.http

.PHONY: migration/create
migration/create:
	@read -p "Enter migration name: " MIGRATION_NAME; \
	migrate create -ext sql -dir db_migrations -seq $${MIGRATION_NAME}; \

.PHONY: migration/up
migration/up:
	@if [ "${DB_ADAPTOR}" == "cockroach" ]; then \
		read -p "Enter version number: (empty for all) " VERSION_NUMBER; \
		migrate -database ${COCKROACHDB_URL} -path db_migrations up $${VERSION_NUMBER}; \
	elif [ "${DB_ADAPTOR}" == "postgres" ]; then \
		read -p "Enter version number: (empty for all) " VERSION_NUMBER; \
		migrate -database ${POSTGRES_URL} -path db_migrations up $${VERSION_NUMBER}; \
	fi

.PHONY: migration/clean
migration/clean:
	@if [ "${DB_ADAPTOR}" == "cockroach" ]; then \
		read -p "Enter version number: (empty for all) " VERSION_NUMBER; \
		migrate -database ${COCKROACHDB_URL} -path db_migrations down $${VERSION_NUMBER}; \
	elif [ "${DB_ADAPTOR}" == "postgres" ]; then \
		read -p "Enter version number: (empty for all) " VERSION_NUMBER; \
		migrate -database ${POSTGRES_URL} -path db_migrations down $${VERSION_NUMBER}; \
	fi

.PHONY: migration/force
migration/force:
	@if [ "${DB_ADAPTOR}" == "cockroach" ]; then \
		read -p "Enter version number: " VERSION_NUMBER; \
		migrate -database ${COCKROACHDB_URL} -path db_migrations force $${VERSION_NUMBER}; \
	elif [ "${DB_ADAPTOR}" == "postgres" ]; then \
		read -p "Enter version number: " VERSION_NUMBER; \
		migrate -database ${POSTGRES_URL} -path db_migrations force $${VERSION_NUMBER}; \
	fi

.PHONY: image-push
image-push:
	@build-tools/tag.sh

.PHONY: docs
docs:
	@swag init --parseDependency --parseInternal --parseDepth 1
	@rm docs/docs.go

.PHONY: postgres
postgres:
	@docker-compose -f docker-compose.yml  up -d
	@sleep 5
