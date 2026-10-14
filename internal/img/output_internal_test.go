package img

import (
	"strings"
	"testing"
)

func TestEnhanceRawContentDoesNotInsertBlankLineForIndentedContinuation(t *testing.T) {
	scaffold := NewImageCreator()
	parsed := scaffold.emulate([]byte("❯ echo foo\n  bar\n"))

	enhanced := scaffold.enhanceRawContent(&parsed)
	if got := strings.Count(enhanced.String(), "\n"); got != 2 {
		t.Fatalf("expected 2 newlines, got %d in %q", got, enhanced.String())
	}
}
