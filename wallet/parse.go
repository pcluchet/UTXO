package main

import "fmt"
import "strconv"
import "os"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func parseAmount(amount string) {
	var dec float64
	var err error

	if dec, err = strconv.ParseFloat(amount, 64); err != nil {
		fmt.Println("ParseError = Amount cannot be a string")
		os.Exit(1)
	}
	if dec <= 0 {
		fmt.Println("ParseError = Amount cannot be less or equal to 0")
		os.Exit(1)
	}
}

func parseOwner(owner string) {
	if owner == "" {
		fmt.Printf("ParseError = Owner cannot be equal to nil\n\n")
		usage()
		os.Exit(1)
	}
}

func parseLabel(label string) {
	if label == "" {
		fmt.Printf("ParseError = Label cannot be equal to nil\n\n")
		usage()
		os.Exit(1)
	}
}

func parsePublicKey(argv string, publicKey string) {
	if argv == publicKey {
		fmt.Println("NiCe TrY !😈 👮")
		os.Exit(666)
	}
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func parseArgv(argv string, transactionType string) {
	if argv != transactionType {
		usage()
		os.Exit(1)
	}
}

func parseSpend(publicKey string, argv []string) {
	parseArgv(argv[0], "spend")
	parsePublicKey(argv[2], publicKey)
	parseAmount(argv[1])
	parseOwner(argv[2])
	parseLabel(argv[3])
}
