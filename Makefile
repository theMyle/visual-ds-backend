# load env variables
include .env
export 

.PHONY: help up down status

help:
	@echo Available commands:
	@echo 	make up  		- run pending migrations
	@echo 	make down		- rollback the last migration
	@echo 	make status		- show which migrations have been applied

# goose migration up by one
up:
	goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema up

# goose migration down by one
down:
	goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema down

# goose status
status:
	goose postgres "$(CONNECTION_STRING_IPV4)" -dir ./sql/schema status