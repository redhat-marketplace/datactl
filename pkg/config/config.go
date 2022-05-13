package config

type Accessor interface {
	Get(string) interface{}
	GetString(string) string
	GetInt(string) int
}
