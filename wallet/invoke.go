package main

import "fmt"
import "os/exec"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func	getInvokeCommand(input Transaction, argv []string) string {
	return fmt.Sprintf(`peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["spend", "[{\"txid\":\"%s\", \"j\":%d}]", "[{\"amount\":%s,\"owner\":\"%s\",\"label\":\"%s\"}]", "sigs"]}'`, input.Txid, input.J, argv[1], argv[2], argv[3])
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	makeTransaction(tx []Transaction, balance []Balance, argv []string) error {
	var cmd		string
	var input	Transaction
	var err		error
	
	if input, err = checkFund(tx, balance, argv); err != nil {
		return fmt.Errorf("Insufficent fund for this transaction")
	}
	cmd = getInvokeCommand(input, argv)
	if _, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
		return err
	}

	return nil
}
