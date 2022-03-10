package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// estrae il file tar.gz contenente il modello in target
func Untar(r io.Reader, target string) error {
	gzr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}

		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

// hash di tutti i file presenti in una directory
func hashDir(filepath string, h hash.Hash) error {
	files, err := os.ReadDir(filepath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = hashDir(path.Join(filepath, file.Name()), h)
			if err != nil {
				return err
			}
		} else {
			fileBytes, err := os.ReadFile(path.Join(filepath, file.Name()))
			if err != nil {
				return err
			}
			h.Write(fileBytes)
		}
	}
	return nil
}

// converte stringa nella forma n,n,n,n in slice int64
func stringToInt(s string) []int64 {
	stringArray := strings.Split(s, ",")

	var result []int64

	for i := range stringArray {
		t, _ := strconv.ParseInt(stringArray[i], 10, 64)
		result = append(result, t)
	}
	return result
}

// funzione che legge dal world state, viene eseguita prima di ogni transazione
// assume che la chiave sia il primo argomento della funzione
func GetWorldState(ctx CustomTransactionContextInterface) error {
	_, params := ctx.GetStub().GetFunctionAndParameters()

	if len(params) > 0 {

		existing, err := ctx.GetStub().GetState(params[0])

		if err != nil {
			return errors.New("unable to interact with world state")
		}

		ctx.SetData(existing)

		return nil
	}
	return nil
}

func UnknownTransactionHandler(ctx CustomTransactionContextInterface) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	return fmt.Errorf("invalid function %s passed with args %v", fcn, args)
}
