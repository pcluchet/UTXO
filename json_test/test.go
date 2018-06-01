package main

import "fmt"
import "os"
import "encoding/json"

func	main() {
	var hash	map[string]string
	var bts		[]byte
	var str		string
	var err		error

	/*
********************************************************************************	
*** FIRST STEP *****************************************************************
********************************************************************************	
	*/
	hash = make(map[string]string)
	hash["a"] = "AHHH"
	hash["b"] = "BEEH"
	hash["c"] = "CEEH"
	bts, err = json.Marshal(hash)
	if err != nil {
		fmt.Println("error: cannot marshal.")
		os.Exit(1)
	}
	str = string(bts)
	fmt.Println(str)
	/*
********************************************************************************	
*** SECOND STEP ****************************************************************
********************************************************************************	
	*/
	json.Unmarshal([]byte(str), &hash)
	fmt.Println(hash)
	fmt.Println(hash["a"])
	fmt.Println(hash["b"])
	fmt.Println(hash["c"])
}
