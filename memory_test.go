package syndicate

import (
	"sync"
	"testing"
)

// TestSimpleMemory_AddAndGet verifica que al agregar mensajes se recuperan correctamente.
func TestSimpleMemory_AddAndGet(t *testing.T) {
	mem := NewSimpleMemory()

	msg1 := Message{
		Role:    RoleUser,
		Content: "Mensaje 1",
	}

	msg2 := Message{
		Role:    RoleAssistant,
		Content: "Mensaje 2",
	}

	mem.Add(msg1)
	mem.Add(msg2)

	messages := mem.Get()
	if len(messages) != 2 {
		t.Fatalf("Se esperaban 2 mensajes, se obtuvieron %d", len(messages))
	}

	if messages[0].Content != "Mensaje 1" {
		t.Errorf("El primer mensaje debía ser 'Mensaje 1', se obtuvo '%s'", messages[0].Content)
	}
	if messages[1].Content != "Mensaje 2" {
		t.Errorf("El segundo mensaje debía ser 'Mensaje 2', se obtuvo '%s'", messages[1].Content)
	}
}

// TestSimpleMemory_GetReturnsCopy verifica que Get retorne una copia de los mensajes.
func TestSimpleMemory_GetReturnsCopy(t *testing.T) {
	mem := NewSimpleMemory()
	origMsg := Message{Role: RoleUser, Content: "Original"}
	mem.Add(origMsg)

	retrieved := mem.Get()
	if len(retrieved) != 1 {
		t.Fatalf("Se esperaba 1 mensaje, se obtuvieron %d", len(retrieved))
	}

	// Se modifica la copia y se verifica que el original no cambie.
	retrieved[0].Content = "Modificado"
	fresh := mem.Get()
	if fresh[0].Content != "Original" {
		t.Errorf("Get no devolvió una copia; se modificó el mensaje original a '%s'", fresh[0].Content)
	}
}

// TestSimpleMemory_Concurrency prueba que las operaciones de Add sean seguras en concurrencia.
func TestSimpleMemory_Concurrency(t *testing.T) {
	mem := NewSimpleMemory()
	const iterations = 1000
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mem.Add(Message{Role: RoleUser, Content: "Mensaje concurrente"})
		}(i)
	}
	wg.Wait()

	messages := mem.Get()
	if len(messages) != iterations {
		t.Errorf("Se esperaban %d mensajes tras operaciones concurrentes, se obtuvieron %d", iterations, len(messages))
	}
}

// TestNewMemory_WithFunctionalOptions verifica que NewMemory funcione con functional options.
func TestNewMemory_WithFunctionalOptions(t *testing.T) {
	var storedMessages []Message
	var mutex sync.Mutex

	mem, err := NewMemory(
		WithAddHandler(func(message Message) {
			mutex.Lock()
			defer mutex.Unlock()
			storedMessages = append(storedMessages, message)
		}),
		WithGetHandler(func() []Message {
			mutex.Lock()
			defer mutex.Unlock()
			// Retornar copia para simular comportamiento thread-safe
			copyMessages := make([]Message, len(storedMessages))
			copy(copyMessages, storedMessages)
			return copyMessages
		}),
	)

	if err != nil {
		t.Fatalf("Error inesperado al crear memoria: %v", err)
	}

	// Agregar mensajes
	msg1 := Message{Role: RoleUser, Content: "Test 1"}
	msg2 := Message{Role: RoleAssistant, Content: "Test 2"}

	mem.Add(msg1)
	mem.Add(msg2)

	// Verificar que se almacenaron correctamente
	messages := mem.Get()
	if len(messages) != 2 {
		t.Fatalf("Se esperaban 2 mensajes, se obtuvieron %d", len(messages))
	}

	if messages[0].Content != "Test 1" {
		t.Errorf("El primer mensaje debía ser 'Test 1', se obtuvo '%s'", messages[0].Content)
	}
	if messages[1].Content != "Test 2" {
		t.Errorf("El segundo mensaje debía ser 'Test 2', se obtuvo '%s'", messages[1].Content)
	}
}

// TestNewMemory_MissingAddHandler verifica que NewMemory falle sin WithAddHandler.
func TestNewMemory_MissingAddHandler(t *testing.T) {
	_, err := NewMemory(
		WithGetHandler(func() []Message {
			return []Message{}
		}),
	)

	if err == nil {
		t.Error("Se esperaba error al faltar WithAddHandler")
	}
	if err != nil && err.Error() != "WithAddHandler is required when creating custom memory" {
		t.Errorf("Error inesperado: %v", err)
	}
}

