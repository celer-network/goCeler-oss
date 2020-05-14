#!/bin/sh

basic_setup() {
    echo "setup testnet"
    rm -rf /tmp/celer_manual_test
    go run ${GOCELER}/test/manual/setup.go -logcolor
}

auto_setup() {
    echo "setup testnet, automatically add/approve fund and register osps"
    rm -rf /tmp/celer_manual_test
    go run ${GOCELER}/test/manual/setup.go -logcolor -auto
}

arg="${1}"
case ${arg} in
    auto)   auto_setup
            ;;
    *)      basic_setup
            ;;
esac