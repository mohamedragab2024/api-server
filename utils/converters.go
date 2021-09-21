package utils

import (
	"encoding/json"
	"io"
	"log"
)

func StructToMap(this interface{}) (newMap map[string]interface{}) {
	data, err := json.Marshal(this)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	err = json.Unmarshal(data, &newMap)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return
}

func MapToStruct(this map[string]interface{}) (Newobj interface{}) {
	data, err := json.Marshal(this)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Newobj)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return
}

func MapToJson(this map[string]interface{}) (jsonObj []byte) {
	jsonObj, err := json.Marshal(this)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return
}

func JsonBodyToMap(this io.ReadCloser) (Newobj map[string]interface{}) {
	err := json.NewDecoder(this).Decode(&Newobj)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return
}

func StructToJson(this interface{}) (jsonObj []byte) {
	jsonObj, err := json.Marshal(this)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return
}

func JsonToStruct(this []byte) interface{} {
	var result struct{}
	err := json.Unmarshal(this, &result)
	if err != nil {
		log.Fatal(err)
	}

	return result
}
