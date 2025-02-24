# list available receipes
@default:
  just --list

@_prepare:
  tar -xvzf geolite2.tgz

# run regular golang tests
test-go:
  go test -v -cover ./...

@_clean-yaegi:
  rm -rf /tmp/yaegi*

# run tests via yaegi
test-yaegi: && _clean-yaegi
  #!/usr/bin/env bash

  TMP=$(mktemp -d yaegi.XXXXXX -p /tmp)
  WRK="${TMP}/go/src/github.com/cmmoran"
  mkdir -p ${WRK}
  ln -s `pwd` "${WRK}"
  cd "${WRK}/$(basename `pwd`)"
  echo "Testing with yaegi"
  env GOPATH="${TMP}/go" yaegi test -v .

# lint and test
test: _prepare test-go test-yaegi

clean:
  rm -rf *.mmdb
