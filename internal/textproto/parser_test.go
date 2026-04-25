package textproto_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/textproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_SimpleStringField(t *testing.T) {
	result := textproto.Parse(`display_name: "My App"`)
	require.NotNil(t, result)
	assert.Equal(t, "My App", result["display_name"])
}

func TestParse_NumberField(t *testing.T) {
	result := textproto.Parse(`count: 42`)
	assert.Equal(t, "42", result["count"])
}

func TestParse_BooleanTrue(t *testing.T) {
	result := textproto.Parse(`enabled: true`)
	assert.Equal(t, true, result["enabled"])
}

func TestParse_BooleanFalse(t *testing.T) {
	result := textproto.Parse(`enabled: false`)
	assert.Equal(t, false, result["enabled"])
}

func TestParse_NestedMessage(t *testing.T) {
	input := `
model_settings {
  model: "gemini-2.0-flash"
}
`
	result := textproto.Parse(input)
	nested, ok := result["model_settings"].(map[string]interface{})
	require.True(t, ok, "model_settings should be a map")
	assert.Equal(t, "gemini-2.0-flash", nested["model"])
}

func TestParse_RepeatedField(t *testing.T) {
	input := `
tag: "P0"
tag: "regression"
tag: "smoke"
`
	result := textproto.Parse(input)
	tags, ok := result["tag"].([]interface{})
	require.True(t, ok, "tag should be a slice")
	assert.Len(t, tags, 3)
	assert.Equal(t, "P0", tags[0])
}

func TestParse_MultipleFields(t *testing.T) {
	input := `
display_name: "Test App"
description: "A test"
root_agent: "projects/p/locations/l/apps/a/agents/root"
`
	result := textproto.Parse(input)
	assert.Equal(t, "Test App", result["display_name"])
	assert.Equal(t, "A test", result["description"])
	assert.Equal(t, "projects/p/locations/l/apps/a/agents/root", result["root_agent"])
}

func TestParse_IgnoresComments(t *testing.T) {
	input := `
# This is a comment
name: "value" # inline comment
`
	result := textproto.Parse(input)
	assert.Equal(t, "value", result["name"])
}

func TestParse_EscapedStringField(t *testing.T) {
	input := `text: "hello\nworld"`
	result := textproto.Parse(input)
	assert.Equal(t, "hello\nworld", result["text"])
}

func TestParse_EmptyInput(t *testing.T) {
	result := textproto.Parse("")
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestParse_DeepNesting(t *testing.T) {
	input := `
outer {
  inner {
    value: "deep"
  }
}
`
	result := textproto.Parse(input)
	outer, ok := result["outer"].(map[string]interface{})
	require.True(t, ok)
	inner, ok := outer["inner"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "deep", inner["value"])
}
