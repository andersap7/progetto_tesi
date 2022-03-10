package main

import (
	"crypto/sha256"
	"errors"
	"log"

	"encoding/base64"
	"encoding/json"
	"fmt"

	tensorflow "github.com/galeone/tensorflow/tensorflow/go"
	shell "github.com/ipfs/go-ipfs-api"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type UserInfo struct {
	Balance int    `json:"balance"`
	Role    string `json:"role"`
}

type Prices struct {
	Upload int `json:"upload"`
	Use    int `json:"use"`
}

type ModelUse struct {
	Creator string `json:"creator"`
	Model   string `json:"model"`
	Hash    string `json:"hash"`
	Price   int    `json:"price"`
	User    string `json:"user"`
}

// salvataggio sul disco di un modello, inviato come hash di ipfs
func (sc *SmartContract) SaveModel(ctx CustomTransactionContextInterface, name string, cid string,
	inputName string, inputDT string, inputShape string, inputIdx int,
	outputName string, outputDT string, outputShape string, outputIdx int) error {

	userID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	response := ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("GetUserInfo"), []byte(userID)}, ctx.GetStub().GetChannelID())

	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)

	user := new(UserInfo)
	err = json.Unmarshal(response.Payload, user)
	if err != nil {
		return err
	}

	if user.Role != "dev" {
		return fmt.Errorf("not allowed to upload a model. role: %s", user.Role)
	}

	response = ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("GetPrices")}, ctx.GetStub().GetChannelID())
	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)

	prices := new(Prices)
	err = json.Unmarshal(response.Payload, prices)
	if err != nil {
		return err
	}

	if user.Balance < prices.Upload {
		return fmt.Errorf("insufficient funds. current balance: %d\nneeded: %d", user.Balance, prices.Upload)
	}

	existing := ctx.GetData()

	if existing != nil {
		return fmt.Errorf("cannot create world state pair with key %s. Already exists", name)
	}

	sh := shell.NewShell("ipfs_host:5001")
	file, err := sh.Cat(cid)
	if err != nil {
		return err
	}

	err = Untar((file), "models/"+cid)
	if err != nil {
		return fmt.Errorf("error extracting file %s", err)
	}

	input := Data{
		Name:     inputName,
		Shape:    stringToInt(inputShape),
		DataType: tfTypes[inputDT],
		Idx:      inputIdx,
	}
	output := Data{
		Name:     outputName,
		DataType: tfTypes[outputDT],
		Shape:    stringToInt(outputShape),
		Idx:      outputIdx,
	}

	response = ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("PayUpload")}, ctx.GetStub().GetChannelID())
	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)
	if response.Status == 500 {
		return fmt.Errorf("error during payment: %s", response.Message)
	}

	h := sha256.New()

	err = hashDir(MODELS_FOLDER+cid, h)
	if err != nil {
		return fmt.Errorf("error hashing directory %s", err)
	}
	hashString := fmt.Sprintf("%x", h.Sum(nil))

	model := Model{
		Id:           cid,
		Name:         name,
		Hash:         hashString,
		Location:     "models/" + cid,
		Input:        input,
		Output:       output,
		Creator:      userID,
		AllowedUsers: []string{userID},
	}

	m, _ := json.Marshal(model)

	devIndexKey, err := ctx.GetStub().CreateCompositeKey("byDev", []string{userID, name})
	if err != nil {
		return err
	}
	ctx.GetStub().PutState(name, m)
	ctx.GetStub().PutState(devIndexKey, m)

	return nil

}

func (sc *SmartContract) GetModel(ctx CustomTransactionContextInterface, name string) (*ModelResult, error) {
	existing := ctx.GetData()

	if existing == nil {
		return nil, fmt.Errorf("no model with key %s found", name)
	}

	model := new(Model)
	err := json.Unmarshal(existing, model)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling model %s", err)
	}

	result := ModelResult{
		Name:   model.Name,
		Input:  model.Input,
		Output: model.Output,
	}
	return &result, nil
}

