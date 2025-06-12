package crawler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripMarkdownLinks(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "strip link",
			input:  "prefix [caption](url) suffix",
			expect: "prefix caption suffix",
		},
		{
			name:   "strip image",
			input:  "prefix ![caption 1](url1) text ![caption 2](url2) suffix",
			expect: "prefix caption 1 text caption 2 suffix",
		},
		{
			name:   "strip image without caption",
			input:  "prefix ![](url) suffix ![](url)",
			expect: "prefix suffix",
		},
		{
			name:   "strip image without caption",
			input:  "character![](/wiki/File:Dr_John_Zoidberg.png)First appearance",
			expect: "character First appearance",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := stripMarkdownLinks(tc.input)

			require.Equal(t, tc.expect, actual)
		})
	}
}
