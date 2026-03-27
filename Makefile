# load env variables
include .env
export 

BINARY_NAME=api

.PHONY: help up down status

help:
	@echo Available commands:
	@echo	build			- build go project
	@echo	run				- run go project
	@echo	clean			- cleans build directory
	@echo 	make up  		- run pending migrations
	@echo 	make down		- rollback the last migration
	@echo 	make status		- show which migrations have been applied

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
	@goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema status