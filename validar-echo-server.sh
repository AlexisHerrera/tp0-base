#!/bin/bash

result=$(docker compose -f docker-compose-dev.yaml run --rm netcat-test 2>&1 | tail -n 1)
echo "$result"
