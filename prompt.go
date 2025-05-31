package syndicate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Section represents a block of content that may include text lines and nested subsections.
// It is used by the PromptBuilder to structure and format the final prompt.
type Section struct {
	Name        string     // Name of the section.
	Lines       []string   // Lines of text contained in the section.
	SubSections []*Section // Nested subsections within this section.
	listCounter int        // Internal counter used to number list items.
}

// addLine appends a single line of text to the Section.
func (s *Section) addLine(line string) {
	s.Lines = append(s.Lines, line)
}

// addListItem formats and adds a numbered list item to the Section.
// It increments the internal listCounter and prefixes the item with its number.
func (s *Section) addListItem(item string) {
	s.listCounter++                                    // Increment the counter for list items.
	line := fmt.Sprintf("%d. %s", s.listCounter, item) // Format the list item with its number.
	s.addLine(line)                                    // Append the formatted list item as a new line.
}

// addBulletItem formats and adds a bullet point item to the Section.
func (s *Section) addBulletItem(item string) {
	line := fmt.Sprintf("- %s", item) // Format the item with a bullet point.
	s.addLine(line)                   // Append the formatted bullet item as a new line.
}

// addCodeBlock formats and adds a code block with optional language specification.
func (s *Section) addCodeBlock(code, language string) {
	if language != "" {
		s.addLine(fmt.Sprintf("```%s", language))
	} else {
		s.addLine("```")
	}
	s.addLine(code)
	s.addLine("```")
}

// addBoldText formats and adds bold text.
func (s *Section) addBoldText(text string) {
	s.addLine(fmt.Sprintf("**%s**", text))
}

// addItalicText formats and adds italic text.
func (s *Section) addItalicText(text string) {
	s.addLine(fmt.Sprintf("*%s*", text))
}

// addHeader formats and adds a header of specified level (1-6).
func (s *Section) addHeader(text string, level int) {
	if level < 1 {
		level = 1
	} else if level > 6 {
		level = 6
	}
	prefix := strings.Repeat("#", level)
	s.addLine(fmt.Sprintf("%s %s", prefix, text))
}

// addBlockquote formats and adds a blockquote.
func (s *Section) addBlockquote(text string) {
	// Split multi-line quotes and prefix each line
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		s.addLine(fmt.Sprintf("> %s", strings.TrimSpace(line)))
	}
}

// addLink formats and adds a hyperlink.
func (s *Section) addLink(text, url string) {
	s.addLine(fmt.Sprintf("[%s](%s)", text, url))
}

// addHorizontalRule adds a horizontal rule.
func (s *Section) addHorizontalRule() {
	s.addLine("---")
}

// addTable formats and adds a simple table with headers and rows.
func (s *Section) addTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Add header row
	headerLine := "| " + strings.Join(headers, " | ") + " |"
	s.addLine(headerLine)

	// Add separator row
	separators := make([]string, len(headers))
	for i := range separators {
		separators[i] = "---"
	}
	separatorLine := "| " + strings.Join(separators, " | ") + " |"
	s.addLine(separatorLine)

	// Add data rows
	for _, row := range rows {
		if len(row) > len(headers) {
			row = row[:len(headers)]
		} else if len(row) < len(headers) {
			// Pad with empty cells
			for i := len(row); i < len(headers); i++ {
				row = append(row, "")
			}
		}
		rowLine := "| " + strings.Join(row, " | ") + " |"
		s.addLine(rowLine)
	}
}

// findSubSection recursively searches for a subsection by its name within the current section.
// If a matching subsection is found, it is returned; otherwise, nil is returned.
func (s *Section) findSubSection(name string) *Section {
	// If the current section's name matches, return it.
	if s.Name == name {
		return s
	}
	// Recursively search through all nested subsections.
	for _, sub := range s.SubSections {
		if found := sub.findSubSection(name); found != nil {
			return found
		}
	}
	return nil // No matching subsection was found.
}

// PromptBuilder facilitates the construction of a prompt by organizing content into sections and subsections.
type PromptBuilder struct {
	sections   []*Section          // Top-level sections in the prompt.
	sectionMap map[string]*Section // Quick lookup map for sections by name.
}

// NewPromptBuilder creates and initializes a new PromptBuilder instance.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		sections:   make([]*Section, 0),       // Initialize an empty slice for sections.
		sectionMap: make(map[string]*Section), // Initialize an empty map for section lookup.
	}
}

// findSection searches for a section (or nested subsection) by name across all top-level sections.
// It returns the first section matching the provided name, or nil if not found.
func (pb *PromptBuilder) findSection(name string) *Section {
	for _, sec := range pb.sections {
		if found := sec.findSubSection(name); found != nil {
			return found
		}
	}
	return nil // Section not found.
}

// CreateSection adds a new top-level section with the given name to the prompt.
// If a section with the same name already exists, it is not created again.
func (pb *PromptBuilder) CreateSection(name string) *PromptBuilder {
	if _, exists := pb.sectionMap[name]; !exists {
		section := &Section{Name: name}            // Create a new section.
		pb.sections = append(pb.sections, section) // Append it to the list of sections.
		pb.sectionMap[name] = section              // Store it in the map for quick lookup.
	}
	return pb
}

// AddSubSection creates a new subsection (child) under the specified parent section.
// If the parent section does not exist, it is created as a top-level section.
func (pb *PromptBuilder) AddSubSection(childName, parentName string) *PromptBuilder {
	// Locate the parent section by name.
	parent := pb.findSection(parentName)
	if parent == nil {
		// If the parent does not exist, create it as a new top-level section.
		parent = &Section{Name: parentName}
		pb.sections = append(pb.sections, parent)
		pb.sectionMap[parentName] = parent
	}
	// If the child subsection doesn't already exist under the parent, add it.
	if parent.findSubSection(childName) == nil {
		child := &Section{Name: childName}
		parent.SubSections = append(parent.SubSections, child)
	}
	return pb
}