func (sc *SmartContract) RunModel(ctx CustomTransactionContextInterface, name string) (string, error) {

	userID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", err
	}

	existing := ctx.GetData()

	if existing == nil {
		return "", fmt.Errorf("no model with key %s found", name)
	}

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", err
	}
	input, exists := transientMap["input"]
	if !exists {
		return "", errors.New("error getting input from transient map")
	}

	decodedInput, err := base64.StdEncoding.DecodeString(string(input))

	if err != nil {
		return "", fmt.Errorf("error decoding input")
	}

	model := new(Model)
	err = json.Unmarshal(existing, model)

	if err != nil {
		return "", fmt.Errorf("error unmarshaling %s", err)
	}

	h := sha256.New()
	err = hashDir(model.Location, h)
	if err != nil {
		return "", err
	}
	hashString := fmt.Sprintf("%x", h.Sum(nil))
	if hashString != model.Hash {
		return "", fmt.Errorf("hash doesn't match")
	}

	log.Printf("checking if %s is authorized to run model %s", userID, name)

	var predictions *tensorflow.Tensor

	response := ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("GetPrices")}, ctx.GetStub().GetChannelID())
	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)

	prices := new(Prices)
	err = json.Unmarshal(response.Payload, prices)
	if err != nil {
		return "", err
	}

	response = ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("GetUserInfo"), []byte(userID)}, ctx.GetStub().GetChannelID())
	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)
	user := new(UserInfo)
	err = json.Unmarshal(response.Payload, user)
	if err != nil {
		return "", err
	}

	if userID == model.Creator {
		if user.Balance < prices.Use {
			return "", fmt.Errorf("insufficient funds. balance: %d\nprice: %d", user.Balance, prices.Use)
		}

	} else {
		if !model.isAllowed(userID) {
			return "", errors.New("user not allowed to run model")
		}
		if user.Balance < prices.Use*2 {
			return "", fmt.Errorf("insufficient funds. balance: %d\nprice: %d", user.Balance, prices.Use)
		}
	}
	predictions, err = model.execute(decodedInput)

	if err != nil {
		return "", fmt.Errorf("error executing model: %s", err)
	}
	response = ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("PayForModel"), []byte(model.Creator), []byte(model.Name)}, ctx.GetStub().GetChannelID())

	log.Printf("status code: %d\n, payload: %s\n, message: %s", response.Status, string(response.Payload), response.Message)
	if response.Status == 500 {
		return "", fmt.Errorf("error during payment: %s", response.Message)
	}

	event := ModelUse{
		Creator: model.Creator,
		Model:   model.Name,
		Hash:    model.Hash,
		Price:   prices.Use,
		User:    userID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("error marshaling event: %v", err)
	}
	err = ctx.GetStub().SetEvent("ModelUse", eventJSON)
	if err != nil {
		return "", fmt.Errorf("error setting event: %v", err)
	}
	return fmt.Sprintf("%v", predictions.Value()), nil
}

func modelsFromIterator(iterator shim.StateQueryIteratorInterface) ([]*ModelResult, error) {
	var models []*ModelResult

	for iterator.HasNext() {
		m, err := iterator.Next()

		if err != nil {
			return nil, err
		}

		var model Model
		err = json.Unmarshal(m.Value, &model)
		if err != nil {
			return nil, err
		}
		result := ModelResult{
			Name:   model.Name,
			Input:  model.Input,
			Output: model.Output,
		}
		models = append(models, &result)

	}
	return models, nil
}

func (sc *SmartContract) GetAllModels(ctx CustomTransactionContextInterface) ([]*ModelResult, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, fmt.Errorf("error reading state: %s", err)
	}
	defer resultsIterator.Close()

	var models []*ModelResult

	for resultsIterator.HasNext() {
		m, err := resultsIterator.Next()

		if err != nil {
			return nil, fmt.Errorf("error reading iterator: %s", err)
		}

		var model Model
		err = json.Unmarshal(m.Value, &model)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling: %s", err)
		}
		r := ModelResult{
			Name:   model.Name,
			Input:  model.Input,
			Output: model.Output,
		}
		models = append(models, &r)
	}
	return models, nil
}

func (sc *SmartContract) GetModelsByDev(ctx CustomTransactionContextInterface, username string) ([]*ModelResult, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("byDev", []string{username})

	if err != nil {
		return nil, err
	}
	return modelsFromIterator(iterator)
}
func (sc *SmartContract) Authorize(ctx CustomTransactionContextInterface, modelID string, id string) error {

	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("no model %s exists", modelID)
	}

	result := ctx.GetStub().InvokeChaincode("tokens", [][]byte{[]byte("GetUserInfo"), []byte(id)}, ctx.GetStub().GetChannelID())
	if result.Status == 500 {
		return fmt.Errorf("error invoking chaincode: %s", result.Message)
	}
	userToAuthorize := new(UserInfo)
	err := json.Unmarshal(result.Payload, userToAuthorize)
	if err != nil {
		return fmt.Errorf("")
	}
	if userToAuthorize.Role == "unauthorized_user" {
		return fmt.Errorf("user %s not authorized by admin", id)
	}
	var model Model

	err = json.Unmarshal(existing, &model)

	if err != nil {
		return err
	}

	userId, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return err
	}

	if userId != model.Creator {
		return fmt.Errorf("you aren't the owner of the model")
	}

	err = model.authorize(id)

	if err != nil {
		return err
	}

	modelBytes, err := json.Marshal(model)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(modelID, modelBytes)
}
