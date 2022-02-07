package mongodb

type Callback interface {
	Before(op *OpTrace) error // 执行前钩子
	After(op *OpTrace) error  // 执行后钩子
}

var middlewareCallback []Callback

func AddMiddleware(middleware Callback) {
	middlewareCallback = append(middlewareCallback, middleware)
}

// middlewareBefore 前置中间件
func middlewareBefore(opTrace *OpTrace) error {
	for _, hook := range middlewareCallback {
		err := hook.Before(opTrace)
		if err != nil {
			return err
		}
	}
	return nil
}

// middlewareAfter 后置中间件
func middlewareAfter(opTrace *OpTrace) error {
	// 逆序执行after
	for i := len(middlewareCallback) - 1; i >= 0; i-- {
		hook := middlewareCallback[i]
		err := hook.After(opTrace)
		if err != nil {
			return err
		}
	}
	return nil
}
