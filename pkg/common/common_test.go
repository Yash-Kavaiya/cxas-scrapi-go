package common_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldMask_Single(t *testing.T) {
	assert.Equal(t, "displayName", common.FieldMask("displayName"))
}

func TestFieldMask_Multiple(t *testing.T) {
	assert.Equal(t, "displayName,description,rootAgent", common.FieldMask("displayName", "description", "rootAgent"))
}

func TestFieldMask_Empty(t *testing.T) {
	assert.Equal(t, "", common.FieldMask())
}

func TestSanitizeID_Deterministic(t *testing.T) {
	a := common.SanitizeID("Hello World")
	b := common.SanitizeID("Hello World")
	assert.Equal(t, a, b)
}

func TestSanitizeID_Format(t *testing.T) {
	id := common.SanitizeID("test input")
	assert.Len(t, id, 32)
	assert.Regexp(t, `^[a-f0-9]+$`, id)
}

func TestSanitizeID_DifferentInputs(t *testing.T) {
	assert.NotEqual(t, common.SanitizeID("foo"), common.SanitizeID("bar"))
}

func TestUnwrapValue_StringValue(t *testing.T) {
	v := map[string]interface{}{"string_value": "hello"}
	assert.Equal(t, "hello", common.UnwrapValue(v))
}

func TestUnwrapValue_NumberValue(t *testing.T) {
	v := map[string]interface{}{"number_value": float64(42)}
	assert.Equal(t, float64(42), common.UnwrapValue(v))
}

func TestUnwrapValue_BoolValue(t *testing.T) {
	v := map[string]interface{}{"bool_value": true}
	assert.Equal(t, true, common.UnwrapValue(v))
}

func TestUnwrapValue_NullValue(t *testing.T) {
	v := map[string]interface{}{"null_value": nil}
	assert.Nil(t, common.UnwrapValue(v))
}

func TestUnwrapValue_Nil(t *testing.T) {
	assert.Nil(t, common.UnwrapValue(nil))
}

func TestUnwrapStruct_Basic(t *testing.T) {
	s := map[string]interface{}{
		"fields": map[string]interface{}{
			"name": map[string]interface{}{"string_value": "Alice"},
			"age":  map[string]interface{}{"number_value": float64(30)},
		},
	}
	result := common.UnwrapStruct(s)
	require.NotNil(t, result)
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, float64(30), result["age"])
}

func TestPaginate_SinglePage(t *testing.T) {
	calls := 0
	fn := common.PageFunc[string](func(token string) ([]string, string, error) {
		calls++
		return []string{"a", "b"}, "", nil
	})
	items, err := common.Paginate(fn)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, items)
	assert.Equal(t, 1, calls)
}

func TestPaginate_MultiPage(t *testing.T) {
	pages := []struct {
		items     []string
		nextToken string
	}{
		{[]string{"a"}, "page2"},
		{[]string{"b"}, "page3"},
		{[]string{"c"}, ""},
	}
	idx := 0
	fn := common.PageFunc[string](func(token string) ([]string, string, error) {
		p := pages[idx]
		idx++
		return p.items, p.nextToken, nil
	})
	items, err := common.Paginate(fn)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, items)
}

func TestMapKeys(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": 2}
	keys := common.MapKeys(m)
	assert.Len(t, keys, 2)
	assert.ElementsMatch(t, []string{"a", "b"}, keys)
}
