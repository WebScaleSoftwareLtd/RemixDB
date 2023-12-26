// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import "testing"

func isAlphaNumeric(s string) bool {
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

func Test_randBytes(t *testing.T) {
	prevCombinations := map[string]struct{}{}
	b := make([]byte, 99)
	for i := 0; i < 100000; i++ {
		randBytes(b)
		if _, ok := prevCombinations[string(b)]; ok {
			t.Error("Duplicate bytes!")
		}
		prevCombinations[string(b)] = struct{}{}
	}
}

func Test_randString(t *testing.T) {
	prevCombinations := map[string]struct{}{}
	for i := 0; i < 100000; i++ {
		s := randString(100)
		if len(s) != 100 {
			t.Errorf("Incorrect length: %d", len(s))
		}
		if _, ok := prevCombinations[s]; ok {
			t.Errorf("Duplicate string: %s", s)
		}
		if !isAlphaNumeric(s) {
			t.Errorf("Non-alphanumeric string: %s", s)
		}
		prevCombinations[s] = struct{}{}
	}
}
