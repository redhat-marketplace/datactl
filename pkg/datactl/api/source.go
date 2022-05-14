package api

import (
	"fmt"
)

type SourceType string

const (
	DataService SourceType = "DataService"
)

func (s SourceType) String() (text string) {
	return string(s)
}

func (s SourceType) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%s", s)), nil
}

func (s *SourceType) UnmarshalText(text []byte) error {
	switch string(text) {
	case DataService.String():
		*s = DataService
	default:
		return fmt.Errorf("source type %s not defined", text)
	}
	return nil
}
