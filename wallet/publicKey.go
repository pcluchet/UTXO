package main

import "os"
import "os/exec"
import "strings"
import "io/ioutil"
import "fmt"

/* ************************************************************************** */
/*		PRIVATE																  */
/* ************************************************************************** */

func	createPublicKey(filepath string, filename string) string {
	var cmd		string
	var stdout	[]byte
	var ret		string
	var err		error

	cmd = fmt.Sprintf("openssl ec -in %s/keystore/%s -pubout", filepath, filename)
	if stdout, err = exec.Command("/bin/sh", "-c", cmd).Output(); err != nil {
		fmt.Println(err)
	}
	
	ret = strings.TrimLeft(string(stdout), "-----BEGIN PUBLIC KEY-----\n")
	ret = strings.TrimRight(ret, "-----END PUBLIC KEY-----\n")
	ret = strings.Replace(ret, "\n", "", -1)

	return ret
}

/* ************************************************************************** */
/*		PUBLIC																  */
/* ************************************************************************** */

func	getPublicKey() string {
	var filepath	string
	var DIR			[]os.FileInfo
	var publicKey	string
	var err			error

	filepath = os.Getenv("CORE_PEER_MSPCONFIGPATH")
	if DIR, err = ioutil.ReadDir(filepath + "/keystore"); err != nil {
		return "" 
	}

	for _, value := range DIR {
		publicKey = createPublicKey(filepath, value.Name())
	}

	return publicKey
}
