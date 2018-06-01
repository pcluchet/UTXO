package main

import "bytes"
import "encoding/json"
import "fmt"
import "math"
import "github.com/hyperledger/fabric/core/chaincode/shim"

const shift = 64 - 11 - 1
const bias  = 1023
const mask  = 0x7FF

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

