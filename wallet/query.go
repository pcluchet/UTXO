package main

import "fmt"
import "os/exec"
import "encoding/json"

type Balance		struct {
	Amount	float64
	Owner	string
	Label	string
}

type Transaction	struct {
	Txid	string
	J		int
}

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func	getQueryCommand(owner string) string {
	return fmt.Sprintf(`peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem  -C ptwist -n Ptwist -c '{"Args":["get", "%s"]}'`, owner)
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	queryTransaction(publicKey string) ([]Transaction, error) {
	var stdout		[]byte
	var tx			[]Transaction
	var ret			string
	var cmd			string
	var err			error

	cmd = getQueryCommand(publicKey)
	if stdout, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("Asset not found for [%s]", publicKey)
	}
	ret = parseStdout(string(stdout))
	if err = json.Unmarshal([]byte(ret), &tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func	queryBalance(tx []Transaction) ([]Balance, error) {
	var stdout		[]byte
	var balance		[]Balance
	var buf			Balance
	var ret			string
	var cmd			string
	var err			error

	for _, value := range tx {
		cmd = getQueryCommand(fmt.Sprintf("%s_%d", value.Txid, value.J))
		if stdout, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
			return nil, err
		}
		ret = parseStdout(string(stdout))
		if err = json.Unmarshal([]byte(ret), &buf); err != nil {
			return nil, err
		}
		balance = append(balance, buf)
	}

	return balance, nil
}