// AddText adds a line of text to the specified section or subsection.
// It trims any extra whitespace before appending.
func (pb *PromptBuilder) AddText(sectionName, text string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addLine(strings.TrimSpace(text))
	}
	return pb
}

// AddTextF is a helper method that converts any value to its string representation
// (using JSON marshaling if necessary) and adds it as a text line to the specified section.
func (pb *PromptBuilder) AddTextF(sectionName string, value interface{}) *PromptBuilder {
	var text string

	// Check if the value is already a string.
	if str, ok := value.(string); ok {
		text = str
	} else {
		// Attempt to marshal the value into JSON.
		bytes, err := json.Marshal(value)
		if err != nil {
			// Fallback: use fmt.Sprintf for conversion.
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddText(sectionName, text)
}

// AddListItem adds a numbered list item to the specified section or subsection.
// The item is trimmed for any unnecessary whitespace.
func (pb *PromptBuilder) AddListItem(sectionName, item string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addListItem(strings.TrimSpace(item))
	}
	return pb
}

// AddListItemF is a helper method that converts any value to its string representation
// (using JSON marshaling if necessary) and adds it as a numbered list item to the specified section.
func (pb *PromptBuilder) AddListItemF(sectionName string, value interface{}) *PromptBuilder {
	var text string

	// Check if the value is already a string.
	if str, ok := value.(string); ok {
		text = str
	} else {
		// Attempt JSON marshaling.
		bytes, err := json.Marshal(value)
		if err != nil {
			// Fallback: use fmt.Sprintf for conversion.
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddListItem(sectionName, text)
}

// AddBulletItem adds a bullet point item to the specified section or subsection.
func (pb *PromptBuilder) AddBulletItem(sectionName, item string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addBulletItem(strings.TrimSpace(item))
	}
	return pb
}

// AddBulletItemF is a helper method that converts any value to its string representation
// and adds it as a bullet point item to the specified section.
func (pb *PromptBuilder) AddBulletItemF(sectionName string, value interface{}) *PromptBuilder {
	var text string

	if str, ok := value.(string); ok {
		text = str
	} else {
		bytes, err := json.Marshal(value)
		if err != nil {
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddBulletItem(sectionName, text)
}

// AddCodeBlock adds a code block with optional language specification to the specified section.
func (pb *PromptBuilder) AddCodeBlock(sectionName, code, language string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addCodeBlock(code, language)
	}
	return pb
}

// AddCodeBlockF is a helper method that converts any value to its string representation
// and adds it as a code block to the specified section.
func (pb *PromptBuilder) AddCodeBlockF(sectionName string, value interface{}, language string) *PromptBuilder {
	var text string

	if str, ok := value.(string); ok {
		text = str
	} else {
		bytes, err := json.Marshal(value)
		if err != nil {
			text = fmt.Sprintf("%v", value)
		} else {
			text = string(bytes)
		}
	}

	return pb.AddCodeBlock(sectionName, text, language)
}

// AddBoldText adds bold text to the specified section.
func (pb *PromptBuilder) AddBoldText(sectionName, text string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addBoldText(strings.TrimSpace(text))
	}
	return pb
}

// AddItalicText adds italic text to the specified section.
func (pb *PromptBuilder) AddItalicText(sectionName, text string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addItalicText(strings.TrimSpace(text))
	}
	return pb
}

// AddHeader adds a header of specified level to the section.
func (pb *PromptBuilder) AddHeader(sectionName, text string, level int) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addHeader(strings.TrimSpace(text), level)
	}
	return pb
}

// AddBlockquote adds a blockquote to the specified section.
func (pb *PromptBuilder) AddBlockquote(sectionName, text string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addBlockquote(strings.TrimSpace(text))
	}
	return pb
}

// AddLink adds a hyperlink to the specified section.
func (pb *PromptBuilder) AddLink(sectionName, text, url string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addLink(strings.TrimSpace(text), url)
	}
	return pb
}

// AddHorizontalRule adds a horizontal rule to the specified section.
func (pb *PromptBuilder) AddHorizontalRule(sectionName string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addHorizontalRule()
	}
	return pb
}

// AddTable adds a table with headers and rows to the specified section.
func (pb *PromptBuilder) AddTable(sectionName string, headers []string, rows [][]string) *PromptBuilder {
	if section := pb.findSection(sectionName); section != nil {
		section.addTable(headers, rows)
	}
	return pb
}

// buildSection recursively generates the string representation of a Section and its nested subsections.
// The indent parameter is used to properly format nested content.
func buildSection(sec *Section, indent string) string {
	var sb strings.Builder

	// Write the opening tag for the section.
	sb.WriteString(fmt.Sprintf("%s<%s>\n", indent, sec.Name))

	// Append each line of the section, ensuring proper trimming and formatting.
	for _, line := range sec.Lines {
		sb.WriteString(indent)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Recursively build all nested subsections with increased indentation.
	for _, sub := range sec.SubSections {
		sb.WriteString(buildSection(sub, indent+"  "))
	}

	// Write the closing tag for the section.
	sb.WriteString(fmt.Sprintf("%s</%s>\n", indent, sec.Name))
	return sb.String()
}

// Build generates the final prompt by concatenating all top-level sections and their nested subsections.
// It returns the fully formatted prompt as a single string.
func (pb *PromptBuilder) Build() string {
	var sb strings.Builder
	for _, sec := range pb.sections {
		// Recursively build each section.
		sb.WriteString(buildSection(sec, ""))
	}
	// Trim any extra spaces or line breaks from the final output.
	return strings.TrimSpace(sb.String())
}
