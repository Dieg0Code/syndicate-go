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

// TestSimpleMemory_Clear verifica que Clear vacía la memoria.
func TestSimpleMemory_Clear(t *testing.T) {
	mem := NewSimpleMemory()
	mem.Add(Message{Role: RoleUser, Content: "Mensaje"})
	mem.Clear()
	messages := mem.Get()
	if len(messages) != 0 {
		t.Errorf("Se esperaban 0 mensajes tras Clear, se obtuvieron %d", len(messages))
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
