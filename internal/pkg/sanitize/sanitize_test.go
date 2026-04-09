package sanitize_test

import (
	"testing"

	"simple-blog-api/internal/pkg/sanitize"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\t\nhello\t\n", "hello"},
		{"hello", "hello"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitize.Trim(tt.input))
		})
	}
}

func TestSanitizeStrict_RemovesHTML(t *testing.T) {
	input := `<script>alert("xss")</script>Hello`
	result := sanitize.SanitizeStrict(input)
	assert.NotContains(t, result, "<script>")
	assert.Contains(t, result, "Hello")
}

func TestSanitizeStrict_RemovesAllTags(t *testing.T) {
	input := `<b>bold</b> <i>italic</i> plain`
	result := sanitize.SanitizeStrict(input)
	assert.NotContains(t, result, "<b>")
	assert.NotContains(t, result, "<i>")
	assert.Contains(t, result, "bold")
	assert.Contains(t, result, "italic")
	assert.Contains(t, result, "plain")
}

func TestSanitizeUGC_AllowsSafeFormatting(t *testing.T) {
	input := `<b>bold</b> <p>paragraph</p> <a href="https://example.com">link</a>`
	result := sanitize.SanitizeUGC(input)
	assert.Contains(t, result, "<b>bold</b>")
	assert.Contains(t, result, "<p>paragraph</p>")
}

func TestSanitizeUGC_RemovesScript(t *testing.T) {
	input := `<script>evil()</script><p>content</p>`
	result := sanitize.SanitizeUGC(input)
	assert.NotContains(t, result, "<script>")
	assert.Contains(t, result, "content")
}

func TestSanitizeUGC_RemovesOnClickAttribute(t *testing.T) {
	input := `<a onclick="evil()" href="https://example.com">click</a>`
	result := sanitize.SanitizeUGC(input)
	assert.NotContains(t, result, "onclick")
}

func TestTrimAndSanitizeStrict(t *testing.T) {
	input := `  <b>hello</b>  `
	result := sanitize.SanitizeStrict(sanitize.Trim(input))
	assert.Equal(t, "hello", result)
}
