#!/bin/sh

response=$(echo "$TEST_TEXT" | nc -N server "$PORT")

if [ "$response" = "$TEST_TEXT" ]; then
  echo "$SUCCESS_MESSAGE"
else
  echo "$FAIL_MESSAGE"
fi
