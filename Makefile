.PHONY: build up down run local docker clean clean-all wrk-local vegeta-local wrk-docker vegeta-docker test-local test-docker

# Local targets
run:
	cd cmd/app && go run main.go

notify:
	cd cmd/notifier && go run publish_to_kafka.go
	
local: run

# Docker targets
build:
	APP_ENV=docker docker-compose build

up:
	APP_ENV=docker docker-compose up -d

down:
	docker-compose down

clean: down
	docker-compose rm -f

clean-all: clean
	docker-compose down --rmi all --volumes --remove-orphans

docker: build up

# Targets for local integration tests
test-integration:
	@echo "Running integration tests"
	@go test -coverprofile=coverage.out ./... -v
	go tool cover -html=coverage.out

# Goals for integration tests in Docker
test-integration-docker:
	@echo "Running integration tests in Docker"
	docker-compose run app go test ./... -v

# Local stress-tests with WRK and Vegeta
wrk-local:
	@echo "Running wrk test on /order endpoint (localhost)"
	wrk -t12 -c400 -d30s http://localhost:8080/order?id=1

vegeta-local:
	@echo "Running vegeta test on /order endpoint (localhost)"
	echo "GET http://localhost:8080/order?id=1" | vegeta attack -duration=30s -rate=1000 | tee results-local.bin | vegeta report
	@echo "Generating vegeta JSON report and plot for local test"
	vegeta report -type=json results-local.bin > metrics-local.json
	vegeta plot results-local.bin > plot-local.html

# Docker stress-tests with WRK and Vegeta
wrk-docker:
	@echo "Running wrk test on /order endpoint (Docker)"
	wrk -t12 -c400 -d30s http://172.17.0.1:8080/order?id=1

vegeta-docker:
	@echo "Running vegeta test on /order endpoint (Docker)"
	echo "GET http://172.17.0.1:8080/order?id=1" | vegeta attack -duration=30s -rate=1000 | tee results-docker.bin | vegeta report
	@echo "Generating vegeta JSON report and plot for Docker test"
	vegeta report -type=json results-docker.bin > metrics-docker.json
	vegeta plot results-docker.bin > plot-docker.html

# Comprehensive targets for local and docker environments
test-local: wrk-local vegeta-local
	@echo "Completed local tests with wrk and vegeta"
	@echo "Local vegeta metrics saved to metrics-local.json and plot to plot-local.html"
	@open plot-local.html

test-docker: wrk-docker vegeta-docker
	@echo "Completed Docker tests with wrk and vegeta"
	@echo "Docker vegeta metrics saved to metrics-docker.json and plot to plot-docker.html"
	open plot-docker.html
