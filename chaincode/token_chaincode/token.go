package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const totalSupplyKey = "totalSupply"
const allowancePrefix = "allowance"
const pricesKey = "prices"
const msp = "Org2MSP"

type event struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
}

type Prices struct {
	Upload int `json:"upload"`
	Use    int `json:"use"`
}

func (sc *SmartContract) SetPrices(ctx contractapi.TransactionContextInterface, upload int, use int) error {
	MSPID, err := ctx.GetClientIdentity().GetMSPID()

	if err != nil {
		return err
	}

	if MSPID != msp {
		return fmt.Errorf("client is not authorized to set prices")
	}

	prices := Prices{Upload: upload, Use: use}

	pricesBytes, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(pricesKey, pricesBytes)
}

func (sc *SmartContract) GetPrices(ctx contractapi.TransactionContextInterface) (*Prices, error) {
	return getPrices(ctx)
}

func (sc *SmartContract) Mint(ctx contractapi.TransactionContextInterface, amount int) error {
	MSPID, err := ctx.GetClientIdentity().GetMSPID()

	if err != nil {
		return err
	}

	if MSPID != msp {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	minter, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return err
	}

	log.Printf("minter id: %s", minter)

	minterUser, err := ctx.GetStub().GetState(minter)

	if err != nil {
		return err
	}
	var currentBalance int

	adminIDBytes, _ := ctx.GetStub().GetState("admin")
	adminID := string(adminIDBytes)
	if adminIDBytes == nil || adminID != minter {
		ctx.GetStub().PutState("admin", []byte(minter))
	}

	var user User
	err = json.Unmarshal(minterUser, &user)
	if err != nil {
		return err
	}
	currentBalance = user.Balance

	log.Printf("current balance: %d", currentBalance)

	updatedBalance := currentBalance + amount

	user.Balance = updatedBalance

	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(minter, userBytes)

	if err != nil {
		return err
	}

	// update total supply
	err = updateTotalSupply(ctx, amount)
	if err != nil {
		return err
	}

	// emit the transfer event
	transferEvent := event{"0x0", minter, amount}

	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return err
	}

	log.Printf("minter account %s balance updated from %d to %d", minter, currentBalance, updatedBalance)
	return nil
}

func updateTotalSupply(ctx contractapi.TransactionContextInterface, amount int) error {

	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)

	if err != nil {
		return err
	}

	var totalSupply int

	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes))
	}

	totalSupply += amount
	log.Printf("total supply: %d", totalSupply)
	return ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
}

func (sc *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, recipient string, amount int) error {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	err = transfer(ctx, clientID, recipient, amount)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SmartContract) GetBalance(ctx contractapi.TransactionContextInterface) (int, error) {
	id, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return -1, err
	}

	userBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return -1, err
	}

	user := new(User)

	err = json.Unmarshal(userBytes, user)

	if err != nil {
		return -1, err
	}
	return user.Balance, nil
}

func (sc *SmartContract) GetUserBalance(ctx contractapi.TransactionContextInterface, id string) (int, error) {
	userBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return -1, err
	}
	if userBytes == nil {
		return -1, errors.New("user not found")
	}
	user := new(User)

	err = json.Unmarshal(userBytes, user)

	if err != nil {
		return -1, err
	}
	return user.Balance, nil
}

func (sc *SmartContract) PayUpload(ctx contractapi.TransactionContextInterface) error {
	prices, err := getPrices(ctx)
	if err != nil {
		return err
	}

	return payAdmin(ctx, prices.Upload)
}

func (sc *SmartContract) PayAdmin(ctx contractapi.TransactionContextInterface) error {
	prices, err := getPrices(ctx)
	if err != nil {
		return err
	}

	return payAdmin(ctx, prices.Use)
}

