package main

import "fmt"
import "github.com/hyperledger/fabric/core/chaincode/shim"

func spend(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	//TODO:
	//[ ] args[2] sign
	//[x] check amount of params

	if len(args) != 3 {
		fmt.Println(args)
		return "", fmt.Errorf("Incorrect amount of arguments. Expecting 3, have = %d, %s", len(args), args)
	}

	spender, er := getPemPublicKeyOfCreator(stub)
	if er != nil {
		return "", fmt.Errorf("Cannot get creator of the transaction : %s", er)
	}
	spender = trimPemPubKey(spender)
	fmt.Printf("spender : %s \n", spender)

	fmt.Println("Spend transaction triggered")

	var inputs Inputs
	var outputs Outputs
	ret := "spending ok"

	decode_inputs_fail := decode_io(args[0], &inputs)
	if decode_inputs_fail != nil {
		return "", fmt.Errorf("Error decoding inputs : %s", decode_inputs_fail)
	}
	check_in_fail, total_in, label := check_inputs(stub, inputs, spender)
	if check_in_fail != nil {
		return "", fmt.Errorf("Error checking inputs : %s", check_in_fail)
	}

	decode_outputs_fail := decode_io(args[1], &outputs)
	if decode_outputs_fail != nil {
		return "", fmt.Errorf("Error decoding outputs")
	}

	check_out_fail, total_out := check_outputs(outputs, label)
	if check_out_fail != nil {
		return "", fmt.Errorf("Error checking outputs : %s", check_out_fail)
	}

	if total_in < total_out {
		return "", fmt.Errorf("Error : input amount is smaller than output amount")
	}

	if total_in > total_out {
		return "", fmt.Errorf("Error : input amount is bigger than output amount")
	}

	deletion_fail := delete_inputs(stub, inputs)
	if deletion_fail != nil {
		return "", fmt.Errorf("Error deleting inputs in ledger : %s", deletion_fail)
	}

	keys, fail := get_keys_for_owners(stub, outputs, inputs)
	if fail != nil {

		return "", fmt.Errorf("Error: %s", fail)
	}
	fmt.Printf("%+v", keys)

	txid := stub.GetTxID()
	write_fail := set_outputs(stub, txid, outputs, inputs, &keys, spender)
	if write_fail != nil {
		return "", fmt.Errorf("Error writing new tokens, possible token destruction happened /!\\ ")
	}

	fmt.Printf("%+v", keys)

	fail = delete_old_keys(&keys, inputs)
	if fail != nil {
		return "", fmt.Errorf("Error: %s", fail)
	}
	fail = commit_updated_keys(stub, keys)
	if fail != nil {
		return "", fmt.Errorf("Error: %s", fail)
	}

	return string(ret), nil
}
