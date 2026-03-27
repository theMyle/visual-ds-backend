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

# build go server
build:
	@go build -o ./bin/$(BINARY_NAME) ./cmd/api/main.go

# run go server
run:
	@./bin/$(BINARY_NAME)

# clean build dir
clean:
	@rm -rf bin/

# goose migration up by one
up:
	@goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema up

# goose migration down by one
down:
	@goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema down

# goose status
status:
<<<<<<< HEAD
	@goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema status
=======
	@goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema status
>>>>>>> 6758e25ec20fa4878db4f31c11f5009bd03bec80
