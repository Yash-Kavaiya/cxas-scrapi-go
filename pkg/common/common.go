// Package common provides shared utilities for CXAS API clients.
package common

import (
	"crypto/md5" //nolint:gosec
	"fmt"
	"strings"
)

// FieldMask joins field names into a comma-separated updateMask query parameter value.
// Used with PATCH requests to the CES REST API.
func FieldMask(fields ...string) string {
	return strings.Join(fields, ",")
}

// SanitizeID returns a stable 32-char MD5 hex string from arbitrary text.
// Mirrors Python Common.sanitize_expectation_id().
func SanitizeID(text string) string {
	h := md5.Sum([]byte(text)) //nolint:gosec
	return fmt.Sprintf("%x", h)
}

// UnwrapValue decodes a google.protobuf.Value JSON map to a native Go value.
// Handles: null_value, bool_value, number_value, string_value, struct_value, list_value.
func UnwrapValue(v map[string]interface{}) interface{} {
	if v == nil {
		return nil
	}
	if _, ok := v["null_value"]; ok {
		return nil
	}
	if bv, ok := v["bool_value"]; ok {
		return bv
	}
	if nv, ok := v["number_value"]; ok {
		return nv
	}
	if sv, ok := v["string_value"]; ok {
		return sv
	}
	if sv, ok := v["struct_value"]; ok {
		if m, ok := sv.(map[string]interface{}); ok {
			return UnwrapStruct(m)
		}
	}
	if lv, ok := v["list_value"]; ok {
		if lm, ok := lv.(map[string]interface{}); ok {
			if vals, ok := lm["values"].([]interface{}); ok {
				result := make([]interface{}, len(vals))
				for i, el := range vals {
					if em, ok := el.(map[string]interface{}); ok {
						result[i] = UnwrapValue(em)
					} else {
						result[i] = el
					}
				}
				return result
			}
		}
	}
	return nil
}

// UnwrapStruct decodes a google.protobuf.Struct JSON representation into a plain Go map.
func UnwrapStruct(s map[string]interface{}) map[string]interface{} {
	fields, ok := s["fields"].(map[string]interface{})
	if !ok {
		return nil
	}
	result := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		if vm, ok := v.(map[string]interface{}); ok {
			result[k] = UnwrapValue(vm)
		}
	}
	return result
}

// MapKeys returns the keys of a map as a string slice.
func MapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
