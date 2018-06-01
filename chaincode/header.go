package main

import "encoding/json"

type SimpleAsset struct {
}

type Input struct {
	Txid string
	J    int
}

type Output struct {
	Amount json.Number
	Owner  string
	Label  string
}

type UserUnspents struct {
	User     string
	Unspents []byte
}

type Outputs []Output
type Inputs []Input


