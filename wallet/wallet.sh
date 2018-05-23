#!/bin/bash

set -e
CHANNEL_NAME="mychannel"
CHAINCODE_NAME="mycc"

function getPublicKey {
	privateKey=$(ls $CORE_PEER_MSPCONFIGPATH/keystore)
	publicKey=$(openssl ec -in $CORE_PEER_MSPCONFIGPATH/keystore/$privateKey 2>&-)
	echo $publicKey | tr -d '\040\011\012\015' | cut -c28- | cut -c -162
}

function balance {
	echo "Balance for [$1]"
	balance=$(peer chaincode query -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem  -C $CHANNEL_NAME -n $CHAINCODE_NAME -c `{"Args":["get", "$1"]}`)
}


if [ $1 ]; then
	publicKey=$(getPublicKey)
	if [ $1 == "balance" ]; then
		balance $publicKey
	fi
else
	echo "Usage: wallet [ comming soon ]"
fi
