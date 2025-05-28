// Package cutjson provides functionality to extract specific parts of a JSON object.
package cutjson

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

var (
	// ErrInvalidJSON is returned when the input is not a valid JSON.
	ErrInvalidJSON = errors.New("invalid JSON input")
	// ErrPathNotFound is returned when the specified path is not found in the JSON.
	ErrPathNotFound = errors.New("path not found in JSON")
	// ErrInvalidPath is returned when the path format is invalid.
	ErrInvalidPath = errors.New("invalid path format")
	// ErrInvalidRule is returned when the rule format is invalid.
	ErrInvalidRule = errors.New("invalid rule format")
)

// Rule represents a JSON cutting rule
type Rule struct {
	Type      RuleType    // 规则类型
	Path      string      // JSON路径
	Value     interface{} // 配置值（用于规则2和规则3）
	ChildPath string      // 子路径（用于规则3）
}

// RuleType defines the type of cutting rule
type RuleType int

const (
	// KeepPath 规则1: 保留指定路径
	KeepPath RuleType = iota
	// KeepParentIfValueMatches 规则2: 如果值匹配，保留父路径
	KeepParentIfValueMatches
	// KeepArrayElementsIfChildValueMatches 规则3: 如果数组元素的子路径值匹配，保留该元素
	KeepArrayElementsIfChildValueMatches
)

// NewKeepPathRule creates a rule to keep a specific JSON path
func NewKeepPathRule(path string) Rule {
	return Rule{
		Type: KeepPath,
		Path: path,
	}
}

// NewKeepParentIfValueMatchesRule creates a rule to keep parent path if child value matches
func NewKeepParentIfValueMatchesRule(path string, value interface{}) Rule {
	return Rule{
		Type:  KeepParentIfValueMatches,
		Path:  path,
		Value: value,
	}
}

// NewKeepArrayElementsIfChildValueMatchesRule creates a rule to keep array elements if child value matches
func NewKeepArrayElementsIfChildValueMatchesRule(arrayPath string, childPath string, value interface{}) Rule {
	return Rule{
		Type:      KeepArrayElementsIfChildValueMatches,
		Path:      arrayPath,
		ChildPath: childPath,
		Value:     value,
	}
}

// CutWithRules cuts a JSON object based on the provided rules
func CutWithRules(jsonData []byte, rules []Rule) ([]byte, error) {
	if len(jsonData) == 0 {
		return nil, ErrInvalidJSON
	}

	// Parse the JSON data
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, ErrInvalidJSON
	}

	// Apply each rule
	result, err := applyRules(data, rules)
	if err != nil {
		return nil, err
	}

	// Marshal the result back to JSON
	output, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// applyRules applies all rules to the JSON data
func applyRules(data interface{}, rules []Rule) (interface{}, error) {
	result := make(map[string]interface{})

	for _, rule := range rules {
		switch rule.Type {
		case KeepPath:
			if err := applyKeepPathRule(data, rule.Path, result); err != nil && err != ErrPathNotFound {
				return nil, err
			}

		case KeepParentIfValueMatches:
			if err := applyKeepParentIfValueMatchesRule(data, rule.Path, rule.Value, result); err != nil && err != ErrPathNotFound {
				return nil, err
			}

		case KeepArrayElementsIfChildValueMatches:
			if err := applyKeepArrayElementsIfChildValueMatchesRule(data, rule.Path, rule.ChildPath, rule.Value, result); err != nil && err != ErrPathNotFound {
				return nil, err
			}

		default:
			return nil, ErrInvalidRule
		}
	}

	return result, nil
}

// applyKeepPathRule applies rule type 1: keep the specified path
func applyKeepPathRule(data interface{}, path string, result map[string]interface{}) error {
	// Split the path into segments
	pathSegments := strings.Split(path, ".")

	// Navigate to the value
	value, err := navigateToValue(data, pathSegments)
	if err != nil {
		return err
	}

	// Build the result structure
	buildResultStructure(result, pathSegments, value)

	return nil
}

// applyKeepParentIfValueMatchesRule applies rule type 2: if the value at path matches, keep the parent path
func applyKeepParentIfValueMatchesRule(data interface{}, path string, expectedValue interface{}, result map[string]interface{}) error {
	// Split the path into segments
	pathSegments := strings.Split(path, ".")

	// Navigate to the value
	value, err := navigateToValue(data, pathSegments)
	if err != nil {
		return err
	}

	// Check if the value matches
	if !valueEquals(value, expectedValue) {
		return nil
	}

	// Keep the parent path (remove the last segment)
	if len(pathSegments) > 1 {
		parentPathSegments := pathSegments[:len(pathSegments)-1]
		parentValue, err := navigateToValue(data, parentPathSegments)
		if err != nil {
			return err
		}

		// Build the result structure for the parent
		buildResultStructure(result, parentPathSegments, parentValue)
	} else {
		// If there's no parent (top-level field), keep the whole field
		buildResultStructure(result, pathSegments, value)
	}

	return nil
}

