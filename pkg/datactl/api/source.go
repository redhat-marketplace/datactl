package api

import (
	"encoding"
	"fmt"
)

type SourceType string

const (
	DataService SourceType = "DataService"
)

var _ encoding.TextMarshaler = SourceType("")
var _ encoding.TextUnmarshaler = SourceType("")

func (s SourceType) String() (text string) {
	return string(s)
}

func (s SourceType) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%s", s)), nil
}

func (s SourceType) UnmarshalText(text []byte) error {
	switch string(text) {
	case DataService.String():
		s = DataService
	default:
		return fmt.Errorf("source type not defined")
	}
	return nil
}
