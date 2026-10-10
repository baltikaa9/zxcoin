package logger

type Logger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)

	With(kv ...any) Logger
	Named(name string) Logger
}

type nop struct{}

func (nop) Debug(string, ...any)  {}
func (nop) Info(string, ...any)   {}
func (nop) Warn(string, ...any)   {}
func (nop) Error(string, ...any)  {}
func (n nop) With(...any) Logger  { return n }
func (n nop) Named(string) Logger { return n }

func Nop() Logger { return nop{} }
