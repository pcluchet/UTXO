package main

import "fmt"
import "os"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func	balance(publicKey string, argv []string) {
	var tx			[]Transaction
	var balance		[]Balance
	var err			error
	
	if tx, err = queryTransaction(publicKey); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if balance, err = queryBalance(tx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if balance == nil {
		fmt.Printf("Balance account = 0\n")
	}
	for _, value := range balance {
		fmt.Printf("Balance account = [%s] [%f]\n", value.Label, value.Amount)
	}
}

func	spend(publicKey string, argv []string) {
	var tx		[]Transaction
	var balance	[]Balance
	var err		error

	if tx, err = queryTransaction(publicKey); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if balance, err = queryBalance(tx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	if err = makeTransaction(tx, balance, argv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	fmt.Println("Money successfully sent 🚀")
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	main() {
	var argv		[]string
	var publicKey	string

	argv = os.Args
	publicKey = getPublicKey()
	fmt.Println(publicKey, "\n")
	
	switch len(argv) {
		case 2:
			parseArgv(argv[1], "balance")
			balance(publicKey, argv[1:])
		case 5:
			parseSpend(publicKey, argv[1:])
			spend(publicKey, argv[1:])
		default:
			usage()
	}
}
