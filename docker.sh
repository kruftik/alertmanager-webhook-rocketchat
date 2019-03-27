#!/usr/bin/env bash
# ./docker.sh <action>

SRC_DIR=bitbucket.org/fxadmin/fxinnovation-itoa-application-cloudcustomchecks

set -o errexit
set -o pipefail
set -o nounset

GOLANG_VERSION="1.10.0-stretch"

docker run --rm -i -v "$PWD":/go/src/${SRC_DIR} -w /go/src/${SRC_DIR} golang:${GOLANG_VERSION} << TOOLSEOF
apt-get update -y
apt-get upgrade -y
apt-get install -y bash make git g++
$@
TOOLSEOF
