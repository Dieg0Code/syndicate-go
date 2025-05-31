package syndicate

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCreateAndBuildSimplePrompt(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Introduction")
	pb.AddText("Introduction", "  Welcome to the system!  ")
	result := pb.Build()
	expected := "<Introduction>\nWelcome to the system!\n</Introduction>"
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddSubSectionAndBuild(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Main")
	pb.AddText("Main", "This is the main section.")
	pb.AddSubSection("Details", "Main")
	pb.AddText("Details", "Some details here.")
	result := pb.Build()
	expected := "<Main>\nThis is the main section.\n  <Details>\n  Some details here.\n  </Details>\n</Main>"
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddTextFWithJSONConversion(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Data")
	data := map[string]int{"count": 10}
	pb.AddTextF("Data", data)
	result := pb.Build()
	jsonData, _ := json.Marshal(data)
	expected := fmt.Sprintf("<Data>\n%s\n</Data>", string(jsonData))
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddListItemAndListItemF(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("ListSection")
	pb.AddListItem("ListSection", "first item")
	pb.AddListItem("ListSection", "second item")
	pb.AddListItemF("ListSection", 123)
	result := pb.Build()
	expectedLines := []string{
		"<ListSection>",
		"1. first item",
		"2. second item",
		"3. 123",
		"</ListSection>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestMultipleNestedSubsections(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Section1")
	pb.AddText("Section1", "Line in Section1")
	pb.AddSubSection("Sub1", "Section1")
	pb.AddText("Sub1", "Line in Sub1")
	pb.AddSubSection("SubSub1", "Sub1")
	pb.AddText("SubSub1", "Deep line")
	result := pb.Build()
	expected := "<Section1>\nLine in Section1\n  <Sub1>\n  Line in Sub1\n    <SubSub1>\n    Deep line\n    </SubSub1>\n  </Sub1>\n</Section1>"
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestCreateDuplicateSection(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("DupSection")
	pb.CreateSection("DupSection")
	pb.AddText("DupSection", "Only once.")
	result := pb.Build()
	expected := "<DupSection>\nOnly once.\n</DupSection>"
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddTextToNonExistentSection(t *testing.T) {
	pb := NewPromptBuilder()
	pb.AddText("NonExistent", "Should not appear")
	result := pb.Build()
	if result != "" {
		t.Errorf("Expected empty prompt, got: %s", result)
	}
}

func TestBuildWithNoSections(t *testing.T) {
	pb := NewPromptBuilder()
	result := pb.Build()
	if result != "" {
		t.Errorf("Expected empty prompt for no sections, got: %s", result)
	}
}

// New tests for markdown features

func TestAddBulletItem(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("BulletList")
	pb.AddBulletItem("BulletList", "First bullet")
	pb.AddBulletItem("BulletList", "Second bullet")
	result := pb.Build()
	expectedLines := []string{
		"<BulletList>",
		"- First bullet",
		"- Second bullet",
		"</BulletList>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddBulletItemF(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("BulletList")
	pb.AddBulletItem("BulletList", "Text bullet")
	pb.AddBulletItemF("BulletList", 456)
	pb.AddBulletItemF("BulletList", map[string]string{"key": "value"})
	result := pb.Build()

	jsonData, _ := json.Marshal(map[string]string{"key": "value"})
	expectedLines := []string{
		"<BulletList>",
		"- Text bullet",
		"- 456",
		fmt.Sprintf("- %s", string(jsonData)),
		"</BulletList>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddCodeBlock(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("CodeSection")
	pb.AddCodeBlock("CodeSection", "func main() {\n\tfmt.Println(\"Hello\")\n}", "go")
	result := pb.Build()
	expectedLines := []string{
		"<CodeSection>",
		"```go",
		"func main() {",
		"\tfmt.Println(\"Hello\")",
		"}",
		"```",
		"</CodeSection>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddCodeBlockNoLanguage(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("CodeSection")
	pb.AddCodeBlock("CodeSection", "plain text code", "")
	result := pb.Build()
	expectedLines := []string{
		"<CodeSection>",
		"```",
		"plain text code",
		"```",
		"</CodeSection>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddCodeBlockF(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("CodeSection")
	data := struct {
		Name string
		Age  int
	}{"John", 30}
	pb.AddCodeBlockF("CodeSection", data, "json")
	result := pb.Build()

	jsonData, _ := json.Marshal(data)
	expectedLines := []string{
		"<CodeSection>",
		"```json",
		string(jsonData),
		"```",
		"</CodeSection>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddBoldText(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Formatting")
	pb.AddBoldText("Formatting", "This is bold text")
	result := pb.Build()
	expectedLines := []string{
		"<Formatting>",
		"**This is bold text**",
		"</Formatting>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddItalicText(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Formatting")
	pb.AddItalicText("Formatting", "This is italic text")
	result := pb.Build()
	expectedLines := []string{
		"<Formatting>",
		"*This is italic text*",
		"</Formatting>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddHeader(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Headers")
	pb.AddHeader("Headers", "Level 1 Header", 1)
	pb.AddHeader("Headers", "Level 3 Header", 3)
	pb.AddHeader("Headers", "Too Large Level", 8) // Should limit to level 6
	pb.AddHeader("Headers", "Too Small Level", 0) // Should default to level 1
	result := pb.Build()
	expectedLines := []string{
		"<Headers>",
		"# Level 1 Header",
		"### Level 3 Header",
		"###### Too Large Level",
		"# Too Small Level",
		"</Headers>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddBlockquote(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Quotes")
	pb.AddBlockquote("Quotes", "This is a blockquote")
	pb.AddBlockquote("Quotes", "Multi-line\nblockquote\nwith three lines")
	result := pb.Build()
	expectedLines := []string{
		"<Quotes>",
		"> This is a blockquote",
		"> Multi-line",
		"> blockquote",
		"> with three lines",
		"</Quotes>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddLink(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Links")
	pb.AddLink("Links", "Go to Google", "https://www.google.com")
	result := pb.Build()
	expectedLines := []string{
		"<Links>",
		"[Go to Google](https://www.google.com)",
		"</Links>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddHorizontalRule(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Divider")
	pb.AddText("Divider", "Before divider")
	pb.AddHorizontalRule("Divider")
	pb.AddText("Divider", "After divider")
	result := pb.Build()
	expectedLines := []string{
		"<Divider>",
		"Before divider",
		"---",
		"After divider",
		"</Divider>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddTable(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Tables")
	headers := []string{"Name", "Age", "Role"}
	rows := [][]string{
		{"John", "30", "Developer"},
		{"Jane", "28", "Designer"},
		{"Bob", "35", "Manager"},
	}
	pb.AddTable("Tables", headers, rows)
	result := pb.Build()
	expectedLines := []string{
		"<Tables>",
		"| Name | Age | Role |",
		"| --- | --- | --- |",
		"| John | 30 | Developer |",
		"| Jane | 28 | Designer |",
		"| Bob | 35 | Manager |",
		"</Tables>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestAddTableWithFewerColumns(t *testing.T) {
	pb := NewPromptBuilder()
	pb.CreateSection("Tables")
	headers := []string{"Name", "Age", "Role"}
	rows := [][]string{
		{"John", "30"},                    // Missing Role
		{"Jane"},                          // Missing Age and Role
		{"Bob", "35", "Manager", "Extra"}, // Extra column
	}
	pb.AddTable("Tables", headers, rows)
	result := pb.Build()
	expectedLines := []string{
		"<Tables>",
		"| Name | Age | Role |",
		"| --- | --- | --- |",
		"| John | 30 |  |",
		"| Jane |  |  |",
		"| Bob | 35 | Manager |",
		"</Tables>",
	}
	expected := strings.Join(expectedLines, "\n")
	if result != expected {
		t.Errorf("Expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestComplexPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()

	// Create a comprehensive prompt with various markdown features
	pb.CreateSection("Instructions")
	pb.AddHeader("Instructions", "Task Instructions", 2)
	pb.AddText("Instructions", "Please follow these instructions carefully:")
	pb.AddBulletItem("Instructions", "Review the provided code")
	pb.AddBulletItem("Instructions", "Identify any bugs or issues")
	pb.AddBulletItem("Instructions", "Suggest improvements")

	pb.CreateSection("CodeExample")
	pb.AddText("CodeExample", "Here's the code to review:")
	pb.AddCodeBlock("CodeExample", "func sum(a, b int) int {\n    return a - b  // Bug: This should be a + b\n}", "go")

	pb.CreateSection("Guidelines")
	pb.AddBoldText("Guidelines", "Important points to consider:")
	pb.AddListItem("Guidelines", "Check for off-by-one errors")
	pb.AddListItem("Guidelines", "Ensure proper error handling")
	pb.AddBlockquote("Guidelines", "Remember that clean code is more important than clever code")

	pb.CreateSection("References")
	pb.AddLink("References", "Go Style Guide", "https://golang.org/doc/effective_go")
	pb.AddHorizontalRule("References")
	pb.AddTable("References",
		[]string{"Resource", "URL", "Notes"},
		[][]string{
			{"Go Documentation", "https://golang.org/doc", "Official docs"},
			{"Go by Example", "https://gobyexample.com", "Practical examples"},
		})

	result := pb.Build()

	// This test ensures that a complex prompt with all features compiles
	// without error and has the expected number of sections
	if !strings.Contains(result, "<Instructions>") ||
		!strings.Contains(result, "<CodeExample>") ||
		!strings.Contains(result, "<Guidelines>") ||
		!strings.Contains(result, "<References>") {
		t.Errorf("Complex prompt missing expected sections")
	}

	// Check for specific markdown elements
	if !strings.Contains(result, "## Task Instructions") ||
		!strings.Contains(result, "```go") ||
		!strings.Contains(result, "**Important points to consider:**") ||
		!strings.Contains(result, "> Remember that clean code") {
		t.Errorf("Complex prompt missing expected markdown elements")
	}
}
