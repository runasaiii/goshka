package logger
import "log"


type Interface interface {
	Debug(msg string, keysAndValues ...interface{})
}

type logger struct{}

func New() Interface {
	return &logger{}
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	log.Println(append([]interface{}{msg}, keysAndValues...)...)
}
