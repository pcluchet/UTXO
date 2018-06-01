/**
 * File              : utxo.go
 * Author            : jle-quel <jle-quel@student.42.us.org>
 * Date              : 01.06.2018
 * Last Modified Date: 01.06.2018
 * Last Modified By  : jle-quel <jle-quel@student.42.us.org>
 */
/*
 * UTXO implementation
 *
 * How to use :
 *
 * --- Minting ---
 *	first argument	: "mint"
 *	second argument	: a signature (not implemented yet, any string will do)
 *	third argument	: list of outputs in json, in the form (amount, owner, label)
 *		example		: [{"amount":42.42, "owner": "USER PUBLIC KEY", "label": "USD"}]
 *					  Label can be any string to identify the currency in the system
 *					  One's public key can obtained by using the wallet,
 					  or by deriving it from its private key or certificate
 *
 *
 * --- Spending ---
 *	first argument	: "spend"
 *	second argument	: list of inputs in json, in the form (txid, j)
 *		example		: [{"txid": "SOME TXID", "j": 0}]
 *	third argument	: list of outputs in json, in the form (amount, owner, label)
 *		example		: [{"amount":42.42, "owner": "USER PUBLIC KEY", "label": "USD"}]
 *
 *
 *
 * --- Getting unspents coins for said public key ---
 *	first argument	: "get"
 *	second argument	: Some user pulic key
 *					  It returns a list of the unspent coins for said owner, in json format
 *			example	: [{"txid": "SOME TXID", "j": 0}, {"txid": "SOME OTHER TXID", "j": 3}]
 *
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

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
// the output list for each user, it returns an updated output list,
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
