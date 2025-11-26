#!/usr/bin/env bash

IMAGE=${IMAGE:=quay.io/fromani/dra-driver-memory:v0.0.2025112401}
COMMAND=${COMMAND:=/bin/dramemory}
BIN_PATH=./bin/dra-deployer

echo "Deploying dra-driver-memory using the following command: ./bin/dra-deployer apply -i ${IMAGE} --command ${COMMAND}"
${BIN_PATH} apply -i "${IMAGE}" --command "${COMMAND}"