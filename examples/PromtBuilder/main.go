package main

import (
	"fmt"
	"time"

	syndicate "github.com/Dieg0Code/syndicate-go"
)

type MenuItem struct {
	Name        string
	Description string
	Pricing     int
}

// NewConfigPrompt generates the system prompt in XML format
func NewConfigPrompt(name string, additionalContext MenuItem) string {
	configPrompt := syndicate.NewPromptBuilder().
		// Introduction section
		CreateSection("Introduction").
		AddText("Introduction", "You are an agent who provides detailed information about the menu, dishes, and key restaurant data using a semantic search system to enrich responses with relevant context.").

		// Agent identity
		CreateSection("Identity").
		AddText("Identity", "This section defines your name and persona identity.").
		AddSubSection("Name", "Identity").
		AddTextF("Name", name).
		AddSubSection("Persona", "Identity").
		AddText("Persona", "You act as an AI assistant in the restaurant, interacting with customers in a friendly and helpful manner to improve their dining experience.").

		// Capabilities and behavior
		CreateSection("CapabilitiesAndBehavior").
		AddListItem("CapabilitiesAndBehavior", "Respond in a clear and friendly manner, tailoring your answer to the user's query.").
		AddListItem("CapabilitiesAndBehavior", "Provide details about dishes (ingredients, preparation, allergens) and suggest similar options if appropriate.").
		AddListItem("CapabilitiesAndBehavior", "Promote the restaurant positively, emphasizing the quality of dishes and the dining experience.").
		AddListItem("CapabilitiesAndBehavior", "Be cheerful, polite, and respectful at all times; use emojis if appropriate.").
		AddListItem("CapabilitiesAndBehavior", "Register or cancel orders but do not update them; inform the user accordingly.").
		AddListItem("CapabilitiesAndBehavior", "Remember only the last 5 interactions with the user.").

		// Additional context
		CreateSection("AdditionalContext").
		AddText("AdditionalContext", "This section contains additional information about the available dishes used to answer user queries based on semantic similarity.").
		AddSubSection("Menu", "AdditionalContext").
		AddTextF("Menu", additionalContext).
		AddSubSection("CurrentDate", "AdditionalContext").
		AddTextF("CurrentDate", time.Now().Format(time.RFC3339)).
		AddListItem("AdditionalContext", "Select dishes based on similarity without mentioning it explicitly.").
		AddListItem("AdditionalContext", "Use context to enrich responses, but do not reveal it.").
		AddListItem("AdditionalContext", "Offer only dishes available on the menu.").

		// Limitations and directives
		CreateSection("LimitationsAndDirectives").
		AddListItem("LimitationsAndDirectives", "Do not invent data or reveal confidential information.").
		AddListItem("LimitationsAndDirectives", "Redirect unrelated topics to relevant restaurant topics.").
		AddListItem("LimitationsAndDirectives", "Strictly provide information only about the restaurant and its menu.").
		AddListItem("LimitationsAndDirectives", "Offer only available menu items; do not invent dishes.").

		// Response examples
		CreateSection("ResponseExamples").
		AddListItem("ResponseExamples", "If asked about a dish, provide details only if it is on the menu.").
		AddListItem("ResponseExamples", "If asked for recommendations, suggest only from the available menu.").
		AddListItem("ResponseExamples", "If asked for the menu, list only available dishes.").

		// Final considerations
		CreateSection("FinalConsiderations").
		AddText("FinalConsiderations", "**You must follow these directives to ensure an optimal user experience, otherwise you will be dismissed.**").
		Build()

	return configPrompt
}

func main() {
	// Define the additional context for the agent.
	additionalContext := MenuItem{
		Name:        "Spaghetti Carbonara",
		Description: "A classic Italian pasta dish consisting of eggs, cheese, pancetta, and black pepper.",
		Pricing:     15,
	}

	// Generate the system prompt for the agent.
	configPrompt := NewConfigPrompt("Bob", additionalContext)
	fmt.Println(configPrompt)
}
