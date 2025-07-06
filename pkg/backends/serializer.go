package backends

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// JSONSerializer implements JSON serialization.
type JSONSerializer struct{}

// Serialize serializes data to JSON.
func (j *JSONSerializer) Serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Deserialize deserializes JSON data.
func (j *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// ContentType returns the content type for JSON.
func (j *JSONSerializer) ContentType() string {
	return "application/json"
}

// GobSerializer implements Go's gob serialization.
type GobSerializer struct{}

// Serialize serializes data using gob.
func (g *GobSerializer) Serialize(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes gob data.
func (g *GobSerializer) Deserialize(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(target)
}

// ContentType returns the content type for gob.
func (g *GobSerializer) ContentType() string {
	return "application/gob"
}

// MsgPackSerializer implements MessagePack serialization.
// Note: This is a placeholder implementation. In a real implementation,
// you would use a library like github.com/vmihailenco/msgpack/v5
type MsgPackSerializer struct{}

// Serialize serializes data using MessagePack (placeholder).
func (m *MsgPackSerializer) Serialize(data interface{}) ([]byte, error) {
	// For now, fall back to JSON
	// In a real implementation, use msgpack library
	return json.Marshal(data)
}

// Deserialize deserializes MessagePack data (placeholder).
func (m *MsgPackSerializer) Deserialize(data []byte, target interface{}) error {
	// For now, fall back to JSON
	// In a real implementation, use msgpack library
	return json.Unmarshal(data, target)
}

// ContentType returns the content type for MessagePack.
func (m *MsgPackSerializer) ContentType() string {
	return "application/msgpack"
}

// SerializeValue is a helper function to serialize any value using the configured serializer.
func SerializeValue(serializer Serializer, value interface{}) ([]byte, error) {
	// If value is already []byte, return as-is
	if byteValue, ok := value.([]byte); ok {
		return byteValue, nil
	}

	// If value is string, convert to []byte
	if strValue, ok := value.(string); ok {
		return []byte(strValue), nil
	}

	// For other types, use the serializer
	return serializer.Serialize(value)
}

// DeserializeValue is a helper function to deserialize a value using the configured serializer.
func DeserializeValue(serializer Serializer, data []byte, target interface{}) error {
	// If target is *[]byte, assign directly
	if bytePtr, ok := target.(*[]byte); ok {
		*bytePtr = data
		return nil
	}

	// If target is *string, convert from []byte
	if strPtr, ok := target.(*string); ok {
		*strPtr = string(data)
		return nil
	}

	// For other types, use the serializer
	return serializer.Deserialize(data, target)
}

// AutoDetectType tries to automatically detect and deserialize the value to its original type.
func AutoDetectType(data []byte) interface{} {
	// Try to detect if it's JSON
	var jsonValue interface{}
	if err := json.Unmarshal(data, &jsonValue); err == nil {
		return jsonValue
	}

	// If not JSON, return as string
	return string(data)
}
