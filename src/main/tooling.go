package main

import (
	"encoding/json"
	"fmt"
)

func main(){

	data := "{\"Data\":\"Data1\",\"Value\":\"Value1\"}"
	c := ContactImplem{}
	parse([]byte(data),&c)
	fmt.Println(c)
}

type  Parsable interface {
	Parse(data []byte)
}

type ContactImplem struct{
	Data string
	Value string
}

func (c * ContactImplem)Parse(data []byte){
	json.Unmarshal(data,c)
}


func parse(data []byte,c Parsable){
	c.Parse(data)
}