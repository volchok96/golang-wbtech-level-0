.PHONY: build up down run local docker clean clean-all wrk-local vegeta-local wrk-docker vegeta-docker test-local test-docker

build:
	APP_ENV=docker docker-compose build

up:
	APP_ENV=docker docker-compose up -d

down:
	docker-compose down

run:
	cd cmd && go run main.go

local: run

docker: build up

clean: down
	docker-compose rm -f

clean-all: clean
	docker-compose down --rmi all --volumes --remove-orphans

# Локальные тесты с использованием localhost
wrk-local:
	@echo "Running wrk test on /order endpoint (localhost)"
	wrk -t12 -c400 -d30s http://localhost:8080/order?id=1

vegeta-local:
	@echo "Running vegeta test on /order endpoint (localhost)"
	echo "GET http://localhost:8080/order?id=1" | vegeta attack -duration=30s -rate=1000 | tee results-local.bin | vegeta report
	@echo "Generating vegeta JSON report and plot for local test"
	vegeta report -type=json results-local.bin > metrics-local.json
	vegeta plot results-local.bin > plot-local.html

# Docker тесты с использованием host.docker.internal (Docker)
wrk-docker:
	@echo "Running wrk test on /order endpoint (Docker)"
	wrk -t12 -c400 -d30s http://172.17.0.1:8080/order?id=1

vegeta-docker:
	@echo "Running vegeta test on /order endpoint (Docker)"
	echo "GET http://172.17.0.1:8080/order?id=1" | vegeta attack -duration=30s -rate=1000 | tee results-docker.bin | vegeta report
	@echo "Generating vegeta JSON report and plot for Docker test"
	vegeta report -type=json results-docker.bin > metrics-docker.json
	vegeta plot results-docker.bin > plot-docker.html

# Комплексные цели для локального и docker окружения
test-local: wrk-local vegeta-local
	@echo "Completed local tests with wrk and vegeta"
	@echo "Local vegeta metrics saved to metrics-local.json and plot to plot-local.html"
	@open plot-local.html

test-docker: wrk-docker vegeta-docker
	@echo "Completed Docker tests with wrk and vegeta"
	@echo "Docker vegeta metrics saved to metrics-docker.json and plot to plot-docker.html"
	open plot-docker.html