// TestNewMemory_MissingGetHandler verifica que NewMemory falle sin WithGetHandler.
func TestNewMemory_MissingGetHandler(t *testing.T) {
	_, err := NewMemory(
		WithAddHandler(func(message Message) {}),
	)

	if err == nil {
		t.Error("Se esperaba error al faltar WithGetHandler")
	}
	if err != nil && err.Error() != "WithGetHandler is required when creating custom memory" {
		t.Errorf("Error inesperado: %v", err)
	}
}

// TestNewMemory_NilAddHandler verifica que WithAddHandler falle con función nil.
func TestNewMemory_NilAddHandler(t *testing.T) {
	_, err := NewMemory(
		WithAddHandler(nil),
		WithGetHandler(func() []Message {
			return []Message{}
		}),
	)

	if err == nil {
		t.Error("Se esperaba error al pasar nil a WithAddHandler")
	}
	if err != nil && err.Error() != "failed to apply memory option: addFunc cannot be nil" {
		t.Errorf("Error inesperado: %v", err)
	}
}

// TestNewMemory_NilGetHandler verifica que WithGetHandler falle con función nil.
func TestNewMemory_NilGetHandler(t *testing.T) {
	_, err := NewMemory(
		WithAddHandler(func(message Message) {}),
		WithGetHandler(nil),
	)

	if err == nil {
		t.Error("Se esperaba error al pasar nil a WithGetHandler")
	}
	if err != nil && err.Error() != "failed to apply memory option: getFunc cannot be nil" {
		t.Errorf("Error inesperado: %v", err)
	}
}

// TestNewMemory_EmptyOptions verifica que NewMemory falle sin opciones.
func TestNewMemory_EmptyOptions(t *testing.T) {
	_, err := NewMemory()

	if err == nil {
		t.Error("Se esperaba error al crear memoria sin opciones")
	}
}

// TestNewMemory_CustomLogic verifica que se pueda implementar lógica personalizada.
func TestNewMemory_CustomLogic(t *testing.T) {
	// Simulamos una memoria que solo guarda los últimos 3 mensajes
	var messages []Message
	maxSize := 3

	mem, err := NewMemory(
		WithAddHandler(func(message Message) {
			messages = append(messages, message)
			if len(messages) > maxSize {
				messages = messages[1:] // Remover el primer mensaje
			}
		}),
		WithGetHandler(func() []Message {
			copyMessages := make([]Message, len(messages))
			copy(copyMessages, messages)
			return copyMessages
		}),
	)

	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	// Agregar más mensajes que el límite
	for i := 0; i < 5; i++ {
		mem.Add(Message{
			Role:    RoleUser,
			Content: "Mensaje " + string(rune('1'+i)),
		})
	}

	// Verificar que solo se mantuvieron los últimos 3
	retrieved := mem.Get()
	if len(retrieved) != maxSize {
		t.Errorf("Se esperaban %d mensajes, se obtuvieron %d", maxSize, len(retrieved))
	}

	// Verificar que son los mensajes correctos (3, 4, 5)
	expectedContents := []string{"Mensaje 3", "Mensaje 4", "Mensaje 5"}
	for i, msg := range retrieved {
		if msg.Content != expectedContents[i] {
			t.Errorf("Mensaje %d: se esperaba '%s', se obtuvo '%s'",
				i, expectedContents[i], msg.Content)
		}
	}
}

// TestNewMemory_Concurrency verifica que una memoria personalizada pueda ser thread-safe.
func TestNewMemory_Concurrency(t *testing.T) {
	var messages []Message
	var mutex sync.RWMutex

	mem, err := NewMemory(
		WithAddHandler(func(message Message) {
			mutex.Lock()
			defer mutex.Unlock()
			messages = append(messages, message)
		}),
		WithGetHandler(func() []Message {
			mutex.RLock()
			defer mutex.RUnlock()
			copyMessages := make([]Message, len(messages))
			copy(copyMessages, messages)
			return copyMessages
		}),
	)

	if err != nil {
		t.Fatalf("Error inesperado: %v", err)
	}

	const iterations = 100
	var wg sync.WaitGroup

	// Ejecutar operaciones Add concurrentes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mem.Add(Message{
				Role:    RoleUser,
				Content: "Mensaje concurrente",
			})
		}(i)
	}

	wg.Wait()

	// Verificar que todos los mensajes se agregaron
	finalMessages := mem.Get()
	if len(finalMessages) != iterations {
		t.Errorf("Se esperaban %d mensajes, se obtuvieron %d",
			iterations, len(finalMessages))
	}
}
