package config

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/yc-alpha/variant"
)

type Content struct {
	content map[string]any
}

func NewContent() *Content {
	return &Content{content: make(map[string]any)}
}

func (c *Content) Merge(other *Content) {
	for key, value := range other.content {
		// if the current value and the value in other have the same type, merge them.
		if existing, ok := c.content[key]; ok {
			// if the value is a map type, merge the map recursively.
			if existingMap, ok := existing.(map[string]any); ok {
				if otherMap, ok := value.(map[string]any); ok {
					// recursively merge nested maps.
					mergeMap(existingMap, otherMap)
				} else {
					// otherwise, replace directly.
					c.content[key] = value
				}
			} else if existingSlice, ok := existing.([]any); ok {
				// if the value is a slice type, merge the slice recursively.
				if otherSlice, ok := value.([]any); ok {
					// recursively merge nested slices.
					mergeSlice(existingSlice, otherSlice)
				} else {
					c.content[key] = value
				}
			} else {
				c.content[key] = value
			}
		} else {
			// if the key does not exist, add it directly.
			c.content[key] = value
		}
	}
}

// mergeSlice is used to recursively merge nested map.
func mergeMap(existing, other map[string]any) {
	for key, value := range other {
		// if the key exists and the value is a map type, continue to merge recursively.
		if existingValue, ok := existing[key]; ok {
			if existingMap, ok := existingValue.(map[string]any); ok {
				if otherMap, ok := value.(map[string]any); ok {
					mergeMap(existingMap, otherMap)
					continue
				}
			}
		}
		// if the key does not exist or the value is not a map type, replace directly.
		existing[key] = value
	}
}

// mergeSlice is used to recursively merge nested slice.
func mergeSlice(existing, other []any) {
	// merge the slice element by element.
	for i := 0; i < len(other); i++ {
		if i < len(existing) {
			// if the element exists, merge recursively.
			existingValue := existing[i]
			otherValue := other[i]
			if reflect.TypeOf(existingValue) == reflect.TypeOf(otherValue) {
				switch v := existingValue.(type) {
				case map[string]any:
					if otherMap, ok := otherValue.(map[string]any); ok {
						mergeMap(v, otherMap)
					}
				case []any:
					if otherSlice, ok := otherValue.([]any); ok {
						mergeSlice(v, otherSlice)
					}
				}
			} else {
				// if the type is different, replace directly.
				existing[i] = other[i]
			}
		} else {
			// if the length of existing is smaller, append directly.
			existing = append(existing, other[i])
		}
	}
}

// Get retrieves a configuration value by its key, supporting nested keys using dot notation (e.g., "database.host")
func (c *Content) Get(path string) variant.Variant {

	parts := parsePath(path)

	var current any = c.content

	for _, part := range parts {
		if part.kind == Index {
			index, err := strconv.Atoi(part.value)
			if err != nil {
				return variant.Nil
			}
			currentSlice, ok := current.([]any)
			if !ok || index >= len(currentSlice) || index < 0 {
				return variant.Nil
			}
			current = currentSlice[index]
		} else if part.kind == Key {
			currentMap, ok := current.(map[string]any)
			if !ok {
				return variant.Nil
			}
			val, exists := currentMap[part.value]
			if !exists {
				return variant.Nil
			}
			current = val
		}
	}

	return variant.New(current)
}

func (c *Content) GetDefault(path string, _default any) variant.Variant {
	value := c.Get(path)
	if value.Equal(nil) {
		return variant.New(_default)
	}
	return value
}

// Set sets a configuration value by its key, supporting nested keys using dot notation (e.g., "database.host").
func (c *Content) Set(path string, value any) error {

	parts := parsePath(path)

	var parent any
	var current any = c.content
	for i, part := range parts {
		// fmt.Printf("parent: %+v, current: %+v\n\n", parent, current)
		if part.kind == Index {
			index, err := strconv.Atoi(part.value)
			if err != nil {
				return err
			}

			currentSlice, ok := current.([]any)
			if !ok {
				return errors.New("invalid type")
			}

			// expand the length of the slice.
			if index > len(currentSlice) || index < 0 {
				return errors.New("index out of range")
			}
			if index == len(currentSlice) {
				newSlice := make([]any, index+1)
				copy(newSlice, currentSlice)
				currentSlice = newSlice

				if part.parent.kind == Key {
					parentMap, _ := parent.(map[string]any)
					parentMap[part.parent.value] = newSlice
				} else {
					parentSlice, _ := parent.([]any)
					parentIndex, _ := strconv.Atoi(part.parent.value)
					parentSlice[parentIndex] = newSlice
				}
			}

			// set the value if it is a leaf node.
			if part.isLeaf {
				currentSlice[index] = value
				return nil
			}

			if currentSlice[index] == nil {
				if parts[i+1].kind == Index {
					currentSlice[index] = make([]any, 0)
				} else {
					currentSlice[index] = make(map[string]any)
				}
			}
			parent = currentSlice
			current = currentSlice[index]
		} else if part.kind == Key {
			currentMap, ok := current.(map[string]any)
			if !ok {
				return errors.New("invalid type")
			}

			// set the value if it is a leaf node.
			if part.isLeaf {
				currentMap[part.value] = value
				return nil
			}

			// create a new map/slice if the key does not exist.a.b.c
			if _, exists := currentMap[part.value]; !exists {
				if parts[i+1].kind == Index {
					currentMap[part.value] = make([]any, 0)
				} else {
					currentMap[part.value] = make(map[string]any)
				}
			}
			parent = currentMap
			current = currentMap[part.value]
		}
	}

	return nil
}
