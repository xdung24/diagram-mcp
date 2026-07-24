package mcpserver

import (
	"os"
	"strings"
	"testing"
)

func TestApplyGanttDefaults(t *testing.T) {
	t.Run("injects axisFormat and excludes when missing", func(t *testing.T) {
		input := "gantt\n    title Project\n    dateFormat YYYY-MM-DD\n    section Phase 1\n    " +
			"Design :d1, 2024-01-01, 5d\n"

		got, err := applyGanttDefaults(input)
		if err != nil {
			t.Fatalf("applyGanttDefaults() error = %v", err)
		}
		if !strings.Contains(got, "axisFormat %m-%d") {
			t.Errorf("output missing default axisFormat, got:\n%s", got)
		}
		if !strings.Contains(got, "excludes weekends") {
			t.Errorf("output missing default excludes, got:\n%s", got)
		}
	})

	t.Run("leaves explicit axisFormat and excludes untouched", func(t *testing.T) {
		input := "gantt\n    title Project\n    axisFormat %d/%m\n    excludes 2024-12-25\n    " +
			"section Phase 1\n    Design :d1, 2024-01-01, 5d\n"

		got, err := applyGanttDefaults(input)
		if err != nil {
			t.Fatalf("applyGanttDefaults() error = %v", err)
		}
		if !strings.Contains(got, "axisFormat %d/%m") {
			t.Errorf("output should keep explicit axisFormat, got:\n%s", got)
		}
		if strings.Contains(got, "axisFormat %m-%d") {
			t.Errorf("output should not override explicit axisFormat, got:\n%s", got)
		}
		if !strings.Contains(got, "excludes 2024-12-25") {
			t.Errorf("output should keep explicit excludes, got:\n%s", got)
		}
	})

	t.Run("invalid content returns error", func(t *testing.T) {
		if _, err := applyGanttDefaults("not a gantt chart"); err == nil {
			t.Error("expected error for invalid gantt content")
		}
	})
}

func TestApplyArchitectureTitleQuoting(t *testing.T) {
	t.Run("quotes titles with hyphens and parentheses", func(t *testing.T) {
		input := "architecture-beta\n" +
			"    group pipeline(cloud)[AI-Powered Pipeline]\n" +
			"    service workers(server)[Worker Pool (Python)] in pipeline\n"

		got, err := applyArchitectureTitleQuoting(input)
		if err != nil {
			t.Fatalf("applyArchitectureTitleQuoting() error = %v", err)
		}
		if !strings.Contains(got, `["AI-Powered Pipeline"]`) {
			t.Errorf("output missing quoted group title, got:\n%s", got)
		}
		if !strings.Contains(got, `["Worker Pool (Python)"]`) {
			t.Errorf("output missing quoted service title, got:\n%s", got)
		}
	})

	t.Run("leaves plain word titles unquoted", func(t *testing.T) {
		input := "architecture-beta\n    group api(cloud)[API]\n    service db(database)[Database] in api\n"

		got, err := applyArchitectureTitleQuoting(input)
		if err != nil {
			t.Fatalf("applyArchitectureTitleQuoting() error = %v", err)
		}
		if !strings.Contains(got, "group api(cloud)[API]") {
			t.Errorf("output should keep unquoted API title, got:\n%s", got)
		}
		if !strings.Contains(got, "service db(database)[Database] in api") {
			t.Errorf("output should keep unquoted Database title, got:\n%s", got)
		}
	})

	t.Run("invalid content returns error", func(t *testing.T) {
		if _, err := applyArchitectureTitleQuoting("not an architecture diagram"); err == nil {
			t.Error("expected error for invalid architecture content")
		}
	})
}

func TestApplyPacketNormalization(t *testing.T) {
	t.Run("quotes unquoted field labels", func(t *testing.T) {
		input := "packet\n    0-7: Version\n    8-15: Type\n"

		got, err := applyPacketNormalization(input)
		if err != nil {
			t.Fatalf("applyPacketNormalization() error = %v", err)
		}
		if !strings.Contains(got, `0-7: "Version"`) {
			t.Errorf("output missing quoted Version label, got:\n%s", got)
		}
		if !strings.Contains(got, `8-15: "Type"`) {
			t.Errorf("output missing quoted Type label, got:\n%s", got)
		}
	})

	t.Run("leaves already-quoted labels untouched", func(t *testing.T) {
		input := "packet\n    0-15: \"Source Port\"\n"

		got, err := applyPacketNormalization(input)
		if err != nil {
			t.Fatalf("applyPacketNormalization() error = %v", err)
		}
		if !strings.Contains(got, `0-15: "Source Port"`) {
			t.Errorf("output missing quoted Source Port label, got:\n%s", got)
		}
	})

	t.Run("invalid content returns error", func(t *testing.T) {
		if _, err := applyPacketNormalization("not a packet diagram"); err == nil {
			t.Error("expected error for invalid packet content")
		}
	})
}

func TestResolveContent(t *testing.T) {
	t.Run("content only", func(t *testing.T) {
		got, err := resolveContent(sourceInput{Content: "flowchart TB\n"})
		if err != nil {
			t.Fatalf("resolveContent() error = %v", err)
		}
		if got != "flowchart TB\n" {
			t.Errorf("resolveContent() = %q", got)
		}
	})

	t.Run("path only", func(t *testing.T) {
		f := t.TempDir() + "/diagram.mermaid"
		want := "flowchart TB\n    A --> B\n"
		if err := os.WriteFile(f, []byte(want), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		got, err := resolveContent(sourceInput{Path: f})
		if err != nil {
			t.Fatalf("resolveContent() error = %v", err)
		}
		if got != want {
			t.Errorf("resolveContent() = %q, want %q", got, want)
		}
	})

	t.Run("neither content nor path", func(t *testing.T) {
		if _, err := resolveContent(sourceInput{}); err == nil {
			t.Error("expected error when neither content nor path is set")
		}
	})

	t.Run("both content and path", func(t *testing.T) {
		if _, err := resolveContent(sourceInput{Content: "x", Path: "y"}); err == nil {
			t.Error("expected error when both content and path are set")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		if _, err := resolveContent(sourceInput{Path: "/nonexistent/does-not-exist.mermaid"}); err == nil {
			t.Error("expected error for nonexistent path")
		}
	})
}
