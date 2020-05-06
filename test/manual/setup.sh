#!/bin/sh
rm -rf /tmp/celer_manual_test
go run ${GOCELER}/test/manual/setup.go -logcolor
