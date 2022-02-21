package core

// ErrStack 包含error堆栈的结构体
type ErrStack struct {
	Hint   string  // 当进行链路追踪 配置tracer id
	errors []error // error 堆栈
}

// Error 返回ErrStack string
func (es *ErrStack) Error() string {
	var msg string
	for _, err := range es.errors {
		msg = err.Error() + "\n"
	}
	return msg
}

// Wrap 追加error
func (es *ErrStack) Wrap(err ...error) *ErrStack {
	es.errors = append(es.errors, err...)
	return es
}

// SetHint 进行链路追踪时 配置tracerID
func (es *ErrStack) SetHint(hint string) *ErrStack {
	es.Hint = hint
	return es
}
