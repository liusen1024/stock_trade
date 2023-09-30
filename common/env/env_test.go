package env

import (
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	env, err := LoadEnv("../test_data/test_dev.json")
	if err != nil {
		t.Fatal(err)
	}
	if env.IsTest() != true {
		t.Fatal()
	}

	os.Setenv("A", "value-a")
	a, ok := env.Get("A")
	if !ok {
		t.Fatal()
	}
	if a != "value-a" {
		t.Fatal()
	}

	os.Setenv("ANOTHER_KEY", "value-b")
	b, ok := env.Get("B")
	if !ok {
		t.Fatal()
	}
	if b != "value-b" {
		t.Fatal()
	}
	os.Setenv("C", "")
	c, ok := env.Get("C")
	if !ok {
		t.Fatal()
	}
	if c != "value-c" {
		t.Fatal()
	}

	_, ok = env.Get("D")
	if ok {
		t.Fatal()
	}

}