func (sc *SmartContract) PayForModel(ctx contractapi.TransactionContextInterface, to string, model string) error {
	prices, err := getPrices(ctx)
	if err != nil {
		return err
	}

	from, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	fromUserBytes, err := ctx.GetStub().GetState(from)
	if err != nil {
		return err
	}

	fromUser := new(User)
	err = json.Unmarshal(fromUserBytes, fromUser)
	if err != nil {
		return err
	}

	adminID, err := getAdminID(ctx)
	if err != nil {
		return err
	}
	adminBytes, err := ctx.GetStub().GetState(adminID)
	if err != nil {
		return err
	}
	admin := new(User)
	err = json.Unmarshal(adminBytes, admin)
	if err != nil {
		return err
	}

	if from != to {
		if fromUser.Balance < prices.Use*2 {
			return fmt.Errorf("client account %s has insufficient funds", from)
		}

		toUserBytes, err := ctx.GetStub().GetState(to)
		if err != nil {
			return err
		}
		toUser := new(User)
		err = json.Unmarshal(toUserBytes, toUser)
		if err != nil {
			return err
		}
		toUser.Balance += prices.Use
		fromUser.Balance -= prices.Use
		updatedToUserBytes, _ := json.Marshal(toUser)

		ctx.GetStub().PutState(to, updatedToUserBytes)
	} else {
		if fromUser.Balance < prices.Use {
			return fmt.Errorf("client account %s has insufficient funds", from)
		}

	}
	fromUser.Balance -= prices.Use
	admin.Balance += prices.Use
	updatedFromUserBytes, _ := json.Marshal(fromUser)
	updatedAdmin, _ := json.Marshal(admin)
	ctx.GetStub().PutState(from, updatedFromUserBytes)
	ctx.GetStub().PutState(adminID, updatedAdmin)

	log.Printf("client %s paid %d to admin and %s to use model %s", from, prices.Use, to, model)
	return nil
}

func transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount int) error {
	if from == to {
		return errors.New("cannot transfer from and to same client")
	}

	if amount < 0 {
		return errors.New("transfer amount can't be negative")
	}

	fromUserBytes, err := ctx.GetStub().GetState(from)

	if err != nil {
		return err
	}

	var fromUser User
	err = json.Unmarshal(fromUserBytes, &fromUser)
	if err != nil {
		return err
	}

	if !fromUser.Authorized {
		return fmt.Errorf("client account %s is unauthorized", from)
	}

	if fromUser.Balance < amount {
		return fmt.Errorf("client account %s has insufficient funds", from)
	}

	toUserBytes, err := ctx.GetStub().GetState(to)

	if err != nil {
		return fmt.Errorf("recipient %s not found", to)
	}

	var toUser User
	err = json.Unmarshal(toUserBytes, &toUser)
	if err != nil {
		return errors.New("error unmarshaling recipient")
	}

	if !toUser.Authorized {
		return fmt.Errorf("recipient account %s is unauthorized", to)
	}

	fromUser.Balance -= amount
	toUser.Balance += amount

	updatedFromUserBytes, err := json.Marshal(fromUser)
	if err != nil {
		return errors.New("error marshaling sender")
	}

	updatedToUserBytes, err := json.Marshal(toUser)
	if err != nil {
		return errors.New("error marshaling recipient")
	}

	err = ctx.GetStub().PutState(from, updatedFromUserBytes)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(to, updatedToUserBytes)
	if err != nil {
		return err
	}

	log.Printf("client %s balance updated to %d", from, fromUser.Balance)
	log.Printf("recipient %s balance updated to %d", to, toUser.Balance)

	// emit the transfer event
	transferEvent := event{from, to, amount}

	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return errors.New("error marshaling event")
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)

	if err != nil {
		return errors.New("error setting event")
	}
	return nil
}

func getPrices(ctx contractapi.TransactionContextInterface) (*Prices, error) {
	pricesBytes, err := ctx.GetStub().GetState(pricesKey)
	if err != nil {
		return nil, err
	}

	prices := new(Prices)
	err = json.Unmarshal(pricesBytes, prices)

	if err != nil {
		return nil, err
	}

	return prices, nil
}

func payAdmin(ctx contractapi.TransactionContextInterface, amount int) error {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	adminID, err := getAdminID(ctx)
	if err != nil {
		return err
	}
	return transfer(ctx, clientID, adminID, amount)
}

