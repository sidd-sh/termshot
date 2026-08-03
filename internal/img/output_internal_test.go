package img

import (
	"strings"
	"testing"

	"github.com/gonvenience/bunt"
)

func TestEnhanceRawContentDoesNotInsertBlankLineForIndentedContinuation(t *testing.T) {
	scaffold := NewImageCreator()
	parsed, err := bunt.ParseStream(strings.NewReader("❯ echo foo\n  bar\n"))
	if err != nil {
		t.Fatalf("parse stream: %v", err)
	}

	enhanced := scaffold.enhanceRawContent(parsed)
	if got := strings.Count(enhanced.String(), "\n"); got != 2 {
		t.Fatalf("expected 2 newlines, got %d in %q", got, enhanced.String())
	}
}
