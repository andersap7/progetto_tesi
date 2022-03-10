package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}
func main() {
	sc := new(SmartContract)

	cc, err := contractapi.NewChaincode(sc)
	
	if err != nil {
		fmt.Printf("error creacting chaincode: %s", err.Error())
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("error starting chaincode: %s", err.Error())
	}
}