package main

import (
	"testing"
)

func TestMain(t *testing.T) {
}

func TestAll(t *testing.T) {
	got := all("foo", "bar", "baz")
	if got != true {
		t.Error("got false, want true")
	}
	got = all("foo", "", "baz")
	if got == true {
		t.Error("got true, want false")
	}
	got = all("", "", "")
	if got == true {
		t.Error("got true, want false")
	}
}

func TestAny(t *testing.T) {
	got := any("foo", "bar", "baz")
	if got != true {
		t.Error("got false, want true")
	}
	got = any("", "bar", "")
	if got != true {
		t.Error("got false, want true")
	}
	got = any("", "", "")
	if got == true {
		t.Error("got true, want false")
	}
}

func TestCast(t *testing.T) {
	if got, want := cast("5"), int(5); got != want {
		t.Errorf("bad cast: got %v, want %v", got, want)
	}
	if got, want := cast("5.5"), float64(5.5); got != want {
		t.Errorf("bad cast: got %v, want %v", got, want)
	}
	if got, want := cast("foo"), "foo"; got != want {
		t.Errorf("bad cast: got %v, want %v", got, want)
	}
	if got, want := cast(""), (interface{})(nil); got != want {
		t.Errorf("bad cast: got %v, want %v", got, want)
	}
}
