package main

import (
	htlc "chaincode/htlc"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	hashedTimeLockContract := new(htlc.HashedTimeLockContract)

	htlcChaincode, err := contractapi.NewChaincode(hashedTimeLockContract)
	if err != nil {
		panic(err.Error())
	}

	err = htlcChaincode.Start()
	if err != nil {
		panic(err.Error())
	}
}
