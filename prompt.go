package gokamy

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Section represents a section with its content and subsections.
type Section struct {
	Name        string     // Name of the section
	Lines       []string   // Lines of text in the section
	SubSections []*Section // Nested subsections within the section
	listCounter int        // Counter for numbering list items
}

// addLine adds a line of text to the section.
func (s *Section) addLine(line string) {
	s.Lines = append(s.Lines, line)
}

// addListItem adds a numbered list item to the section.
func (s *Section) addListItem(item string) {
	s.listCounter++                                    // Increment the list counter
	line := fmt.Sprintf("%d. %s", s.listCounter, item) // Format the list item
	s.addLine(line)                                    // Add the item as a line
}

// findSubSection recursively searches for a subsection by name within the current section.
func (s *Section) findSubSection(name string) *Section {
	// If the current section matches, return it
	if s.Name == name {
		return s
	}
	// Recursively search within subsections
	for _, sub := range s.SubSections {
		if found := sub.findSubSection(name); found != nil {
			return found
		}
	}
	return nil // Return nil if not found
}

// PromptBuilder is used to construct the prompt with sections and nested subsections.
type PromptBuilder struct {
	sections   []*Section          // Sections at the top level
	sectionMap map[string]*Section // A map for quick access to sections by name
}

// NewPromptBuilder creates a new instance of the prompt builder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		sections:   make([]*Section, 0),       // Initialize an empty slice of sections
		sectionMap: make(map[string]*Section), // Initialize an empty map for sections
	}
}

// findSection searches for a section (including subsections) by name.
func (pb *PromptBuilder) findSection(name string) *Section {
	for _, sec := range pb.sections {
		// Recursively search for the section in subsections
		if found := sec.findSubSection(name); found != nil {
			return found
		}
	}
	return nil // Return nil if the section is not found
}

// CreateSection creates a new section with the given name.
func (pb *PromptBuilder) CreateSection(name string) *PromptBuilder {
	// Check if the section already exists
	if _, exists := pb.sectionMap[name]; !exists {
		// Create a new section and add it to the builder
		section := &Section{Name: name}
		pb.sections = append(pb.sections, section)
		pb.sectionMap[name] = section
	}
	return pb
}

// AddSubSection adds a child subsection under the specified parent section.
func (pb *PromptBuilder) AddSubSection(childName, parentName string) *PromptBuilder {
	// Find the parent section
	parent := pb.findSection(parentName)
	if parent == nil {
		// If the parent doesn't exist, create it
		parent = &Section{Name: parentName}
		pb.sections = append(pb.sections, parent)
		pb.sectionMap[parentName] = parent
	}
	// Check if the child section already exists
	if parent.findSubSection(childName) == nil {
		// Create and add the child subsection
		child := &Section{Name: childName}
		parent.SubSections = append(parent.SubSections, child)
	}
	return pb
}

// AddText adds a line of text to any section or subsection.
func (pb *PromptBuilder) AddText(sectionName, text string) *PromptBuilder {
	// Find the section and add the text to it
	if section := pb.findSection(sectionName); section != nil {
		section.addLine(strings.TrimSpace(text)) // Trim any unnecessary spaces
	}
	return pb
}

// AddTextF is a helper method that accepts any value, converts it to a string (using JSON marshaling if needed),
// and then adds it as text to the specified section.
func (pb *PromptBuilder) AddTextF(sectionName string, value interface{}) *PromptBuilder {
	var text string

	// Check if the value is already a string.
	if str, ok := value.(string); ok {
		text = str
	} else {
		// Try to marshal the value to JSON.
		bytes, err := json.Marshal(value)
		if err != nil {
			// Fallback: use fmt.Sprintf to convert the value.
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddText(sectionName, text)
}

// AddListItem adds a numbered item to a list within any section or subsection.
func (pb *PromptBuilder) AddListItem(sectionName, item string) *PromptBuilder {
	// Find the section and add the list item
	if section := pb.findSection(sectionName); section != nil {
		section.addListItem(strings.TrimSpace(item)) // Trim any unnecessary spaces
	}
	return pb
}

// AddListItemF is a helper method that accepts any value, converts it to a string (using JSON marshaling if needed),
// and then adds it as a list item to the specified section.
func (pb *PromptBuilder) AddListItemF(sectionName string, value interface{}) *PromptBuilder {
	var text string

	// Check if the value is already a string.
	if str, ok := value.(string); ok {
		text = str
	} else {
		// Try to marshal the value to JSON.
		bytes, err := json.Marshal(value)
		if err != nil {
			// Fallback: use fmt.Sprintf to convert the value.
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddListItem(sectionName, text)
}

// buildSection generates the final string representation of a section and its subsections.
func buildSection(sec *Section, indent string) string {
	var sb strings.Builder

	// Write the opening tag for the section
	sb.WriteString(fmt.Sprintf("%s<%s>\n", indent, sec.Name))

	// Add the lines of the section, trimming unnecessary spaces
	if len(sec.Lines) > 0 {
		for _, line := range sec.Lines {
			sb.WriteString(indent)
			sb.WriteString(strings.TrimSpace(line)) // Trim unnecessary spaces
			sb.WriteString("\n")
		}
	}

	// Recursively add subsections
	for _, sub := range sec.SubSections {
		sb.WriteString(buildSection(sub, indent+"  "))
	}

	// Write the closing tag for the section
	sb.WriteString(fmt.Sprintf("%s</%s>\n", indent, sec.Name))
	return sb.String()
}

// Build generates the final prompt by concatenating all sections and subsections.
func (pb *PromptBuilder) Build() string {
	var sb strings.Builder
	for _, sec := range pb.sections {
		// Recursively build the section
		sb.WriteString(buildSection(sec, ""))
	}

	// Trim any unnecessary spaces or line breaks in the final output
	return strings.TrimSpace(sb.String())
}
