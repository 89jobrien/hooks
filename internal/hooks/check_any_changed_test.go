package hooks

import (
	"strings"
	"testing"
)

func TestCheckAnyChanged_NonWriteTool(t *testing.T) {
	input := HookInput{
		ToolName: "Shell",
		ToolInput: []byte(`{"command": "ls"}`),
	}
	result, code := CheckAnyChanged(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-Write tool, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCheckAnyChanged_NonTypeScriptFile(t *testing.T) {
	input := writeInput("app.go", `package main
func main() {
	var x any
}`)
	result, code := CheckAnyChanged(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for non-TypeScript file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCheckAnyChanged_TypeScriptWithAny(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		contents string
		wantDeny bool
	}{
		{
			name:     "type annotation any",
			file:     "app.ts",
			contents: "const x: any = 5;",
			wantDeny: true,
		},
		{
			name:     "as any cast",
			file:     "component.tsx",
			contents: "const y = x as any;",
			wantDeny: true,
		},
		{
			name:     "generic any",
			file:     "util.ts",
			contents: "function f<T = any>() {}",
			wantDeny: true,
		},
		{
			name:     "Array<any>",
			file:     "data.ts",
			contents: "const arr: Array<any> = [];",
			wantDeny: true,
		},
		{
			name:     "expect.any() allowed",
			file:     "test.ts",
			contents: "expect(value).toBe(expect.any(String));",
			wantDeny: false,
		},
		{
			name:     ".any() allowed",
			file:     "test.ts",
			contents: "mockFn.mockReturnValue(jest.any(String));",
			wantDeny: false,
		},
		{
			name:     "clean TypeScript",
			file:     "app.ts",
			contents: "const x: string = 'hello';",
			wantDeny: false,
		},
		{
			name:     "multiple any types",
			file:     "bad.ts",
			contents: "const a: any = 1;\nconst b: any = 2;",
			wantDeny: true,
		},
		{
			name:     "any in function parameter",
			file:     "handler.ts",
			contents: "function handle(data: any) {}",
			wantDeny: true,
		},
		{
			name:     "any in return type",
			file:     "api.ts",
			contents: "function get(): any {}",
			wantDeny: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := writeInput(tt.file, tt.contents)
			result, code := CheckAnyChanged(input)
			if tt.wantDeny {
				if code != 2 || result.Decision != "deny" {
					t.Errorf("expected deny, got decision=%s code=%d", result.Decision, code)
				}
				if result.Reason == "" {
					t.Error("expected reason in deny result")
				}
				if !strings.Contains(result.Reason, "any") {
					t.Errorf("expected reason to mention 'any', got %q", result.Reason)
				}
			} else {
				if code != 0 || result.Decision != "allow" {
					t.Errorf("expected allow, got decision=%s code=%d", result.Decision, code)
				}
			}
		})
	}
}

func TestCheckAnyChanged_IncludesLineNumbers(t *testing.T) {
	contents := `const x: string = 'hello';
const y: any = 5;
const z: number = 10;
const w: any = 'world';`
	input := writeInput("app.ts", contents)
	result, code := CheckAnyChanged(input)
	if code != 2 {
		t.Fatalf("expected deny, got code=%d", code)
	}
	if !strings.Contains(result.Reason, "Line 2") {
		t.Errorf("expected reason to include line 2, got %q", result.Reason)
	}
	if !strings.Contains(result.Reason, "Line 4") {
		t.Errorf("expected reason to include line 4, got %q", result.Reason)
	}
}

func TestCheckAnyChanged_EmptyFile(t *testing.T) {
	input := writeInput("empty.ts", "")
	result, code := CheckAnyChanged(input)
	if code != 0 || result.Decision != "allow" {
		t.Errorf("expected allow for empty file, got decision=%s code=%d", result.Decision, code)
	}
}

func TestCheckAnyChanged_TSXFile(t *testing.T) {
	input := writeInput("component.tsx", "const props: any = {};")
	result, code := CheckAnyChanged(input)
	if code != 2 || result.Decision != "deny" {
		t.Errorf("expected deny for .tsx file with any, got decision=%s code=%d", result.Decision, code)
	}
}

