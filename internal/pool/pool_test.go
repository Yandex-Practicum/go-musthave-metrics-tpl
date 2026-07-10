package pool

import "testing"

// buffer — пример «тяжёлого» объекта с методом Reset.
type buffer struct {
	data []byte
	uses int
}

// Reset сбрасывает состояние буфера в пустое.
func (b *buffer) Reset() {
	b.data = b.data[:0]
	b.uses = 0
}

// TestPool_GetUsesFactory проверяет, что при пустом пуле Get создаёт объект
// через переданную в New функцию.
func TestPool_GetUsesFactory(t *testing.T) {
	created := 0
	p := New(func() *buffer {
		created++
		return &buffer{}
	})

	b := p.Get()
	if b == nil {
		t.Fatal("Get вернул nil")
	}
	if created != 1 {
		t.Fatalf("фабрика вызвана %d раз, ожидался 1", created)
	}
}

// TestPool_NilFactory проверяет, что при nil-конструкторе пул выдаёт
// нулевое значение типа (для *buffer это nil), а не паникует.
func TestPool_NilFactory(t *testing.T) {
	p := New[*buffer](nil)

	if got := p.Get(); got != nil {
		t.Fatalf("Get при nil-конструкторе вернул %v, ожидался nil", got)
	}
}

// TestPool_PutResets проверяет, что Put сбрасывает состояние объекта перед
// возвратом в пул.
func TestPool_PutResets(t *testing.T) {
	p := New(func() *buffer { return &buffer{} })

	b := p.Get()
	b.data = append(b.data, 'a', 'b', 'c')
	b.uses = 5

	p.Put(b)

	if len(b.data) != 0 || b.uses != 0 {
		t.Fatalf("Put не сбросил состояние: data=%v uses=%d", b.data, b.uses)
	}
}