func getAdminID(ctx contractapi.TransactionContextInterface) (string, error) {
	admin, err := ctx.GetStub().GetState("admin")
	if err != nil {
		return "", err
	}
	adminID := string(admin)
	return adminID, nil
}

func (sc *SmartContract) Burn(ctx contractapi.TransactionContextInterface, amount int) error {
	mspid, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if mspid != msp {
		return fmt.Errorf("client is not authorized to burn tokens")
	}

	minter, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}
	if amount <= 0 {
		return errors.New("burn amount must be a positive integer")
	}
	adminBytes, err := ctx.GetStub().GetState(minter)
	if err != nil {
		return fmt.Errorf("failed to read minter account: %v", err)
	}
	admin := new(User)
	err = json.Unmarshal(adminBytes, admin)
	if err != nil {
		return fmt.Errorf("error unmarshaling admin user: %v", err)
	}

	admin.Balance -= amount
	updatedAdminBytes, err := json.Marshal(admin)
	if err != nil {
		return fmt.Errorf("error marshaling admin user: %v", err)
	}
	err = ctx.GetStub().PutState(minter, updatedAdminBytes)
	if err != nil {
		return err
	}

	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("error reading total supply: %v", err)
	}
	if totalSupplyBytes == nil {
		return errors.New("total supply does not exist")
	}
	totalSupply, _ := strconv.Atoi(string(totalSupplyBytes))

	totalSupply -= amount
	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return err
	}

	e := event{minter, "0x0", amount}
	eventJSON, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("error marshaling event: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", eventJSON)
	if err != nil {
		return err
	}
	return nil
}

func (sc *SmartContract) TotalSupply(ctx contractapi.TransactionContextInterface) (int, error) {
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve total supply: %v", err)
	}

	var totalSupply int
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes))
	}
	log.Printf("TotalSupply: %d tokens", totalSupply)

	return totalSupply, nil
}

// previsto dallo standard ERC-20, consente allo spender di prelevare token dall'account del chiamante
func (sc *SmartContract) Approve(ctx contractapi.TransactionContextInterface, spender string, amount int) error {
	owner, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client's id: %v", err)
	}
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})

	if err != nil {
		return fmt.Errorf("failed to update state for key %s: %v", allowanceKey, err)
	}

	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(amount)))
	if err != nil {
		return err
	}

	event := event{owner, spender, amount}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling: %v", err)
	}
	err = ctx.GetStub().SetEvent("Approval", eventJSON)
	if err != nil {
		return fmt.Errorf("error setting event: %v", err)
	}
	log.Printf("%s approved a withdrawal allowance of %d from %s", owner, amount, spender)
	return nil
}

func (sc *SmartContract) Allowance(ctx contractapi.TransactionContextInterface, owner string, spender string) (int, error) {
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
	if err != nil {
		return 0, fmt.Errorf("error creating composite key: %v", err)
	}

	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return 0, fmt.Errorf("error reading world state for key %s: %v", allowanceKey, err)
	}
	var allowance int
	if allowanceBytes == nil {
		allowance = 0
	} else {
		allowance, _ = strconv.Atoi(string(allowanceBytes))
	}
	log.Printf("allowance left for spender %s from owner %s: %d", spender, owner, allowance)
	return allowance, nil
}

func (sc *SmartContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, amount int) error {
	spender, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error getting client identity: %v", err)
	}
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{from, spender})
	if err != nil {
		return fmt.Errorf("error creating composite key: %v", err)
	}
	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return fmt.Errorf("error reading world state for key %s: %v", allowanceKey, err)
	}

	allowance, _ := strconv.Atoi(string(allowanceBytes))

	if allowance < amount {
		return fmt.Errorf("not enough allowance for transfer")
	}

	err = transfer(ctx, from, to, amount)
	if err != nil {
		return fmt.Errorf("failed transfer: %v", err)
	}
	updatedAllowance := allowance - amount
	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(updatedAllowance)))

	if err != nil {
		return err
	}

	log.Printf("spender %s allowance updated from %d to %d", spender, allowance, updatedAllowance)
	return nil
}
