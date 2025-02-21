package main

import (
	"encoding/json"
	"fmt"
	"log"

	gokamy "github.com/Dieg0Code/gokamy-ai"
)

type MenuItemSchema struct {
	ItemName string `json:"item_name" description:"Menu item name" required:"true"`
	Quantity int    `json:"quantity" description:"Quantity ordered by the user" required:"true"`
	Price    int    `json:"price" description:"Menu item price" required:"true"`
}

type UserOrderFunctionSchema struct {
	MenuItems       []MenuItemSchema `json:"menu_items" description:"List of ordered menu items" required:"true"`
	DeliveryAddress string           `json:"delivery_address" description:"Order delivery address" required:"true"`
	UserName        string           `json:"user_name" description:"User's name placing the order" required:"true"`
	PhoneNumber     string           `json:"phone_number" description:"User's phone number" required:"true"`
	PaymentMethod   string           `json:"payment_method" description:"Payment method (cash or transfer only)" required:"true" enum:"cash,transfer"`
}

type Product struct {
	ID        int     `json:"id" description:"Unique product identifier" required:"true"`
	Name      string  `json:"name" description:"Product name" required:"true"`
	Category  string  `json:"category" description:"Category of the product" enum:"Electronic,Furniture,Clothing"`
	Price     float64 `json:"price" description:"Price of the product"`
	Available bool    `json:"available" description:"Product availability" required:"true"`
}

func main() {
	schema, err := gokamy.GenerateRawSchema(Product{})
	if err != nil {
		log.Fatal(err)
	}

	schema2, err := gokamy.GenerateRawSchema(UserOrderFunctionSchema{})
	if err != nil {
		log.Fatal(err)
	}

	pretty, err := json.MarshalIndent(json.RawMessage(schema), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	pretty2, err := json.MarshalIndent(json.RawMessage(schema2), "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Product schema:")
	fmt.Println(string(pretty))
	fmt.Println("UserOrderFunction schema:")
	fmt.Println(string(pretty2))
}
