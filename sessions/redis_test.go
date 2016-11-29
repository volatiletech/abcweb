package sessions

import (
	"testing"
	"time"

	redis "gopkg.in/redis.v5"
)

func TestRedisStorerNew(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping long test")
	}

	r := redis.Options{
		Password: "test",
	}
	storer, err := NewRedisStorer(r, 2)
	if err != nil {
		t.Error(err)
	}

	if storer.maxAge != 2 {
		t.Error("expected max age to be 2")
	}

	if storer.client == nil {
		t.Error("Expected client to be created")
	}
}

func TestRedisStorerNewDefault(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping long test")
	}

	storer, err := NewDefaultRedisStorer("", "", 0)
	if err != nil {
		t.Error(err)
	}

	if storer.client == nil {
		t.Error("Expected client to be created")
	}

	if storer.maxAge != time.Hour*24*7 {
		t.Error("expected max age to be a week")
	}
}

func TestRedisStorerGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	storer, err := NewDefaultRedisStorer("", "", 0)
	if err != nil {
		t.Error(err)
	}

	val, err := storer.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	storer.Put("fruit", "banana")

	val, err = storer.Get("fruit")
	if err != nil {
		t.Error(err)
	}
	if val != "banana" {
		t.Errorf("Expected %q, got %q", "banana", val)
	}
}

func TestRedisStorerPut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	storer, err := NewDefaultRedisStorer("", "", 0)
	if err != nil {
		t.Error(err)
	}

	storer.Put("hi", "hello")
	storer.Put("hi", "whatsup")
	storer.Put("yo", "friend")

	val, err := storer.Get("hi")
	if err != nil {
		t.Fatal(err)
	}
	if val != "whatsup" {
		t.Errorf("Expected %q, got %q", "whatsup", val)
	}

	val, err = storer.Get("yo")
	if err != nil {
		t.Error(err)
	}
	if val != "friend" {
		t.Errorf("Expected %q, got %q", "friend", val)
	}
}

func TestRedisStorerDel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test")
	}

	storer, err := NewDefaultRedisStorer("", "", 0)
	if err != nil {
		t.Error(err)
	}

	storer.Put("hi", "hello")
	storer.Put("hi", "whatsup")
	storer.Put("yo", "friend")

	err = storer.Del("hi")
	if err != nil {
		t.Error(err)
	}

	_, err = storer.Get("hi")
	if err == nil {
		t.Errorf("Expected get hi to fail")
	}
}
