.PHONY: build up down run local docker clean clean-all

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

run:
	APP_ENV=local cd cmd && go run main.go

local: run

docker: build up run

clean: down
	docker-compose rm -f

clean-all: clean
	docker-compose down --rmi all --volumes --remove-orphans

