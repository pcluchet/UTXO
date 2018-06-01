package main

import "fmt"
import "github.com/hyperledger/fabric/core/chaincode/shim"

/* ************************************************************************** */
/*	PRIVATE																	  */
/* ************************************************************************** */

/* ************************************************************************** */
/*	PUBLIC																	  */
/* ************************************************************************** */

func mint(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//TODO:
	//Arg[0] is a key to identify CB
	//Arg[2] is sign

	fmt.Println("Minting")
	var outputs Outputs
	ret := "coin minted"

	decode_outputs_fail := decode_io(args[1], &outputs)
	if decode_outputs_fail != nil {
		return "", fmt.Errorf("Error in output decoding : %s", decode_outputs_fail)
	}

	fmt.Println("\n\n\n\n\nGetting keys\n\n\n\n")
	keys, fail := get_keys_for_owners(stub, outputs, nil)
	if fail != nil {
		return "", fmt.Errorf("Error:  %s", fail)
	}

	txid := stub.GetTxID()
	fmt.Println("Minting with txid=", txid)
	set_outputs_fail := set_outputs(stub, txid, outputs, nil, &keys, "")
	if set_outputs_fail != nil {
		fmt.Println("Error decoding outputs")
		return "", fmt.Errorf("Error in outputs : %s", set_outputs_fail)
	}
	err := commit_updated_keys(stub, keys)
	if err != nil {
		return "", fmt.Errorf("Err : %s", err)
	}

	return string(ret), nil
}

