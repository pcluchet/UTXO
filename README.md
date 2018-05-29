# UTXO

## Description
	This project is an implementation of UTXO (Unspent Transaction Output), the Data Model introduced by Bitcoin.
	Coded in Golang, on top of Hyperledger Fabric.
	[Whitepaper](https://arxiv.org/pdf/1801.10228.pdf) (Chapter 5)

## Installation and Usage

	This will create and run the network:

	```
	git clone https://github.com/pcluchet/UTXO.git
	cd UTXO
	./network/byfn.sh up
	```

	If you want to play with it you need to create a Money supply, to do so:

	```
	docker exec -t alice peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem  -C ptwist -n Ptwist -c '{"Args":["mint", "sign", "[{\"amount\":42.42,\"owner\":\"Alice\",\"label\":\"USD\"}]"]}'
	```

## Usage Example

	```
	Check Alice balance: docker exec -t alice balance
	Send money to Bob from Alice: docker exec -t bob spend (amount, owner, label)
	Get you public key: docker exec -t bob address
	```

## Author
Pierre Cluchet [pcluchet](https://github.com/pcluchet) 🐝
Sebastien Huertas [cactusfluo](https://gitlab.com/cactusfluo) 🦍
Jefferson Le Quellec [jle-quel](https://github.com/jle-quel) 🐜