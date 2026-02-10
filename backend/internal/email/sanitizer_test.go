package email

import (
	"strings"
	"testing"
)

func TestHTMLToText_SimpleHTML(t *testing.T) {
	got := HTMLToText("<p>Hello</p>")
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestHTMLToText_StripsScriptAndStyle(t *testing.T) {
	input := `<html><head><style>body{color:red}</style></head><body>
		<script>alert('xss')</script><p>Visible</p></body></html>`
	got := HTMLToText(input)
	if strings.Contains(got, "alert") {
		t.Errorf("output contains script content: %q", got)
	}
	if strings.Contains(got, "color:red") {
		t.Errorf("output contains style content: %q", got)
	}
	if !strings.Contains(got, "Visible") {
		t.Errorf("output missing visible text: %q", got)
	}
}

func TestHTMLToText_PreservesBreaks(t *testing.T) {
	input := `<p>First</p><p>Second</p><div>Third</div>`
	got := HTMLToText(input)
	if !strings.Contains(got, "First") || !strings.Contains(got, "Second") || !strings.Contains(got, "Third") {
		t.Errorf("missing text content: %q", got)
	}
	// Paragraphs should produce newlines
	parts := strings.Fields(got)
	if len(parts) < 3 {
		t.Errorf("expected at least 3 words, got %d: %q", len(parts), got)
	}
}

func TestHTMLToText_NestedElements(t *testing.T) {
	input := `<div><span><strong>Deep</strong> <em>nested</em></span> text</div>`
	got := HTMLToText(input)
	if !strings.Contains(got, "Deep") || !strings.Contains(got, "nested") || !strings.Contains(got, "text") {
		t.Errorf("missing nested text: %q", got)
	}
}

func TestHTMLToText_EmptyInput(t *testing.T) {
	got := HTMLToText("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestHTMLToText_PlainText(t *testing.T) {
	got := HTMLToText("Just plain text")
	if !strings.Contains(got, "Just plain text") {
		t.Errorf("got %q, want to contain %q", got, "Just plain text")
	}
}
