/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import "testing"

func TestUpgradeHint(t *testing.T) {
	tests := []struct {
		exe, goos, want string
	}{
		{"/opt/homebrew/Caskroom/dewlon/0.2.0/dewlon", "darwin", "brew upgrade dewlon"},
		{"/usr/local/Cellar/dewlon/0.2.0/bin/dewlon", "darwin", "brew upgrade dewlon"},
		{`C:\Users\me\scoop\apps\dewlon\current\dewlon.exe`, "windows", "scoop update dewlon"},
		{"/home/me/go/bin/dewlon", "linux", "go install github.com/dewlonsystems/dewlon@latest"},
	}
	for _, tc := range tests {
		if got := upgradeHint(tc.exe, tc.goos); got != tc.want {
			t.Errorf("upgradeHint(%q) = %q, want %q", tc.exe, got, tc.want)
		}
	}

	if got := upgradeHint("/usr/local/bin/dewlon", "linux"); got == "" {
		t.Error("expected the install script as a fallback")
	}
}
