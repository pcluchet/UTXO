/*
 * UTXO implementation
 */

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

type Input struct {
	Txid string
	J    int
}

type Inputs []Input

type Output struct {
	Amount json.Number
	Owner  string
	Label  string
}

type Outputs []Output

type UserUnspents struct {
	User     string
	Unspents []byte
}

const (
	shift = 64 - 11 - 1
	bias  = 1023
	mask  = 0x7FF
)

// Round returns the nearest integer, rounding half away from zero.
// This function is available natively in Go 1.10
//
// Special cases are:
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round(x float64) float64 {
	// Round is a faster implementation of:
	//
	// func Round(x float64) float64 {
	//   t := Trunc(x)
	//   if Abs(x-t) >= 0.5 {
	//     return t + Copysign(1, x)
	//   }
	//   return t
	// }
	const (
		signMask = 1 << 63
		fracMask = 1<<shift - 1
		half     = 1 << (shift - 1)
		one      = bias << shift
	)

	bits := math.Float64bits(x)
	e := uint(bits>>shift) & mask
	if e < bias {
		// Round abs(x) < 1 including denormals.
		bits &= signMask // +-0
		if e == bias-1 {
			bits |= one // +-1
		}
	} else if e < bias+shift {
		// Round any abs(x) >= 1 containing a fractional component [0,1).
		//
		// Numbers with larger exponents are returned unchanged since they
		// must be either an integer, infinity, or NaN.
		e -= bias
		bits += half >> e
		bits &^= fracMask >> e
	}
	return math.Float64frombits(bits)
}

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

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}

	a := make([]string, 3)

	a[0] = "case0"
	a[1] = "case 1"

	value, errr := stub.CreateCompositeKey("objtt", a)
	fmt.Printf(value)
	//vl := stub.SplitCompositeKey(value);

	//fmt.Printf(vl);

	if errr != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

func delete(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.DelState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to delete asset: %s", args[0])
	}

	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

