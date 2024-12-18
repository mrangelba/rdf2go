package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	rdf "github.com/mrangelba/rdf2go"
)

func main() {
	file, err := os.Open("example/example.ttl")
	if err != nil {
		log.Panicf("failed reading file: %s", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)

	if err != nil {
		println(err.Error())
	}

	var person Person

	err = rdf.Unmarshal(data, &person)

	if err != nil {
		println(err.Error())
	}

	// Serializa e imprime o JSON resultante
	outputJSON, err := json.Marshal(person)

	if err != nil {
		println(err.Error())
	}

	fmt.Println(string(outputJSON))
}

type Person struct {
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Email       string    `json:"email,omitempty"`
	Description string    `json:"description,omitempty"`
	Knows       []Know    `json:"knows,omitempty"`
	Address     []Address `json:"address,omitempty"`
}

type Address struct {
	AddressCountry  string `json:"addressCountry,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	AddressRegion   string `json:"addressRegion,omitempty"`
	Key             string `json:"key,omitempty"`
}

type Know struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}
