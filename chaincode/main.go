package main

import "fmt"
import "github.com/hyperledger/fabric/core/chaincode/shim"
import "github.com/hyperledger/fabric/protos/peer"

/* ************************************************************************** */
/*	PUBLIC																	  */
/* ************************************************************************** */

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	fmt.Printf("Invoque Request\n")

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "get" {

		result, err = get(stub, args)
	} else if fn == "gethistory" {

		result, err = gethistory(stub, args)
	} else if fn == "spend" {

		result, err = spend(stub, args)
	} else if fn == "mint" {

		result, err = mint(stub, args)
	} else if fn == "delete" {

		result, err = delete(stub, args)
	} else if fn == "getUnspentForUser" {

		result, err = getUnspentForUser(stub, args)
	} else {

		return shim.Error("inadequate key sent, aborting")
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
