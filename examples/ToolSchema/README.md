# **GenerateRawSchema: JSON Schema Generator for Go Structs**

`GenerateRawSchema` is a tool that generates a **JSON Schema** from a Go `struct`. The schema is built based on the **tags** defined in the struct fields.  

### **Supported Tags**
- **`description`** → Describes the purpose of the field to help the LLM understand its role.  
- **`required`** → Marks the field as mandatory.  
- **`enum`** → Specifies a set of allowed values for the field.  

The **field name** is extracted from the `json:"name"` tag, and the **type** is inferred based on the Go data type.

---

## **Example 1: Basic Struct Conversion**  

Consider the following Go struct:  

```go
type Product struct {
	ID        int     `json:"id" description:"Unique product identifier" required:"true"`
	Name      string  `json:"name" description:"Product name" required:"true"`
	Category  string  `json:"category" description:"Category of the product" enum:"Electronic,Furniture,Clothing"`
	Price     float64 `json:"price" description:"Price of the product"`
	Available bool    `json:"available" description:"Product availability" required:"true"`
}
```

This struct generates the following **JSON Schema**:

```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "integer",
      "description": "Unique product identifier"
    },
    "name": {
      "type": "string",
      "description": "Product name"
    },
    "category": {
      "type": "string",
      "description": "Category of the product",
      "enum": ["Electronic", "Furniture", "Clothing"]
    },
    "price": {
      "type": "number",
      "description": "Price of the product"
    },
    "available": {
      "type": "boolean",
      "description": "Product availability"
    }
  },
  "required": ["id", "name", "category", "price", "available"],
  "additionalProperties": false
}
```

### **Key Features**
✅ **Automatic Type Inference** → The `type` field is determined based on the Go data type.  
✅ **Strict Schema Enforcement** → The field `"additionalProperties": false` ensures no extra fields can be added.  
✅ **Enum Support** → The `"category"` field includes predefined values.  

---

## **Example 2: Complex Nested Structs**  

Go structs can also include nested objects, such as an order system:

```go
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
```

This struct generates the following **JSON Schema**:

```json
{
  "type": "object",
  "properties": {
    "menu_items": {
      "type": "array",
      "description": "List of ordered menu items",
      "items": {
        "type": "object",
        "properties": {
          "item_name": {
            "type": "string",
            "description": "Menu item name"
          },
          "quantity": {
            "type": "integer",
            "description": "Quantity ordered by the user"
          },
          "price": {
            "type": "integer",
            "description": "Menu item price"
          }
        },
        "required": ["item_name", "quantity", "price"],
        "additionalProperties": false
      }
    },
    "delivery_address": {
      "type": "string",
      "description": "Order delivery address"
    },
    "user_name": {
      "type": "string",
      "description": "User's name placing the order"
    },
    "phone_number": {
      "type": "string",
      "description": "User's phone number"
    },
    "payment_method": {
      "type": "string",
      "description": "Payment method (cash or transfer only)",
      "enum": ["cash", "transfer"]
    }
  },
  "required": ["menu_items", "delivery_address", "user_name", "phone_number", "payment_method"],
  "additionalProperties": false
}
```

### **Understanding the Nested Structure**
- **`menu_items`** is an **array of objects**, each containing:  
  - `item_name` (string)  
  - `quantity` (integer)  
  - `price` (integer)  

- `"additionalProperties": false` is automatically applied to all objects to prevent extra fields.  

---

## **Limitations & Further Exploration**
🔹 **This tool supports a subset of JSON Schema features** and may not handle very complex schemas.  
🔹 For **advanced use cases**, consider using [`invopop/jsonschema`](https://github.com/invopop/jsonschema), a Go library for more powerful JSON Schema generation.  
🔹 For deeper understanding, refer to the [official JSON Schema documentation](https://json-schema.org/).