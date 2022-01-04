#!/usr/bin/env bash

TESTS=$(go test -v -covermode=count -coverprofile=count.txt ./...)
echo "$TESTS"

if echo "$TESTS" | grep -q "FAIL" ; then
  echo ""
  echo "One or more Unit Tests for app have Failed. Build will now fail. Pipeline will also fail..."
  echo ""
  exit 1
else
  echo ""
  echo "All Unit Tests for application have passed!"
  echo "Running Code Coverage..."
  echo ""
  COVERAGE=$(go tool cover -func=./count.txt)
  echo "$COVERAGE"
  exit 0
fi