package hooks

import (
	"os"
	"testing"
)

func TestIsHookDisabled_Unset(t *testing.T) {
	os.Unsetenv("HOOK_DISABLED")
	if IsHookDisabled("validate-shell") {
		t.Error("expected false when HOOK_DISABLED unset")
	}
}

func TestIsHookDisabled_Single(t *testing.T) {
	os.Setenv("HOOK_DISABLED", "validate-shell")
	defer os.Unsetenv("HOOK_DISABLED")
	if !IsHookDisabled("validate-shell") {
		t.Error("expected true when hook in HOOK_DISABLED")
	}
	if IsHookDisabled("audit") {
		t.Error("expected false for other hook")
	}
}

func TestIsHookDisabled_List(t *testing.T) {
	os.Setenv("HOOK_DISABLED", "rate-limiter, secret-scanner ,audit")
	defer os.Unsetenv("HOOK_DISABLED")
	if !IsHookDisabled("rate-limiter") {
		t.Error("expected true for rate-limiter")
	}
	if !IsHookDisabled("secret-scanner") {
		t.Error("expected true for secret-scanner")
	}
	if !IsHookDisabled("audit") {
		t.Error("expected true for audit")
	}
	if IsHookDisabled("validate-shell") {
		t.Error("expected false for validate-shell")
	}
}

func TestIsHookDisabled_Empty(t *testing.T) {
	os.Setenv("HOOK_DISABLED", "")
	defer os.Unsetenv("HOOK_DISABLED")
	if IsHookDisabled("validate-shell") {
		t.Error("expected false when HOOK_DISABLED empty")
	}
}