// applyKeepArrayElementsIfChildValueMatchesRule applies rule type 3: keep array elements where child value matches
func applyKeepArrayElementsIfChildValueMatchesRule(data interface{}, arrayPath string, childPath string, expectedValue interface{}, result map[string]interface{}) error {
	// Split the array path into segments
	arrayPathSegments := strings.Split(arrayPath, ".")

	// Navigate to the array
	arrayValue, err := navigateToValue(data, arrayPathSegments)
	if err != nil {
		return err
	}

	// Check if it's an array
	array, ok := arrayValue.([]interface{})
	if !ok {
		return errors.New("path does not point to an array")
	}

	// Split the child path into segments
	childPathSegments := strings.Split(childPath, ".")

	// Filter the array elements
	filteredArray := make([]interface{}, 0)
	for _, element := range array {
		// Try to navigate to the child value
		childValue, err := navigateToValue(element, childPathSegments)
		if err == nil && valueEquals(childValue, expectedValue) {
			filteredArray = append(filteredArray, element)
		}
	}

	// If we found matching elements, build the result structure
	if len(filteredArray) > 0 {
		buildResultStructure(result, arrayPathSegments, filteredArray)
	}

	return nil
}

// navigateToValue traverses the JSON structure following the path segments
func navigateToValue(data interface{}, pathSegments []string) (interface{}, error) {
	if len(pathSegments) == 0 {
		return data, nil
	}

	currentSegment := pathSegments[0]
	remaining := pathSegments[1:]

	switch v := data.(type) {
	case map[string]interface{}:
		val, ok := v[currentSegment]
		if !ok {
			return nil, ErrPathNotFound
		}
		return navigateToValue(val, remaining)

	case []interface{}:
		// Try to parse the segment as an array index
		index := 0
		for i, c := range currentSegment {
			if i == 0 && c == '-' {
				continue
			}
			if c < '0' || c > '9' {
				return nil, ErrInvalidPath
			}
			index = index*10 + int(c-'0')
		}

		// Handle negative index (count from the end)
		if currentSegment[0] == '-' {
			index = -index
		}

		// Convert negative index to positive
		if index < 0 {
			index = len(v) + index
		}

		if index < 0 || index >= len(v) {
			return nil, ErrPathNotFound
		}

		return navigateToValue(v[index], remaining)

	default:
		return nil, ErrPathNotFound
	}
}

// buildResultStructure builds a nested structure in the result map based on path segments
func buildResultStructure(result map[string]interface{}, pathSegments []string, value interface{}) {
	if len(pathSegments) == 0 {
		return
	}

	current := result
	for i, segment := range pathSegments {
		if i == len(pathSegments)-1 {
			// Last segment, set the value
			current[segment] = value
		} else {
			// Create intermediate objects if needed
			if _, exists := current[segment]; !exists {
				// Check if the next segment is a number (array index)
				isNextSegmentArrayIndex := false
				if i+1 < len(pathSegments) {
					nextSegment := pathSegments[i+1]
					isNextSegmentArrayIndex = isNumeric(nextSegment)
				}

				if isNextSegmentArrayIndex {
					current[segment] = make([]interface{}, 0)
				} else {
					current[segment] = make(map[string]interface{})
				}
			}

			// Move to the next level
			if nextMap, ok := current[segment].(map[string]interface{}); ok {
				current = nextMap
			} else {
				// Handle array case
				// This is a simplified implementation and might need enhancement for complex cases
				break
			}
		}
	}
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Check for negative sign
	start := 0
	if s[0] == '-' {
		if len(s) == 1 {
			return false
		}
		start = 1
	}

	// Check if all remaining characters are digits
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// valueEquals checks if two values are equal
func valueEquals(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// Cut extracts a portion of a JSON object based on the given path (for backward compatibility)
func Cut(jsonData []byte, path string) ([]byte, error) {
	rules := []Rule{NewKeepPathRule(path)}
	return CutWithRules(jsonData, rules)
}

// CutMultiple extracts multiple portions of a JSON object based on the given paths (for backward compatibility)
func CutMultiple(jsonData []byte, paths []string) (map[string]json.RawMessage, error) {
	result := make(map[string]json.RawMessage)

	for _, path := range paths {
		rules := []Rule{NewKeepPathRule(path)}
		cut, err := CutWithRules(jsonData, rules)
		if err != nil {
			result[path] = nil
			continue
		}
		result[path] = cut
	}

	return result, nil
}
