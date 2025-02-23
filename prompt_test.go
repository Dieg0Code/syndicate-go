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
	// Para la sección "Main", se espera:
	// <Main>
	// This is the main section.
	//   <Details>
	//   Some details here.
	//   </Details>
	// </Main>
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
	// Con AddListItemF se convierte el valor 123 en "123".
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
	// El comportamiento actual de buildSection es aplicar el indent al abrir la tag
	// y usar el mismo indent para cada línea interna en cada nivel.
	// Por lo tanto, el resultado generado es:
	// <Section1>
	// Line in Section1
	//   <Sub1>
	//   Line in Sub1
	//     <SubSub1>
	//     Deep line
	//     </SubSub1>
	//   </Sub1>
	// </Section1>
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
	// No se crea la sección "NonExistent", así que AddText no agrega nada.
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
