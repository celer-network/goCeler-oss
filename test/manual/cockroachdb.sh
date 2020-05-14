#!/bin/sh
set -e

start_db() {
    echo "start cockroach db"
    cockroach start --insecure \
        --listen-addr=localhost:26257 \
        --store=path=/tmp/celer_manual_test/cockroach &
}

stop_db() {
    echo "stop cockroach db"
    pkill cockroach
    rm -rf /tmp/celer_manual_test/cockroach
}

setup_osp_1() {
    echo "create database for OSP 1"
    cat ../../storage/schema.sql | sed 's/celer/celer_test_o1/g' | cockroach sql --insecure
}

setup_osp_2() {
     echo "create database for OSP 2"
    cat ../../storage/schema.sql | sed 's/celer/celer_test_o2/g' | cockroach sql --insecure
}

setup_osp_3() {
     echo "create database for OSP 3"
    cat ../../storage/schema.sql | sed 's/celer/celer_test_o3/g' | cockroach sql --insecure
}

setup_osp_4() {
     echo "create database for OSP 4"
    cat ../../storage/schema.sql | sed 's/celer/celer_test_o4/g' | cockroach sql --insecure
}

setup_osp_5() {
     echo "create database for OSP 5"
    cat ../../storage/schema.sql | sed 's/celer/celer_test_o5/g' | cockroach sql --insecure
}

arg="${1}"
case ${arg} in
    start)  start_db
            ;;
    stop)   stop_db
            ;;
    1)      setup_osp_1
            ;;
    2)      setup_osp_2
            ;;
    3)      setup_osp_3
            ;;
    4)      setup_osp_4
            ;;
    5)      setup_osp_5
            ;;
    *)      echo "invalid arg"
esac