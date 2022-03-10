#!/bin/bash

# avvia il test network e copia i file necessari alla connessione

export DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export TEST_NETWORK="/home/andrea/go/src/github.com/sapone.andrea/fabric-samples/test-network"

rm -rf "${DIR}/client/connection"
rm -rf "${DIR}/server/connection"
rm -rf "${DIR}/client/wallet"
rm -rf "${DIR}/server/wallet"
mkdir "${DIR}/client/connection"
mkdir "${DIR}/server/connection"
mkdir "${DIR}/client/connection/org1"
mkdir "${DIR}/server/connection/org1"
mkdir "${DIR}/server/connection/org2"
mkdir "${DIR}/client/wallet"
mkdir "${DIR}/server/wallet"

cd "$TEST_NETWORK"
docker network disconnect fabric_test ipfs_host
./network.sh down

./network.sh up createChannel -ca
docker network connect fabric_test ipfs_host

cp "${TEST_NETWORK}/organizations/peerOrganizations/org1.example.com/connection-org1.json" "${DIR}/client/connection/"
cp "${TEST_NETWORK}/organizations/peerOrganizations/org1.example.com/connection-org1.json" "${DIR}/server/connection/"
cp "${TEST_NETWORK}/organizations/peerOrganizations/org2.example.com/connection-org2.json" "${DIR}/server/connection/"