package main

import "strings"
import "fmt"
import "strconv"

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	usage() {
	fmt.Println("Usage:\t\t./wallet")
	fmt.Println("Balance:\t./wallet balance")
	fmt.Println("Spend:\t\t./wallet spend [Amount] [Owner] [Label]")
}

func	parseStdout(stdout string) string {
	var index	int
	
	index = strings.Index(stdout, "payload:")
	stdout = stdout[index + len("payload:\""):]
	stdout = strings.Split(stdout, "\n")[0]
	stdout = stdout[:(len(stdout) - 2)]
	stdout = strings.Replace(stdout, "\\", "", -1)

	return stdout
}

func	checkFund(tx []Transaction, balance []Balance, argv []string) (Transaction, Balance, error) {
	var ret	Transaction
	var val Balance
	var dec	float64
	var err	error

	dec, _= strconv.ParseFloat(argv[1], 64)
	err = fmt.Errorf("")

	for index, value := range balance {
		if value.Label == argv[3] && dec <= value.Amount {
			ret = tx[index]
			val = value
			err = nil
		}
	}
	
	return ret, val, err 
}

