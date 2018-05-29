package main

import "fmt"
import "strconv"
import "os/exec"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

/*func	getPartCommand(input Transaction, argv []string) string {
	
}*/

func	getInvokeCommand(input Transaction, output Balance, argv []string) string {
	var sum	[]float64

	sum = getSum(argv[1], output.Amount)

	if sum[1] == 0 {
		return fmt.Sprintf(`peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C ptwist -n Ptwist -c '{"Args":["spend", "[{\"txid\":\"%s\", \"j\":%d}]", "[{\"amount\":%s,\"owner\":\"%s\",\"label\":\"%s\"}]", "sigs"]}'`, input.Txid, input.J, argv[1], argv[2], argv[3])
	}
	
	return fmt.Sprintf(`peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C ptwist -n Ptwist -c '{"Args":["spend", "[{\"txid\":\"%s\", \"j\":%d}]", "[{\"amount\":%s,\"owner\":\"%s\",\"label\":\"%s\"}, {\"amount\":%f,\"owner\":\"%s\",\"label\":\"%s\"}]]", "sigs"]}'`, input.Txid, input.J, argv[1], argv[2], argv[3], sum[1], output.Owner, output.Label)
}

func	getSum(amount string, fund float64) []float64 {
	var dec	float64

	dec, _= strconv.ParseFloat(amount, 64)
	
	return []float64{fund, fund - dec}
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	makeTransaction(tx []Transaction, balance []Balance, argv []string) error {
	var cmd		string
	var input	Transaction
	var output	Balance
	var err		error
	
	if input, output, err = checkFund(tx, balance, argv); err != nil {
		return fmt.Errorf("Insufficent fund for this transaction")
	}

	cmd = getInvokeCommand(input, output, argv)
	if _, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
		return err
	}

	return nil
}
