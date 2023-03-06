

.PHONY: docker-infra
docker-infra:
	@echo "Starting server in docker container..."
	@docker-compose -p infra -f build/dev/docker-compose-infra.yml up -d

.PHONY: docker-infra-down
docker-infra-down:
	 @echo "Down server in docker container..."
	 @docker-compose -p infra -f build/dev/docker-compose-infra.yml down -v

.PHONY: docker-infra-ps
docker-infra-ps:
	@docker-compose -p infra -f build/dev/docker-compose-infra.yml ps

.PHONY: docker-infra-logs
docker-infra-logs:
	@docker-compose -p infra -f build/dev/docker-compose-infra.yml logs -f

.PHONY: docker-dev
docker-dev:
	@docker compose -f build/dev/docker-compose.yml up --build

.PHONY: install-all
install-all: install-mockgen

# Mockgen
.PHONY: install-mockgen
install-mockgen:
	go install github.com/golang/mock/mockgen@v1.6.0

.PHONY: gen-mock
gen-mock:
	go generate ./...

# Get test coverage
.PHONY: test-coverage
test-coverage:
	@echo "Run test with coverage"
	@go test -p 1  ./internal/... ./cmd/... -cover -count=1 -coverprofile cover_full.out
	@go tool cover -func cover_full.out | grep "^total" | awk '{print $3}'