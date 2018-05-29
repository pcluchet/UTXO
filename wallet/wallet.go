package main

import "fmt"
import "os"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func address(publicKey string) {
	fmt.Printf("Public Key = %s 🗝\n", publicKey)
}

func balance(publicKey string, argv []string) {
	var tx []Transaction
	var balance []Balance
	var err error

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

func spend(publicKey string, argv []string) {
	var tx []Transaction
	var balance []Balance
	var err error

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

func main() {
	var argv []string
	var publicKey string

	argv = os.Args
	if publicKey = getPublicKey(); publicKey == "" {
		fmt.Fprintf(os.Stderr, "Network is down\n")
		os.Exit(1)
	}

	switch len(argv) {
	case 2:
		// Need to optimise this
		switch argv[1] {
		case "balance":
			balance(publicKey, argv[1:])
		case "address":
			address(publicKey)
		default:
			parseArgv(argv[1], "default")
		}
	case 5:
		parseSpend(publicKey, argv[1:])
		spend(publicKey, argv[1:])
	default:
		usage()
	}
}
