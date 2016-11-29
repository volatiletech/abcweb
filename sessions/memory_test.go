package sessions

import (
	"testing"
	"time"
)

func TestMemoryStorerNew(t *testing.T) {
	t.Parallel()

	m, err := NewMemoryStorer(2, 2)
	if err != nil {
		t.Error(err)
	}

	if m.maxAge != 2 {
		t.Error("expected max age to be 2")
	}
}

func TestMemoryStorerNewDefault(t *testing.T) {
	t.Parallel()

	m, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Error(err)
	}

	if m.maxAge != time.Hour*24*7 {
		t.Error("expected max age to be a week")
	}
}

func TestMemoryStorerGet(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	val, err := m.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	m.Put("hi", "hello")

	val, err = m.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "hello" {
		t.Errorf("Expected %q, got %s", "hello", val)
	}
}

func TestMemoryStorerPut(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	if len(m.sessions) != 0 {
		t.Errorf("Expected len 0, got %d", len(m.sessions))
	}

	m.Put("hi", "hello")
	m.Put("hi", "whatsup")
	m.Put("yo", "friend")

	if len(m.sessions) != 2 {
		t.Errorf("Expected len 2, got %d", len(m.sessions))
	}

	val, err := m.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "whatsup" {
		t.Errorf("Expected %q, got %s", "whatsup", val)
	}

	val, err = m.Get("yo")
	if err != nil {
		t.Error(err)
	}
	if val != "friend" {
		t.Errorf("Expected %q, got %s", "friend", val)
	}
}

func TestMemoryStorerDel(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	if len(m.sessions) != 0 {
		t.Errorf("Expected len 0, got %d", len(m.sessions))
	}

	m.Put("hi", "hello")
	m.Put("hi", "whatsup")
	m.Put("yo", "friend")

	if len(m.sessions) != 2 {
		t.Errorf("Expected len 2, got %d", len(m.sessions))
	}

	err := m.Del("hi")
	if err != nil {
		t.Error(err)
	}

	_, err = m.Get("hi")
	if err == nil {
		t.Errorf("Expected get hi to fail")
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}
}

func TestMemoryStorerCleaner(t *testing.T) {
	wait := make(chan struct{})

	memorySleepFunc = func(time.Duration) {
		<-wait
	}

	m, err := NewMemoryStorer(time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	m.sessions["testid1"] = memorySession{
		value:   "test1",
		expires: time.Now().Add(time.Hour),
	}
	m.sessions["testid2"] = memorySession{
		value:   "test2",
		expires: time.Now().AddDate(0, 0, -1),
	}

	m.mut.RLock()
	if len(m.sessions) != 2 {
		t.Error("expected len 2")
	}
	m.mut.RUnlock()

	// stop sleep in cleaner loop
	wait <- struct{}{}
	wait <- struct{}{}

	m.mut.RLock()
	if len(m.sessions) != 1 {
		t.Errorf("expected len 1, got %d", len(m.sessions))
	}

	_, ok := m.sessions["testid2"]
	if ok {
		t.Error("expected testid2 to be deleted, but was not")
	}
	m.mut.RUnlock()
}
