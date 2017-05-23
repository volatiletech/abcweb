package abcsessions

import (
	"testing"
	"time"
)

func TestMemoryStorerNew(t *testing.T) {
	m, err := NewMemoryStorer(2, 2)
	if err != nil {
		t.Error(err)
	}

	if m.maxAge != 2 {
		t.Error("expected max age to be 2")
	}

	m.wg.Wait()
}

func TestMemoryStorerNewDefault(t *testing.T) {
	t.Parallel()

	m, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Error(err)
	}

	if m.maxAge != time.Hour*24*2 {
		t.Error("expected max age to be 2 days")
	}

	m.wg.Wait()
}

func TestMemoryStorerAll(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	list, err := m.All()
	if err != nil {
		t.Error("expected no error on empty list")
	}
	if len(list) > 0 {
		t.Error("Expected len 0")
	}

	m.Set("hi", "hello")
	m.Set("yo", "friend")

	list, err = m.All()
	if err != nil {
		t.Error(err)
	}
	if len(list) != 2 {
		t.Errorf("Expected len 2, got %d", len(list))
	}
	if (list[0] != "hi" && list[0] != "yo") || list[0] == list[1] {
		t.Errorf("Expected list[0] to be %q or %q, got %q", "yo", "hi", list[0])
	}
	if (list[1] != "yo" && list[1] != "hi") || list[1] == list[0] {
		t.Errorf("Expected list[1] to be %q or %q, got %q", "hi", "yo", list[1])
	}
}

func TestMemoryStorerGet(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	val, err := m.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	m.Set("hi", "hello")

	val, err = m.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "hello" {
		t.Errorf("Expected %q, got %s", "hello", val)
	}
}

func TestMemoryStorerSet(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()

	if len(m.sessions) != 0 {
		t.Errorf("Expected len 0, got %d", len(m.sessions))
	}

	m.Set("hi", "hello")
	m.Set("hi", "whatsup")
	m.Set("yo", "friend")

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

	m.Set("hi", "hello")
	m.Set("hi", "whatsup")
	m.Set("yo", "friend")

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

// memoryTestTimer is used in the timerTestHarness override so we can
// control sending signals to the sleep channel and trigger cleans manually
type memoryTestTimer struct{}

func (memoryTestTimer) Reset(time.Duration) bool {
	return true
}

func (memoryTestTimer) Stop() bool {
	return true
}

func TestMemoryStorerCleaner(t *testing.T) {
	m, err := NewMemoryStorer(time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	tm := memoryTestTimer{}
	ch := make(chan time.Time)
	timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
		return tm, ch
	}

	m.sessions["testid1"] = memorySession{
		value:   "test1",
		expires: time.Now().Add(time.Hour),
	}
	m.sessions["testid2"] = memorySession{
		value:   "test2",
		expires: time.Now().AddDate(0, 0, -1),
	}

	if len(m.sessions) != 2 {
		t.Error("expected len 2")
	}

	// Start the cleaner go routine
	m.StartCleaner()

	// Signal the timer channel to execute the clean
	ch <- time.Time{}

	// Stop the cleaner, this will block until the cleaner has finished its operations
	m.StopCleaner()

	if len(m.sessions) != 1 {
		t.Errorf("expected len 1, got %d", len(m.sessions))
	}

	_, ok := m.sessions["testid2"]
	if ok {
		t.Error("expected testid2 to be deleted, but was not")
	}
}

func TestMemoryStorerResetExpiry(t *testing.T) {
	t.Parallel()

	m, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Error(err)
	}

	err = m.Set("test", "val")
	if err != nil {
		t.Error(err)
	}

	sess := m.sessions["test"]
	oldExpires := sess.expires

	time.Sleep(time.Nanosecond * 1)

	err = m.ResetExpiry("test")
	if err != nil {
		t.Error(err)
	}

	sess = m.sessions["test"]
	newExpires := sess.expires

	if !newExpires.After(oldExpires) || newExpires == oldExpires {
		t.Errorf("Expected newexpires to be newer than old expires, got: %#v, %#v", oldExpires, newExpires)
	}
}
