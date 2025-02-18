package main

import (
	"encoding/json"
	"fmt"
	"log"

	gokamy "github.com/Dieg0Code/gokamy-ai"
)

type Product struct {
	ID        int     `json:"id" description:"Unique product identifier" required:"true"`
	Name      string  `json:"name" description:"Product name" required:"true"`
	Category  string  `json:"category" description:"Category of the product" enum:"Electronic,Furniture,Clothing"`
	Price     float64 `json:"price" description:"Price of the product"`
	Available bool    `json:"available" description:"Product availability" required:"true"`
}

func main() {
	schema, err := gokamy.GenerateSchema(Product{})
	if err != nil {
		log.Fatal(err)
	}
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
