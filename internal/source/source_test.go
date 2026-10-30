package source

import (
	"context"
	"strings"
	"testing"

	"github.com/dewlonsystems/dewlon/internal/templates"
)

const sampleSum = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestParseSums(t *testing.T) {
	data := []byte(sampleSum + "  template.zip\n" +
		strings.Repeat("b", 64) + " *other.zip\n")

	got, err := ParseSums(data, "template.zip")
	if err != nil || got != sampleSum {
		t.Fatalf("ParseSums = %q, %v", got, err)
	}
	if _, err := ParseSums(data, "missing.zip"); err == nil {
		t.Error("expected an error for a file that isn't listed")
	}
	if _, err := ParseSums([]byte("nothex  template.zip\n"), "template.zip"); err == nil {
		t.Error("expected an error for an invalid checksum")
	}
}

func TestVerify(t *testing.T) {
	plan := Plan{Ref: "v1.0.0", Want: sampleSum}
	if err := Verify(plan, strings.ToUpper(sampleSum)); err != nil {
		t.Errorf("matching checksum rejected: %v", err)
	}
	if err := Verify(plan, strings.Repeat("c", 64)); err == nil {
		t.Error("mismatch should fail")
	}
	if err := Verify(Plan{}, "anything"); err != nil {
		t.Errorf("a plan with no checksum should pass: %v", err)
	}
}

func TestResolveBranchWithPin(t *testing.T) {
	tpl, err := templates.Get("fiber")
	if err != nil {
		t.Fatal(err)
	}

	plan, err := Resolve(context.Background(), tpl, Options{Ref: "main", SHA256: "sha256:" + sampleSum})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Checked() || plan.Want != sampleSum {
		t.Errorf("expected the pin to become the expected checksum, got %+v", plan)
	}

	if _, err := Resolve(context.Background(), tpl, Options{Ref: "main", SHA256: "nope"}); err == nil {
		t.Error("a malformed --sha256 should be rejected")
	}
}

func TestResolveBranchIsUnverified(t *testing.T) {
	tpl, _ := templates.Get("fiber")
	plan, err := Resolve(context.Background(), tpl, Options{Ref: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Checked() {
		t.Error("a branch can't be verified without a pin")
	}
}
