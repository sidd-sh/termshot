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

package img_test

import (
	"bytes"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/homeport/termshot/internal/img"
)

type panicReader struct{}

func (p panicReader) Read(_ []byte) (int, error) {
	panic("synthetic panic")
}

var _ = Describe("Creating images", func() {
	Context("Use scaffold to create PNG file", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should write a PNG stream based on provided input", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-foobar.png")))
		})

		It("should omit the window decorations when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawDecorations(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-decoration.png")))
		})

		It("should omit the window shadow when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawShadow(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-shadow.png")))
		})

		It("should clip the canvas when configured", func() {
			scaffold := NewImageCreator()
			scaffold.ClipCanvas(true)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-clip-canvas.png")))
		})

		It("should wrap the content when configured", func() {
			scaffold := NewImageCreator()
			scaffold.SetColumns(4)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-wrapping.png")))
		})

		It("should show the command when configured", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddCommand("echo", "foobar")).To(Succeed())
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).To(Succeed())
			Expect(scaffold).To(LookLike(testdata("expected-show-cmd.png")))
		})

		It("should write a PNG stream based on provided input with ANSI sequences", func() {
			var buf bytes.Buffer
			_, _ = Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
			_, _ = Fprintf(&buf, "Colors:\n")
			_, _ = Fprintf(&buf, "\tRed{Red}\n")
			_, _ = Fprintf(&buf, "\tGreen{Green}\n")
			_, _ = Fprintf(&buf, "\tBlue{Blue}\n")
			_, _ = Fprintf(&buf, "\tMintCream{MintCream}\n")

			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(&buf)).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-ansi.png")))
		})
	})

	Context("Use scaffold to create raw output file", func() {
		var buf bytes.Buffer

		BeforeEach(func() {
			SetColorSettings(ON, ON)
			buf.Reset()
		})

		It("should write an output file with the content as-is", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader(Sprintf("MintCream{foobar}")))).To(Succeed())
			Expect(scaffold.WriteRaw(&buf)).To(Succeed())
			Expect(buf.String()).To(Equal("\x1b[38;2;245;255;250mfoobar\x1b[0m"))
		})
	})

	Context("Font configuration", func() {
		It("should support switching to hack font family", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.ConfigureFonts(FontConfig{Name: "hack"})).To(Succeed())
		})

		It("should reject unsupported font families", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.ConfigureFonts(FontConfig{Name: "unsupported"})).To(HaveOccurred())
		})

		It("should require regular font path for custom font files", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.ConfigureFonts(FontConfig{BoldPath: "font.ttf"})).To(HaveOccurred())
		})
	})

	Context("Malformed input handling", func() {
		It("should return an error when reader panics", func() {
			scaffold := NewImageCreator()
			err := scaffold.AddContent(panicReader{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("panic while parsing input stream"))
		})

		It("should not fail on cursor movements beyond the line start", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("abc\x1b[10Ddef\x1b[D\x1b[999DX\n"))).To(Succeed())

			var buf bytes.Buffer
			Expect(scaffold.WriteRaw(&buf)).To(Succeed())
			Expect(buf.String()).To(Equal("Xef\n"))
		})

		It("should not block on terminal queries in the output", func() {
			done := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(done)

				scaffold := NewImageCreator()
				Expect(scaffold.AddContent(strings.NewReader("a\x1b[6nb\x1b[cc\n"))).To(Succeed())
			}()

			Eventually(done).Should(BeClosed())
		})
	})

	Context("Terminal redraws in captured output", func() {
		var raw = func(input string) string {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader(input))).To(Succeed())

			var buf bytes.Buffer
			Expect(scaffold.WriteRaw(&buf)).To(Succeed())
			return buf.String()
		}

		It("should render a tab completion redraw as the final line only", func() {
			input := "PS C:\\> Get-Ch" + "\r\n" +
				"Get-ChildItem   Get-ChildHash\r\n" +
				"Get-ChildThing\r\n" +
				"\x1b[3A" + // cursor up 3
				"\x1b[J" + // erase to end of screen
				"\r" + "PS C:\\> Get-ChildItem"

			Expect(raw(input)).To(Equal(`PS C:\> Get-ChildItem`))
		})

		It("should render only the final state of a carriage return overwritten line", func() {
			Expect(raw("Loading |\rLoading /\rLoading -\rDone!     \nnext\n")).To(Equal("Done!\nnext\n"))
		})

		It("should keep output that scrolled off the screen", func() {
			var input, expected strings.Builder
			for i := 1; i <= 1000; i++ {
				fmt.Fprintf(&input, "line %d\r\n", i)
				fmt.Fprintf(&expected, "line %d\n", i)
			}

			Expect(raw(input.String())).To(Equal(expected.String()))
		})

		It("should keep the colors of plain colored output", func() {
			Expect(raw("\x1b[1mbold\x1b[0m \x1b[38;2;10;20;30mrgb\x1b[0m\r\n")).
				To(Equal("\x1b[1mbold\x1b[0m \x1b[38;2;10;20;30mrgb\x1b[0m\n"))
		})
	})
})
