package searchterms

import (
	"fmt"
	"time"
)

type Value interface {
	// Type returns the type for the value.
	Type() Type
	// Value returns the value with the correct Go type as an interface{}.
	Value() interface{}
	// String formats the value as a string.
	String() string
}

func String(s string) StringValue {
	return StringValue(s)
}

type StringValue string

func (s StringValue) Type() Type {
	return StringType
}

func (s StringValue) Value() interface{} {
	return string(s)
}

func (s StringValue) String() string {
	return fmt.Sprintf(`"%s"`, string(s))
}

func Int(v int64) IntValue {
	return IntValue(v)
}

type IntValue int64

func (s IntValue) Type() Type {
	return IntType
}

func (s IntValue) Value() interface{} {
	return int64(s)
}

func (s IntValue) String() string {
	return fmt.Sprint(int64(s))
}

func Duration(ts time.Duration) DurationValue {
	return DurationValue(ts)
}

type DurationValue time.Duration

func (s DurationValue) Type() Type {
	return DurationType
}

func (s DurationValue) Value() interface{} {
	return time.Duration(s)
}

func (s DurationValue) String() string {
	return time.Duration(s).String()
}
