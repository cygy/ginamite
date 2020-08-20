package random

import "testing"

func TestRandomMessage(t *testing.T) {
	for i := 0; i <= 1000; i++ {
		string := String(i)

		if len(string) != i {
			t.Errorf("Generate a random string with %d characters must return s tring with %d characters.", i, i)
		}
	}
}
