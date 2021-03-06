#!/bin/sh

run_osp_1() {
  echo "run OSP 1"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o1_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp1.json \
    -port 10001 \
    -adminrpc localhost:11001 \
    -adminweb localhost:8190 \
    -svrname o1 \
    -logprefix o1 \
    -storedir /tmp/celer_manual_test/store \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_1_crdb() {
  echo "run OSP 1 w/ cockroach db"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o1_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp1.json \
    -port 10001 \
    -adminrpc localhost:11001 \
    -adminweb localhost:8190 \
    -svrname o1 \
    -logprefix o1 \
    -storesql "postgresql://celer_test_o1@localhost:26257/celer_test_o1?sslmode=disable" \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_2() {
  echo "run OSP 2"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o2_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp2.json \
    -port 10002 \
    -adminrpc localhost:11002 \
    -adminweb localhost:8290 \
    -svrname o2 \
    -logprefix o2 \
    -storedir /tmp/celer_manual_test/store \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_2_crdb() {
  echo "run OSP 2 w/ cockroach db"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o2_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp2.json \
    -port 10002 \
    -adminrpc localhost:11002 \
    -adminweb localhost:8290 \
    -svrname o2 \
    -logprefix o2 \
    -storesql "postgresql://celer_test_o2@localhost:26257/celer_test_o2?sslmode=disable" \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_3() {
  echo "run OSP 3"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o3_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp3.json \
    -port 10003 \
    -adminrpc localhost:11003 \
    -adminweb localhost:8390 \
    -svrname o3 \
    -logprefix o3 \
    -storedir /tmp/celer_manual_test/store \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_3_crdb() {
  echo "run OSP 3 w/ cockroach db"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o3_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp3.json \
    -port 10003 \
    -adminrpc localhost:11003 \
    -adminweb localhost:8390 \
    -svrname o3 \
    -logprefix o3 \
    -storesql "postgresql://celer_test_o3@localhost:26257/celer_test_o3?sslmode=disable" \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_4() {
  echo "run OSP 4"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o4_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp4.json \
    -port 10004 \
    -adminrpc localhost:11004 \
    -adminweb localhost:8490 \
    -svrname o4 \
    -logprefix o4 \
    -storedir /tmp/celer_manual_test/store \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_4_crdb() {
  echo "run OSP 4 w/ cockroach db"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o4_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp4.json \
    -port 10004 \
    -adminrpc localhost:11004 \
    -adminweb localhost:8490 \
    -svrname o4 \
    -logprefix o4 \
    -storesql "postgresql://celer_test_o4@localhost:26257/celer_test_o4?sslmode=disable" \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_5() {
  echo "run OSP 5"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o5_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp5.json \
    -port 10005 \
    -adminrpc localhost:11005 \
    -adminweb localhost:8590 \
    -svrname o5 \
    -logprefix o5 \
    -storedir /tmp/celer_manual_test/store \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

run_osp_5_crdb() {
  echo "run OSP 5 w/ cockroach db"
  go run ${GOCELER}/server/server.go \
    -profile /tmp/celer_manual_test/profile/o5_profile.json \
    -ks ${GOCELER}/testing/env/keystore/osp5.json \
    -port 10005 \
    -adminrpc localhost:11005 \
    -adminweb localhost:8590 \
    -svrname o5 \
    -logprefix o5 \
    -storesql "postgresql://celer_test_o5@localhost:26257/celer_test_o5?sslmode=disable" \
    -rtc ${GOCELER}/test/manual/rt_config.json \
    -nopassword \
    -logcolor
}

osp="${1}"
case ${osp} in
  1)      run_osp_1
          ;;
  2)      run_osp_2
          ;;
  3)      run_osp_3
          ;;
  4)      run_osp_4
          ;;
  5)      run_osp_5
          ;;
  1_crdb) run_osp_1_crdb
          ;;
  2_crdb) run_osp_2_crdb
          ;;
  3_crdb) run_osp_3_crdb
          ;;
  4_crdb) run_osp_4_crdb
          ;;
  5_crdb) run_osp_5_crdb
          ;;
  *)  echo "please specify OSP [1-5] or [1-5]_crdb"
esac