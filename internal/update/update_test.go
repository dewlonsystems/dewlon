package update

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "v1.2.3", 0},
		{"v1.2.3", "1.2.3", 0},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.10.0", "v1.9.9", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"v1.0.0-rc.1", "v1.0.0", -1},
		{"v1.0.0-rc.2", "v1.0.0-rc.10", -1},
		{"v1.0.0-alpha", "v1.0.0-alpha.1", -1},
		{"v1.0.0+build5", "v1.0.0", 0},
	}

	for _, tc := range tests {
		got, err := Compare(tc.a, tc.b)
		if err != nil {
			t.Errorf("Compare(%q, %q) error: %v", tc.a, tc.b, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompareRejectsGarbage(t *testing.T) {
	for _, bad := range []string{"dev", "1.2", "a.b.c", ""} {
		if _, err := Compare(bad, "v1.0.0"); err == nil {
			t.Errorf("Compare(%q, ...) should fail", bad)
		}
	}
}

func TestIsPseudo(t *testing.T) {
	v, err := parse("v0.0.0-20260101120000-abcdef123456")
	if err != nil || !isPseudo(v) {
		t.Errorf("pseudo version not recognised: %+v, %v", v, err)
	}
}
