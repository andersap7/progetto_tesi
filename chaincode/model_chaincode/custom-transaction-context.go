package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

type CustomTransactionContext struct {
	contractapi.TransactionContext
	data []byte
}

type CustomTransactionContextInterface interface {
	contractapi.TransactionContextInterface
	GetData() []byte
	SetData([]byte)
}

func (ctc *CustomTransactionContext) GetData() []byte {
	return ctc.data
}

func (ctc *CustomTransactionContext) SetData(data []byte) {
	ctc.data = data
}
