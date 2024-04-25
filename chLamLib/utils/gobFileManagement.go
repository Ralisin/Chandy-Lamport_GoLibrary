package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
)

var MapReflectType = map[string]reflect.Type{}

// getTypeByName retrieve reflect.Type from string
func getTypeByName(typeName string) (reflect.Type, bool) {
	t, ok := MapReflectType[typeName]
	return t, ok
}

// RegisterType usage: utils.RegisterType((typeCast)(pointer))
func RegisterType(ptr interface{}) {
	t := reflect.TypeOf(ptr).Elem()
	MapReflectType[t.String()] = t

	gob.Register(reflect.New(t).Interface())
}

func SaveToFile(fileName string, data interface{}) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", fileName, err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	if err = file.Sync(); err != nil {
		return fmt.Errorf("error file.Sync(): %v", err)
	}

	return nil
}

func RetrieveFromFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("[RetrieveFromFile] os.Open: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Verifica se il file è vuoto
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file is empty")
	}

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("error decode data: %v", err)
	}

	return nil
}

func ConvertInterfaceToBytes(data interface{}) ([]byte, error) {
	var dataBuf bytes.Buffer
	encReqIntoBuf := gob.NewEncoder(&dataBuf)
	if err := encReqIntoBuf.Encode(data); err != nil {
		fmt.Println("Encode error:", err)
		return nil, err
	}

	return dataBuf.Bytes(), nil
}

func ConvertDataBytesToInterface(data []byte, typeStr string) (interface{}, error) {
	// Retrieve data type
	reflectType, ok := getTypeByName(typeStr)
	if !ok {
		return nil, fmt.Errorf("[ConvertDataBytesToInterface] error retriving data type: %s.\n"+
			"\tTo use this type you must register it via utils.RegisterType", typeStr)
	}

	// New empty value from reflect.Type
	emptyType := reflect.Zero(reflectType).Interface()

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.DecodeValue(reflect.ValueOf(&emptyType)); err != nil {
		return nil, fmt.Errorf("[ConvertDataBytesToInterface] decodeValue error: %v", err)
	}

	return emptyType, nil
}
