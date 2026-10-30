package project

import (
	"fmt"
	"regexp"
	"strings"
)

const maxNameLength = 100

var (
	// Letters, digits, hyphens and underscores, starting with a letter or
	// digit. Rules out path tricks like "../../etc" and anything that would
	// make a bad Go module or folder name.
	validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

	// Go module paths: slash-separated segments of safe characters.
	validModule = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]*(/[A-Za-z0-9._~-]+)*$`)

	// Branch, tag or commit.
	validRef = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
)

// Device names Windows won't allow for a folder, whatever the extension.
var windowsReserved = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true,
	"com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true,
	"lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// ValidateName checks a project name and says what's wrong with it, so the
// message can go straight to the user.
func ValidateName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("project name cannot be empty")
	case len(name) > maxNameLength:
		return fmt.Errorf("project name is too long (%d characters, max %d)", len(name), maxNameLength)
	case !validName.MatchString(name):
		return fmt.Errorf("invalid project name %q: use letters, numbers, hyphens and underscores only, starting with a letter or number", name)
	case windowsReserved[strings.ToLower(name)]:
		return fmt.Errorf("%q is a reserved name on Windows, please pick another", name)
	}
	return nil
}

func validateModule(module string) error {
	if !validModule.MatchString(module) || strings.Contains(module, "..") {
		return fmt.Errorf("invalid module path %q: expected something like github.com/acme/my-api", module)
	}
	return nil
}

func validateRef(ref string) error {
	if !validRef.MatchString(ref) || strings.Contains(ref, "..") {
		return fmt.Errorf("invalid --ref %q: use a branch, tag or commit", ref)
	}
	return nil
}
