package email

import (
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	wp := NewWorkerPool(2)

	// Тест занимания и освобождения воркеров
	wp.Acquire()
	wp.Acquire()

	// Должен блокироваться пока не освободится воркер
	go func() {
		wp.Acquire()
		t.Log("✅ Third worker acquired")
	}()

	time.Sleep(100 * time.Millisecond)
	wp.Release() // Освобождаем воркера
	time.Sleep(100 * time.Millisecond)
}
