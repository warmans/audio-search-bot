package bot

import (
	"encoding/json"
	"fmt"
)

type customIDOpt func(c *CustomID)

func withModifier(mod ContentModifier) customIDOpt {
	return func(c *CustomID) {
		c.ContentModifier = mod
	}
}

func withStartLine(pos int32) customIDOpt {
	return func(c *CustomID) {
		c.StartLine = pos
	}
}
func withEndLine(pos int32) customIDOpt {
	return func(c *CustomID) {
		c.EndLine = pos
	}
}

type CustomID struct {
	MediaID         string          `json:"e,omitempty"`
	StartLine       int32           `json:"s,omitempty"`
	EndLine         int32           `json:"f,omitempty"`
	NumContextLines int             `json:"c,omitempty"`
	ContentModifier ContentModifier `json:"t,omitempty"`
}

func (c CustomID) String() string {
	data, err := json.Marshal(c)
	if err != nil {
		// this should never happen
		fmt.Printf("failed to encode customID: %s\n", err.Error())
		return ""
	}
	return string(data)
}

func (c CustomID) withOption(options ...customIDOpt) CustomID {
	clone := &CustomID{
		MediaID:         c.MediaID,
		StartLine:       c.StartLine,
		EndLine:         c.EndLine,
		NumContextLines: c.NumContextLines,
		ContentModifier: c.ContentModifier,
	}
	for _, v := range options {
		v(clone)
	}
	return *clone
}

type ContentModifier uint8

const (
	ContentModifierNone ContentModifier = iota
	ContentModifierTextOnly
	ContentModifierAudioOnly
)

func encodeCustomIDForAction(action string, customID CustomID) string {
	return fmt.Sprintf("%s:%s", action, customID.String())
}

func decodeCustomIDPayload(data string) (CustomID, error) {
	decoded := &CustomID{}
	return *decoded, json.Unmarshal([]byte(data), decoded)
}
