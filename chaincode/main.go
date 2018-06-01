package main

import "fmt"
import "github.com/hyperledger/fabric/core/chaincode/shim"
import "github.com/hyperledger/fabric/protos/peer"

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("--------------------> Init <--------------------")

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	var fct		string
	var argv	[]string
	var ret		string
	var err		error

	fmt.Println("--------------------> Invoke <--------------------")
	fct, argv = stub.GetFunctionAndParameters()

	switch fct {
		case "set":
			ret, err = set(stub, argv)
		case "get":
			ret, err = get(stub, argv)
		case "delete":
			ret, err = delete(stub, argv)
		case "spend":
			ret, err = spend(stub, argv)
		case "mint":
			ret, err = mint(stub, argv)
		case "gethistory":
			ret, err = getHistory(stub, argv)
		default:
			return shim.Error("inadequate key sent, aborting")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(ret))
}

func main() {
	var err	error

	if err = shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
