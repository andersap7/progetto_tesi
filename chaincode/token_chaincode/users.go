package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type User struct {
	Name       string `json:"name"`
	Id         string `json:"id"`
	Role       string `json:"role"`
	Balance    int    `json:"balance"`
	Authorized bool   `json:"authorized"`
}

type UserInfo struct {
	Balance int    `json:"balance"`
	Role    string `json:"role"`
}

func (sc *SmartContract) GetClientId(ctx contractapi.TransactionContextInterface) (string, error) {
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", err
	}
	return id, nil
}

func (sc *SmartContract) Register(ctx contractapi.TransactionContextInterface, name string) (string, error) {

	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", err
	}

	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", errors.New("already exists")
	}

	user := User{
		Name:       name,
		Id:         id,
		Balance:    0,
		Role:       "unauthorized_user",
		Authorized: false,
	}

	u, _ := json.Marshal(user)

	ctx.GetStub().PutState(id, u)

	return fmt.Sprintf("user %s registered", id), nil

}

func (sc *SmartContract) Authorize(ctx contractapi.TransactionContextInterface, id string, role string) error {
	existing, err := ctx.GetStub().GetState(id)
	if err != nil {
		return err
	}

	if existing == nil {
		return fmt.Errorf("user %s not registered", id)
	}

	user := new(User)

	err = json.Unmarshal(existing, user)
	if err != nil {
		return err
	}

	mspid, err := ctx.GetClientIdentity().GetMSPID()

	if err != nil {
		return err
	}

	if mspid == "Org2MSP" {
		user.Authorized = true
		user.Role = role

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(id, userBytes)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("client can't authorize users")
}

func GetUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("user does not exist")
	}
	user := new(User)

	err = json.Unmarshal(existing, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (sc *SmartContract) GetUserInfo(ctx contractapi.TransactionContextInterface, id string) (*UserInfo, error) {
	existing, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("user does not exist")
	}
	user := new(User)
	err = json.Unmarshal(existing, user)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{Balance: user.Balance, Role: user.Role}

	return &userInfo, nil
}
