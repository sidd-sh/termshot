// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func wrapErrorWithRecoveryDump(context string, cause error, captured []byte) error {
	if len(captured) == 0 {
		return fmt.Errorf("%s: %w", context, cause)
	}

	path, err := saveRecoveryDump(captured, context)
	if err != nil {
		return fmt.Errorf("%s: %w (also failed to save captured output: %v)", context, cause, err)
	}

	return fmt.Errorf("%s: %w (captured output saved to %q)", context, cause, path)
}

func saveRecoveryDump(captured []byte, reason string) (string, error) {
	if len(captured) == 0 {
		return "", fmt.Errorf("captured output is empty")
	}

	baseDir, err := os.UserCacheDir()
	if err != nil || baseDir == "" {
		baseDir = os.TempDir()
	}

	dumpDir := filepath.Join(baseDir, "termshot", "recovery")
	if mkErr := os.MkdirAll(dumpDir, 0o755); mkErr != nil {
		return "", fmt.Errorf("failed to create recovery directory %q: %w", dumpDir, mkErr)
	}

	timestamp := time.Now().Format("20060102-150405.000000000")
	fileName := fmt.Sprintf("termshot-%s-%s-%d.log", sanitizeReason(reason), timestamp, os.Getpid())
	path := filepath.Join(dumpDir, fileName)

	if writeErr := os.WriteFile(path, captured, 0o600); writeErr != nil {
		return "", fmt.Errorf("failed to write recovery file %q: %w", path, writeErr)
	}

	return path, nil
}

func sanitizeReason(reason string) string {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, string(filepath.Separator), "-")
	normalized = strings.ReplaceAll(normalized, ":", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}

	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "unknown"
	}

	return normalized
}
