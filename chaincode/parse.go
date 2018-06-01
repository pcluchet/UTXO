package main

import "fmt"
import "bytes"
import "encoding/json"

/* ************************************************************************** */
/*	PUBLIC																	  */
/* ************************************************************************** */

func decode_io(arg string, address interface{}) error {
	var err	error

	b := bytes.NewReader([]byte(arg))
	if err = json.NewDecoder(b).Decode(address); err != nil {
		return err
	}
	/* If address is equal to nit, err should be != nil */
	//if address == nil {
	//	return fmt.Errorf("Nil address given")
	//}
	fmt.Printf("Parsed at io level : %+v", address)
	return nil
}

