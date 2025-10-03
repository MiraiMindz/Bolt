package bolt

import (
	"fmt"
	"strconv"
	"time"
)

// Field represents a strongly-typed key-value pair for zero-allocation operations
type Field struct {
	Key       string
	Type      FieldType
	IntVal    int64
	StringVal string
	BytesVal  []byte
	BoolVal   bool
	FloatVal  float64
	TimeVal   time.Time
	AnyVal    interface{}
}

// FieldType identifies the type of data stored in a Field
type FieldType uint8

const (
	UnknownType FieldType = iota
	StringType
	IntType
	Int64Type
	Float64Type
	BoolType
	BytesType
	TimeType
	DurationType
	AnyType
)

// Strongly-typed field constructors for zero-allocation JSON building

// String creates a string field
func String(key, val string) Field {
	return Field{Key: key, Type: StringType, StringVal: val}
}

// Bytes creates a bytes field (zero-allocation for pre-allocated slices)
func Bytes(key string, val []byte) Field {
	return Field{Key: key, Type: BytesType, BytesVal: val}
}

// Int creates an int field
func Int(key string, val int) Field {
	return Field{Key: key, Type: IntType, IntVal: int64(val)}
}

// Int64 creates an int64 field
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: Int64Type, IntVal: val}
}

// Float64 creates a float64 field
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: Float64Type, FloatVal: val}
}

// Bool creates a bool field
func Bool(key string, val bool) Field {
	return Field{Key: key, Type: BoolType, BoolVal: val}
}

// Time creates a time field
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: TimeType, TimeVal: val}
}

// Duration creates a duration field
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: DurationType, IntVal: int64(val)}
}

// Any creates a field with any value (will use reflection - slower)
func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: AnyType, AnyVal: val}
}

// fieldsToMap converts strongly-typed Fields to a map for JSON serialization
// This is optimized to minimize allocations
func fieldsToMap(fields []Field) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}

	m := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		switch f.Type {
		case StringType:
			m[f.Key] = f.StringVal
		case IntType, Int64Type:
			m[f.Key] = f.IntVal
		case Float64Type:
			m[f.Key] = f.FloatVal
		case BoolType:
			m[f.Key] = f.BoolVal
		case BytesType:
			m[f.Key] = string(f.BytesVal)
		case TimeType:
			m[f.Key] = f.TimeVal.Format(time.RFC3339)
		case DurationType:
			m[f.Key] = time.Duration(f.IntVal).String()
		case AnyType:
			m[f.Key] = f.AnyVal
		}
	}
	return m
}

// writeFieldsDirectJSON writes fields as JSON without intermediate allocations
// This is the fastest path for zero-allocation JSON responses
func writeFieldsDirectJSON(fields []Field) []byte {
	if len(fields) == 0 {
		return []byte("{}")
	}

	// Estimate buffer size to reduce reallocations
	buf := make([]byte, 0, len(fields)*64)
	buf = append(buf, '{')

	first := true
	for _, f := range fields {
		if !first {
			buf = append(buf, ',')
		}
		first = false

		// Write key
		buf = append(buf, '"')
		buf = append(buf, f.Key...)
		buf = append(buf, `":`...)

		// Write value based on type
		switch f.Type {
		case StringType:
			buf = append(buf, '"')
			buf = append(buf, f.StringVal...)
			buf = append(buf, '"')
		case IntType, Int64Type:
			buf = strconv.AppendInt(buf, f.IntVal, 10)
		case Float64Type:
			buf = strconv.AppendFloat(buf, f.FloatVal, 'g', -1, 64)
		case BoolType:
			if f.BoolVal {
				buf = append(buf, "true"...)
			} else {
				buf = append(buf, "false"...)
			}
		case BytesType:
			buf = append(buf, '"')
			buf = append(buf, f.BytesVal...)
			buf = append(buf, '"')
		case TimeType:
			buf = append(buf, '"')
			buf = append(buf, f.TimeVal.Format(time.RFC3339)...)
			buf = append(buf, '"')
		case DurationType:
			buf = append(buf, '"')
			buf = append(buf, time.Duration(f.IntVal).String()...)
			buf = append(buf, '"')
		case AnyType:
			// Fallback to string representation (still allocates, use JSONFields with typed fields for best performance)
			buf = append(buf, []byte(fmt.Sprintf("%v", f.AnyVal))...)
		}
	}

	buf = append(buf, '}')
	return buf
}

// writeFieldsDirectToWriter writes fields as JSON directly to a writer with zero allocations
// This avoids the intermediate buffer allocation of writeFieldsDirectJSON
func writeFieldsDirectToWriter(w ResponseWriter, fields []Field) error {
	if len(fields) == 0 {
		_, err := w.Write([]byte("{}"))
		return err
	}

	// Use a stack-allocated buffer for small responses (common case)
	var stackBuf [512]byte
	buf := stackBuf[:0]

	buf = append(buf, '{')

	first := true
	for _, f := range fields {
		if !first {
			buf = append(buf, ',')
		}
		first = false

		// Write key
		buf = append(buf, '"')
		buf = append(buf, f.Key...)
		buf = append(buf, `":`...)

		// Write value based on type
		switch f.Type {
		case StringType:
			buf = append(buf, '"')
			// Check if buffer is getting full, flush if needed
			if len(buf)+len(f.StringVal)+2 > cap(buf) {
				if _, err := w.Write(buf); err != nil {
					return err
				}
				buf = stackBuf[:0]
			}
			buf = append(buf, f.StringVal...)
			buf = append(buf, '"')
		case IntType, Int64Type:
			buf = strconv.AppendInt(buf, f.IntVal, 10)
		case Float64Type:
			buf = strconv.AppendFloat(buf, f.FloatVal, 'g', -1, 64)
		case BoolType:
			if f.BoolVal {
				buf = append(buf, "true"...)
			} else {
				buf = append(buf, "false"...)
			}
		case BytesType:
			buf = append(buf, '"')
			if len(buf)+len(f.BytesVal)+2 > cap(buf) {
				if _, err := w.Write(buf); err != nil {
					return err
				}
				buf = stackBuf[:0]
			}
			buf = append(buf, f.BytesVal...)
			buf = append(buf, '"')
		case TimeType:
			buf = append(buf, '"')
			buf = append(buf, f.TimeVal.Format(time.RFC3339)...)
			buf = append(buf, '"')
		case DurationType:
			buf = append(buf, '"')
			buf = append(buf, time.Duration(f.IntVal).String()...)
			buf = append(buf, '"')
		case AnyType:
			// Fallback to string representation (still allocates, use JSONFields with typed fields for best performance)
			buf = append(buf, []byte(fmt.Sprintf("%v", f.AnyVal))...)
		}
	}

	buf = append(buf, '}')
	_, err := w.Write(buf)
	return err
}
