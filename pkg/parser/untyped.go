package parser

type AnyType struct {
	V any
}

// Get an array from AnyType or nil if cast fails.
func (t AnyType) Array() []interface{} {
	v := t.V.(any)
	if v != nil {
		if arr, ok := v.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// Get an object (Map) from AnyType or nil if cast fails.
func (t AnyType) Object() map[string]interface{} {
	v := t.V.(any)
	if v != nil {
		if obj, ok := v.(map[string]interface{}); ok {
			return obj
		}
	}
	return nil
}

// Get an AnyType from object (Map) by a given key.
func (t AnyType) Get(key string) AnyType {
	v := t.Object()
	if v != nil {
		if val, exists := v[key]; exists {
			return AnyType{V: val}
		}
	}
	return AnyType{}
}
