package sinks

type Sink interface {
	Config()
	Push() error
}
