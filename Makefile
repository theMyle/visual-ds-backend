# load env variables
include .env
export 

BINARY_NAME=api

.PHONY: help up down status

help:
	@echo Available commands:
	@echo 	build			- build go project
	@echo 	run 			- run go project
	@echo 	build-run 		- builds and run go project
	@echo 	clean			- cleans build directory
	@echo 	sqlc			- generate sql query go bindings
	@echo 	up  			- run pending migrations
	@echo 	down			- rollback the last migration
	@echo 	status			- show which migrations have been applied

# build go server
build:
	@go build -o ./bin/$(BINARY_NAME) ./cmd/api/main.go

# run go server
run:
	@./bin/$(BINARY_NAME)

build-run: build run

# clean build dir
clean:
	@rm -rf bin/

# generates sql query go bindings
sqlc:
	@sqlc generate

# goose migration up by one
up:
	@goose postgres "$(DB_URL_IPV4)" -dir ./sql/schemas up

# goose migration down by one
down:
	@goose postgres "$(DB_URL_IPV4)" -dir ./sql/schemas down

# goose status
status:
	@goose postgres "$(DB_URL_IPV4)" -dir ./sql/schemas status
