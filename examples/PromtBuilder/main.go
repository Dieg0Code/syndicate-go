package main

import (
	"fmt"
	"time"

	syndicate "github.com/Dieg0Code/syndicate"
)

// User represents a user profile that might come from a database.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func main() {
	// Example dynamic data
	currentUser := User{ID: 101, Name: "Alice", Role: "Administrator"}
	currentTime := time.Now()

	// Create a new prompt using both formatted and non-formatted functions
	prompt := syndicate.NewPromptBuilder().
		// Static "Rules" section
		CreateSection("Rules").
		AddText("Rules", "System-wide policies to follow:").
		AddListItem("Rules", "Respect user privacy").
		AddListItem("Rules", "Provide accurate responses").
		// Dynamic "UserProfile" section
		CreateSection("UserProfile").
		AddText("UserProfile", "User Details:").
		AddTextF("UserProfile", currentUser). // Formatted: converts struct to JSON
		// Mixed "SessionInfo" section
		CreateSection("SessionInfo").
		AddText("SessionInfo", "Session Details:").
		AddTextF("SessionInfo", currentTime.Format(time.RFC3339)).
		AddTextF("SessionInfo", map[string]interface{}{
			"timestamp": currentTime.Format(time.RFC3339),
			"active":    true,
		}).                                              // Formatted: converts map to JSON
		AddListItem("SessionInfo", "Live chat enabled"). // Non-formatted list item
		// Mixed "Examples" section
		CreateSection("Examples").
		AddText("Examples", "Example use cases:").
		AddListItemF("Examples", []string{"Generate reports", "Analyze metrics"}). // Formatted: converts slice to JSON
		AddListItem("Examples", "Assist with documentation").                      // Non-formatted list item
		// Build the final prompt
		Build()

	// Print the generated prompt
	fmt.Println(prompt)
}
