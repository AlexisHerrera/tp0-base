#!/bin/bash

result=$(docker run --rm --env-file ./netcat-test/config.env --network tp0_testing_net netcat-test:latest | tail -n 1)
echo "$result"