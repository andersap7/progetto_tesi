package main

import (
	"bytes"
	"errors"

	tf "github.com/galeone/tensorflow/tensorflow/go"
	tg "github.com/galeone/tfgo"
)

const MODELS_FOLDER = "./models/"

type Model struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Hash         string   `json:"hash"`
	Location     string   `json:"location"`
	Input        Data     `json:"input"`
	Output       Data     `json:"output"`
	Creator      string   `json:"creator"`
	AllowedUsers []string `json:"allowed_users"`
}

type Data struct {
	Name     string      `json:"name"`
	DataType tf.DataType `json:"datatype"`
	Shape    []int64     `json:"shape"`
	Idx      int         `json:"idx"`
}

var tfTypes = map[string]tf.DataType{
	"float":      tf.Float,
	"int32":      tf.Int32,
	"double":     tf.Double,
	"int64":      tf.Int64,
	"uint32":     tf.Uint32,
	"uint64":     tf.Uint64,
	"bool":       tf.Bool,
	"complex":    tf.Complex,
	"complex64":  tf.Complex64,
	"complex128": tf.Complex128,
	"half":       tf.Half,
	"bfloat16":   tf.Bfloat16,
	"int8":       tf.Int8,
	"int16":      tf.Int16,
	"uint8":      tf.Uint8,
	"uint16":     tf.Uint16,
}

type ModelResult struct {
	Name   string `json:"name"`
	Input  Data   `json:"input"`
	Output Data   `json:"output"`
}

func (m *Model) execute(input []byte) (*tf.Tensor, error) {

	model := tg.LoadModel(m.Location, []string{"serve"}, nil)

	inputTensor, err := tf.ReadTensor(m.Input.DataType, m.Input.Shape, bytes.NewReader(input))

	if err != nil {
		return nil, errors.New("error creating input tensor")
	}

	results := model.Exec([]tf.Output{
		model.Op(m.Output.Name, m.Output.Idx),
	}, map[tf.Output]*tf.Tensor{
		model.Op(m.Input.Name, m.Input.Idx): inputTensor,
	})

	return results[0], nil
}

func (m *Model) authorize(userID string) error {

	for _, id := range m.AllowedUsers {
		if id == userID {
			return errors.New("user already authorized")
		}
	}
	m.AllowedUsers = append(m.AllowedUsers, userID)
	return nil
}

func (m *Model) isAllowed(userID string) bool {
	for _, id := range m.AllowedUsers {
		if userID == id {
			return true
		}
	}
	return false
}
