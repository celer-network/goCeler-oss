#!/bin/sh
set -e

STOREPATH="${HOME}/celerdb" # replace this to your preferred location

start_db() {
    echo "start cockroach db, store path: ${STOREPATH}"
    cockroach start --insecure --listen-addr=localhost:26257 --store=path=${STOREPATH} &
    # Ensure CockroachDB is up
    sleep 2
    cat ../../storage/schema.sql | cockroach sql --insecure
}

stop_db() {
    echo "stop cockroach db"
    pkill cockroach
}

op="${1}"
case ${op} in
    start)  start_db
            ;;
    stop)   stop_db
            ;;
    *)      echo "invalid arg"
esac