package main

import "fmt"
import "encoding/pem"
import "crypto/x509"
import "strings"
import "crypto/ecdsa"
import "github.com/hyperledger/fabric/core/chaincode/shim"
import "github.com/hyperledger/fabric/core/chaincode/lib/cid"

func pem_encode_pubkey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return string(pemEncodedPub)
}

func getPemPublicKeyOfCreator(stub shim.ChaincodeStubInterface) (string, error) {

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return "", fmt.Errorf("Error : %s", err)
	}
	ecPublicKey := cert.PublicKey.(*ecdsa.PublicKey)
	//fmt.Println(ecPublicKey)
	//fmt.Printf("PUB : %x\n", ecdPublicKey)
	return pem_encode_pubkey(ecPublicKey), nil
}

func trimPemPubKey(key string) string {
	key = strings.Replace(key, "\n", "", -1)
	key = strings.Replace(key, "-----BEGIN PUBLIC KEY-----", "", -1)
	key = strings.Replace(key, "-----END PUBLIC KEY-----", "", -1)
	return key
}


