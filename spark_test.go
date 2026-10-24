package main

import "testing"

func TestSparkDoesNotMutateInput(t *testing.T) {
	values := []float64{1, 2, 3}

	if got, want := Spark(values), "▁▄▇"; got != want {
		t.Fatalf("Spark(values) = %q, want %q", got, want)
	}
	if values[0] != 1 || values[1] != 2 || values[2] != 3 {
		t.Fatalf("Spark mutated its input: got %v, want [1 2 3]", values)
	}
}
