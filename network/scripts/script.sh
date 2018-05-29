#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build your first network (BYFN) end-to-end test"
echo
DELAY="3"
TIMEOUT="10"
CHANNEL_NAME="ptwist"
LANGUAGE="golang"
LANGUAGE=`echo "$LANGUAGE" | tr [:upper:] [:lower:]`
COUNTER=1
MAX_RETRY=5
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
CHAINCODE=Ptwist

CC_SRC_PATH="github.com/chaincode/"
if [ "$LANGUAGE" = "node" ]; then
	CC_SRC_PATH="/opt/gopath/src/github.com/chaincode/chaincode_example02/node/"
fi

echo "Channel name : "$CHANNEL_NAME

# import utils
. scripts/utils.sh

createChannel() {
	echo
	echo "--------------------------------------------> $CORE_PEER_ADDRESS <--------------------------------------------------"
	echo

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
		peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
		res=$?
                set +x
	else
				set -x
		peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
		res=$?
				set +x
	fi
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
	echo
}

joinChannel () {
	for org in MEDSOS ; do
		for peer in 0 1 ; do
			joinChannelWithRetry $peer $org
			echo "===================== peer${peer}.${org} joined on the channel \"$CHANNEL_NAME\" ===================== "
			sleep $DELAY
			echo
		done
	done
}

## Create channel
echo "Creating channel..."
setGlobals 0 MEDSOS
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

exit 0

## Set the anchor peers for each org in the channel
echo "Updating anchor peers for MEDSOS..."
updateAnchorPeers 0 "MEDSOS"

## Install chaincode on peer0.MEDSOS
echo "Installing chaincode on peer0.MEDSOS..."
installChaincode 0 "MEDSOS" 

# Instantiate chaincode on peer0.MEDSOS
echo "Instantiating chaincode on peer0.MEDSOS..."
instantiateChaincode 0 "MEDSOS"

<< --MULTILINE-COMMENT--
# Query chaincode on peer0.org1
echo "Querying chaincode on peer0.org1..."
chaincodeQuery 0 1 100

# Invoke chaincode on peer0.org1
echo "Sending invoke transaction on peer0.org1..."
chaincodeInvoke 0 1
--MULTILINE-COMMENT--

echo
echo "========= All GOOD, BYFN execution completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
