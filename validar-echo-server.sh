#!/bin/bash

COMPOSE_FILE="docker-compose-dev.yaml"

if [ -n "$(docker compose -f $COMPOSE_FILE ps -q)" ]; then
  result=$(docker compose -f $COMPOSE_FILE run --rm netcat-test 2>&1 | tail -n 1)
  echo "$result"
fi