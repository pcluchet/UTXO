package main

import "encoding/json"

/* ************************************************************************** */
/*	STRUCTURES															      */
/* ************************************************************************** */

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

/* ************************************************************************** */
/*	TYPEDEF																	  */
/* ************************************************************************** */

type Inputs []Input
type Outputs []Output
