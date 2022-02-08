package mongodb

var handlers []HandlerFunc

// HandlerFunc 执行中间件
type HandlerFunc func(op *OpTrace)

func AddMiddleware2(handler ...HandlerFunc) {
	handlers = append(handlers, handler...)
}

func do(f HandlerFunc, op *OpTrace) {
	op.handlers = append(handlers, f)
	if len(op.handlers) == 0 {
		return
	}
	// 执行对应的操作链
	op.handlers[0](op)
}