// Get returns the value of the specified asset key
func getUnspentForUser(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a user")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

func gethistory(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetHistoryForKey(args[0])

	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	var history string
	history = "\n"

	for value.HasNext() {
		history = fmt.Sprintf("%s%s", history, fmt.Sprintln(value.Next()))
	}

	return string(history), nil
}

func decode_io(arg string, adress interface{}) error {

	b := bytes.NewReader([]byte(arg))
	fmt.Printf("arg = %s\n", arg)

	err := json.NewDecoder(b).Decode(adress)

	if err != nil {
		return fmt.Errorf("Error %s", err)
	}

	if adress == nil {
		return fmt.Errorf("nil adress given")
	}

	fmt.Printf("Parsed at io level : %+v", adress)
	return nil
}

func delete_inputs(stub shim.ChaincodeStubInterface, inputs Inputs) error {

	fmt.Println("Now deleting inputs")

	var i int
	i = 0
	for k, v := range inputs {
		fmt.Println("Handling now : key[%s] value[%s]\n", k, v)
		retreiving_key := fmt.Sprintf("%s_%d", v.Txid, v.J)

		fmt.Println("DELETED KEY : %s", retreiving_key)
		err := stub.DelState(retreiving_key)
		if err != nil {
			return fmt.Errorf("Error %s", err)
		}
		i++
	}
	return nil
}

// Given a owner and an output list, this function returns the index of the owner
// in the outputs list, returns -1 if user is not found
func IndexOfUserInOutputs(user string, outputs Outputs) int {

	for k, v := range outputs {
		if user == v.Owner {
			return k
		}
		//outputs = append(outputs[:k], outputs[k+1:]...)
	}

	return -1

}

// This function sums up the outputs that have the same owner and currency (label)
// And returns a new output list where each user have only one output per currency
func regroup_outputs(outputs Outputs) (Outputs, error) {
	var ret Outputs

	for _, v := range outputs {
		fmt.Println(ret)
		index := IndexOfUserInOutputs(v.Owner, ret)
		if index != -1 && ret[index].Label == v.Label {

			value0, err := v.Amount.Float64()
			if err != nil {
				return ret, fmt.Errorf("Parsing error : %s", err)
			}

			value1, err := ret[index].Amount.Float64()
			if err != nil {
				return ret, fmt.Errorf("Parsing error : %s", err)
			}

			fl := Round((value0+value1)*1e8) / 1e8
			ret[index].Amount = json.Number(strconv.FormatFloat(fl, 'f', -1, 64))

		} else {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

// this function returns the unspent inputs for the user given in parameter
func UnspentTxForUser(stub shim.ChaincodeStubInterface, user string) (Inputs, error) {
	var unspents Inputs

	value, err := stub.GetState(user)
	if err != nil {
		return unspents, fmt.Errorf("Error %s", err)
	}
	if value == nil {
		return nil, nil
	}
	b := bytes.NewReader([]byte(value))
	err0 := json.NewDecoder(b).Decode(&unspents)

	if err0 != nil {
		return unspents, fmt.Errorf("Error %s", err0)
	}
	return unspents, nil
}

// this function returns the list of every unspent transaction owned by the users un an output list
// minus the user identified by the string spender
func UnspentTxForUsersInOuputs(stub shim.ChaincodeStubInterface, outputs Outputs, spender string) (Inputs, error) {
	var ret Inputs
	for _, otp := range outputs {
		if otp.Owner != spender {
			more, failure := UnspentTxForUser(stub, otp.Owner)
			if failure != nil {
				return ret, fmt.Errorf("Error %s", failure)
			}
			ret = append(ret, more...)
		}
	}
	return ret, nil
}

// this function add the amount of the last unspent transaction for a user to
// the output list for each user, it returns a updated output list,
// and the inputs that have been used (which will need to be deleted in order to avoid double spending)
func AddUnspentsToOutputs(stub shim.ChaincodeStubInterface, outputs Outputs, spender string) (Outputs, Inputs, error) {

	var unspent_value Output
	var inputs_to_add Inputs
	unspents, fail := UnspentTxForUsersInOuputs(stub, outputs, spender)
	if fail != nil {
		return outputs, inputs_to_add, fmt.Errorf("Error : %s", fail)
	}
	fmt.Println(unspents)

	for _, unspent := range unspents {

		fmt.Println("loop")
		retreiving_key := fmt.Sprintf("%s_%d", unspent.Txid, unspent.J)

		raw, err := stub.GetState(retreiving_key)
		if err != nil {

			fmt.Println("Error getting")
			return outputs, inputs_to_add, fmt.Errorf("Error : %s", err)
		}
		if raw == nil {
			return outputs, inputs_to_add, fmt.Errorf("Error : cant retreive transaction", err)
		}

		b := bytes.NewReader(raw)
		err0 := json.NewDecoder(b).Decode(&unspent_value)
		if err0 != nil {

			fmt.Println("Error parsing", err0)
			//return outputs, inputs_to_add, false
		}

		for key, otp := range outputs {
			if unspent_value.Owner == otp.Owner && unspent_value.Label == otp.Label {
				inputs_to_add = append(inputs_to_add, unspent)

				value0, err := unspent_value.Amount.Float64()
				if err != nil {
					return outputs, inputs_to_add, fmt.Errorf("Error : %s", err)
				}

				value1, err := otp.Amount.Float64()
				if err != nil {
					return outputs, inputs_to_add, fmt.Errorf("Error : %s", err)
				}

				fl := Round((value0+value1)*1e8) / 1e8
				outputs[key].Amount = json.Number(strconv.FormatFloat(fl, 'f', -1, 64))
			}
		}
	}

	fmt.Println(outputs)
	return outputs, inputs_to_add, nil
}

// this function create the new assets given in outputs
// it makes sure users does not have multiple unspents transactions.
// it also maintain a local copy of each unspent transaction for
// users concerned by the transaction (kl)
func set_outputs(stub shim.ChaincodeStubInterface, txid string, outputs Outputs, inputs Inputs, kl *([]UserUnspents), spender string) error {

	var i int
	var new_input Input
	var failure error
	i = 0

	fmt.Println("Before")
	fmt.Println(outputs)

	outputs, failure = regroup_outputs(outputs)
	if failure != nil {
		return fmt.Errorf("Errora : %s", failure)
	}

	fmt.Println("Between")
	fmt.Println(outputs)

	var inputs_to_add Inputs
	outputs, inputs_to_add, failure = AddUnspentsToOutputs(stub, outputs, spender)
	if failure != nil {
		return fmt.Errorf("Errorb : %s", failure)
	}

	failureOut, _ := check_outputs(outputs, outputs[0].Label)
	if failureOut != nil {
		return fmt.Errorf("Errorc : %s", failureOut)
	}

	fmt.Println(inputs_to_add)
	failure = delete_inputs(stub, inputs_to_add)
	if failure != nil {
		return fmt.Errorf("Errord : %s", failure)
	}
	failure = delete_old_keys(kl, inputs_to_add)
	if failure != nil {
		return fmt.Errorf("Errore : %s", failure)
	}

	fmt.Println("After")
	fmt.Println(outputs)

	for k, v := range outputs {
		fmt.Println("Handling now : key[%s] value[%s]\n", k, v)

		//Unspents := UnspentTxForUser()
		//outputs, inputs_to_add, success = AddUnspentsToOutputs(stub, Unspents, outputs)

		new_input.Txid = txid
		new_input.J = i

		aku_fail := update_keys_list(kl, new_input, v)
		if aku_fail != nil {
			return fmt.Errorf("Error : %s", aku_fail)
		}

		createdkey := fmt.Sprintf("%s_%d", txid, i)

		rejson, err := json.Marshal(v)

		if err != nil {
			return fmt.Errorf("Error parsing : %s", err)
		}
		fmt.Println("CREATED KEY : %s", createdkey)
		fmt.Println(string(rejson))
		failure = stub.PutState(createdkey, rejson)
		if failure != nil {
			return fmt.Errorf("Error : %s", failure)
		}
		i++
	}
	return nil
}

//This function mint the coins given in arg[1] (outputs)
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

// given an array of byte representing json format of an output,
// this function parse it and write it to adress
func decode_single_transaction(arg string, adress *Output) error {

	b := bytes.NewReader([]byte(arg))

	err := json.NewDecoder(b).Decode(adress)

	if err != nil {
		return fmt.Errorf("Error %s", err)
	}

	if adress == nil {
		return fmt.Errorf("nil pointer received")
	}

	fmt.Printf("Parsed at single transaction level : %+v", adress)
	return nil
}

// this function returns a boolean indicating if a given key has
// already been updated in the ledger
func is_version_0(stub shim.ChaincodeStubInterface, key string) bool {
	value, err := stub.GetHistoryForKey(key)

	if err != nil {
		fmt.Println("Failed to get asset: %s with error: %s", key, err)
		return false
	}
	if value == nil {
		fmt.Println("Asset not found: %s", key)
		return false
	}
	value.Next()
	return !value.HasNext()
}

// this function parse an input list and makes sure they all have the same owner and label,
// and have never been updated. It also returns the total amount of the inputs as a float64,
// and the label of the input's transaction
func check_inputs(stub shim.ChaincodeStubInterface, inputs Inputs, who string) (error, float64, string) {
	//TODO
	//[x] same owner
	//[x] same labels
	//[x] sum
	//[x] version is 0 for key
	//[ ] use the real composite keys
	//[ ] shady floating point imprecision

	var output Output
	var label string
	var total_amount float64

	fmt.Println("Now checking inputs")

	total_amount = 0.0
	for k, v := range inputs {
		fmt.Println("Handling now : key[%s] value[%s]\n", k, v)

		retreiving_key := fmt.Sprintf("%s_%d", v.Txid, v.J)

		//Checking if key is version 0 (mandatory)
		v0 := is_version_0(stub, retreiving_key)

		if !v0 {

			return fmt.Errorf("Error : the key has been updated at leat once"), total_amount, label
		}

		fmt.Println("Retreiving key is :%s", retreiving_key)
		tran, err := stub.GetState(retreiving_key)
		if err != nil {
			return fmt.Errorf("Error :%s", err), total_amount, label
		}

		if tran == nil {
			return fmt.Errorf("Error getting transaction"), total_amount, label
		}

		fmt.Println(string(tran))

		err = decode_single_transaction(string(tran), &output)
		if err != nil {
			return fmt.Errorf("Error : %s", err), total_amount, label
		}

		//Check always said owner
		if output.Owner != who {
			return fmt.Errorf("Error : owners mismatch"), total_amount, label

		}

		//Check always same label (currency)
		if label == "" {
			label = output.Label
		} else {
			if output.Label != label {
				return fmt.Errorf("Error : labels mismatch"), total_amount, label
			}
		}
		amount, err := output.Amount.Float64()
		if err != nil {
			return fmt.Errorf("Error : %s", err), total_amount, label
		}

		fmt.Println("amount =", amount)
		amount = Round(amount*1e8) / 1e8
		total_amount += amount
		total_amount = Round(total_amount*1e8) / 1e8
	}

	fmt.Println("Total amount of inputs : %f", total_amount)
	return nil, total_amount, label
}

// this function checks if the outputs all have the label given as parameter
// and returns a sum of them as float64
func check_outputs(outputs Outputs, label string) (error, float64) {
	//TODO
	//[x] same labels
	//[x] sum
	//[ ] shady floating point imprecision

	var output Output
	var total_amount float64

	fmt.Println("Now checking Outputs")

	var i int
	i = 0
	total_amount = 0.0
	for k, v := range outputs {
		fmt.Println("Handling now : key[%s] value[%s]\n", k, v)

		output = v

		//Check always same label (currency)
		if output.Label != label {
			return fmt.Errorf("Labels mismatch in outputs"), total_amount
		}
		amount, err := output.Amount.Float64()
		if err != nil {
			return fmt.Errorf("Error : %s", err), total_amount
		}

		fmt.Println("amount =", amount)
		amount = Round(amount*1e8) / 1e8
		total_amount += amount

		total_amount = Round(total_amount*1e8) / 1e8
		i++
	}

	fmt.Println("Total amount of outputs : %f", total_amount)
	return nil, total_amount
}

// given a local copy of users unspent transactions (kl), and an input list, this function
// delete these inputs in kl
func delete_old_keys(keylist *([]UserUnspents), inputs Inputs) error {

	for key, usr := range *keylist {
		var ownerUnspents Inputs
		if usr.Unspents != nil {
			b := bytes.NewReader(usr.Unspents)
			err := json.NewDecoder(b).Decode(&ownerUnspents)
			if err != nil {
				return fmt.Errorf("Error : %s", err)
			}
			//deleting old coin
			//ownerUnspents = append(ownerUnspents, input)
			for k, value := range ownerUnspents {
				for _, ip := range inputs {
					if value.Txid == ip.Txid && value.J == ip.J {

						fmt.Println("LEN : %d k = %d", len(ownerUnspents), k)
						if len(ownerUnspents) > k {
							ownerUnspents = append(ownerUnspents[:k], ownerUnspents[k+1:]...)
						} else {
							ownerUnspents = ownerUnspents[:k]
						}
					}
				}
			}
		}

		//reconverting his coin list in bytes
		rejson, err := json.Marshal(ownerUnspents)
		if err != nil {
			return fmt.Errorf("Error_reparse : %s", err)
		}
		//rewrting his coin list
		(*keylist)[key].Unspents = rejson
	}
	return nil
}

// given a local copy of users unspent transactions (kl), and a couple of form (input, output),
// this function update kl to add the new reference to unspent transaction for the user.
func update_keys_list(keylist *([]UserUnspents), input Input, output Output) error {

	fmt.Println("Updating\n\n")
	for key, usr := range *keylist {
		//if new coin concerns user
		if usr.User == output.Owner {

			var ownerUnspents Inputs
			if usr.Unspents != nil {
				//converting his coins in obj
				b := bytes.NewReader(usr.Unspents)
				err := json.NewDecoder(b).Decode(&ownerUnspents)
				if err != nil {
					return fmt.Errorf("Error : %s", err)
				}
			}
			//adding new coin
			ownerUnspents = append(ownerUnspents, input)

			//reconverting his coin list in bytes
			rejson, err := json.Marshal(ownerUnspents)
			if err != nil {
				return fmt.Errorf("Error : %s", err)
			}
			//rewrting his coin list
			(*keylist)[key].Unspents = rejson
		}
	}

	return nil
}

func indexOfString(list []string, st string) int {

	for key, value := range list {
		if value == st {
			return (key)
		}
	}
	return (-1)
}

// this function update the list of users owning coins in the ledger
func updateUserList(stub shim.ChaincodeStubInterface, keylist []UserUnspents) error {

	//generating user list as array of strings
	fmt.Println("USERLIST UPDATE")
	var usersList []string
	var usersListInLedger []string
	for _, usr := range keylist {
		usersList = append(usersList, usr.User)
	}

	fmt.Println("USERLIST : ", usersList)

	value, err := stub.GetState("UserList")
	if err != nil {

		fmt.Println("la ")
		return fmt.Errorf("Error : %s", err)
	}

	if value != nil {
		b := bytes.NewReader(value)
		err = json.NewDecoder(b).Decode(&usersListInLedger)
		if err != nil {
			return fmt.Errorf("Error : %s", err)
		}

	}

	fmt.Println("USERLIST UPDATE VALUE = ", string(value))

	for _, usr := range usersList {
		if indexOfString(usersListInLedger, usr) == -1 {
			usersListInLedger = append(usersListInLedger, usr)
		}
	}

	rejson, err0 := json.Marshal(usersListInLedger)
	if err0 != nil {
		return fmt.Errorf("Error : %s", err0)
	}

	err = stub.PutState("UserList", rejson)
	if err != nil {
		return fmt.Errorf("Error : %s", err)
	}
	return nil
}

// this function update the ledger with the local copy of users unspent transactions (kl)
func commit_updated_keys(stub shim.ChaincodeStubInterface, keylist []UserUnspents) error {

	for key, usr := range keylist {
		fmt.Println("key = %d, user : %s", key, usr.User)

		var coinlist Inputs
		b := bytes.NewReader(usr.Unspents)
		err := json.NewDecoder(b).Decode(&coinlist)
		if err != nil {
			fmt.Errorf("Err commiting : %s", err)
		}
		for _, coin := range coinlist {
			fmt.Println("txid = %s, id : %d", coin.Txid, coin.J)
		}

		err = stub.PutState(usr.User, usr.Unspents)
		if err != nil {
			fmt.Errorf("Errc : %s", err)
		}
	}

	err := updateUserList(stub, keylist)
	if err != nil {
		return fmt.Errorf("Errc : %s", err)
	}

	return nil
}

//given a single UserUnspents, this function tells you if it is
// already or not in the given UserUnspents List
func user_already_in_list(new_user UserUnspents, UsersToUpdate []UserUnspents) bool {

	for _, user := range UsersToUpdate {

		if new_user.User == user.User {
			return true
		}
	}
	return false

}

// this function generate a list of UserUnspents containing all the users
// mentionned in outputs or inputs
func get_keys_for_owners(stub shim.ChaincodeStubInterface, outputs Outputs, inputs Inputs) ([]UserUnspents, error) {
	//TODO
	//[ ] tidy, restructure flow (two putstate...)

	var err error
	//var rejson []byte
	var UsersToUpdate []UserUnspents
	var current_user UserUnspents

	for _, output := range outputs {

		//fmt.Println("Handling now : key[%s] value[%s]\n", kkk, output)
		current_user.User = output.Owner
		if !user_already_in_list(current_user, UsersToUpdate) {
			current_user.Unspents, err = stub.GetState(output.Owner)
			if err != nil {
				return UsersToUpdate, fmt.Errorf("Errgk : %s", err)
			}

			UsersToUpdate = append(UsersToUpdate, current_user)
		}
	}

	if inputs != nil {
		for _, input := range inputs {
			var output Output
			var tran []byte
			retreiving_key := fmt.Sprintf("%s_%d", input.Txid, input.J)
			fmt.Println("Retreiving key is :%s", retreiving_key)
			tran, err = stub.GetState(retreiving_key)
			if err != nil {
				return UsersToUpdate, fmt.Errorf("Errgk : %s", err)
			}
			if tran == nil {
				return UsersToUpdate, fmt.Errorf("Err : failed to retreive transaction")
			}

			fmt.Println(string(tran))

			ret := decode_single_transaction(string(tran), &output)
			if ret != nil {
				return UsersToUpdate, fmt.Errorf("Errgk : %s", ret)
			}
			current_user.User = output.Owner
			if !user_already_in_list(current_user, UsersToUpdate) {
				current_user.Unspents, err = stub.GetState(output.Owner)
				UsersToUpdate = append(UsersToUpdate, current_user)
			}
		}
	}

	fmt.Println("\n\nTo update ::: \n\n")
	for _, val := range UsersToUpdate {
		fmt.Println("To update : %s", val.User)
	}

	for key, usr := range UsersToUpdate {

		fmt.Println("key = %d, user : %s", key, usr.User)

		var coinlist Inputs

		if usr.Unspents != nil {
			b := bytes.NewReader(usr.Unspents)
			err := json.NewDecoder(b).Decode(&coinlist)

			if err != nil {
				return UsersToUpdate, fmt.Errorf("Errgki : %s", err)
			}
			for _, coin := range coinlist {
				fmt.Println("txid = %s, id : %d", coin.Txid, coin.J)
			}
		}
	}

	return UsersToUpdate, nil
}

func pem_encode_pubkey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return string(pemEncodedPub)
}

func getPemPublicKeyOfCreator(stub shim.ChaincodeStubInterface) (string, error) {

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return "", fmt.Errorf("Error : %s", err)
	}
	ecPublicKey := cert.PublicKey.(*ecdsa.PublicKey)
	//fmt.Println(ecPublicKey)
	//fmt.Printf("PUB : %x\n", ecdPublicKey)
	return pem_encode_pubkey(ecPublicKey), nil
}

func trimPemPubKey(key string) string {
	key = strings.Replace(key, "\n", "", -1)
	key = strings.Replace(key, "-----BEGIN PUBLIC KEY-----", "", -1)
	key = strings.Replace(key, "-----END PUBLIC KEY-----", "", -1)
	return key
}

// This function takes inputs in arg[0] and outputs in arg[1],
// and performs the transaction the UTXO way
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

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
