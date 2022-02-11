package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
)

// BindGroupRouteSrv 使用组路由注册的结构体都需要实现这个接口
type BindGroupRouteSrv interface {
	// Bind 返回这个结构体绑定了哪个 proto service 的名称
	Bind() string
}

// Register 是一个组路由注射器
// 只能通过NewRegister()进行实例化 否则会panic
type Register struct {
	// 路由配置文件 是proto同级的 autogen_router_module.go 文件中的 GroupRouterMap
	routeMap map[string]*GroupRouter
}

// NewRegister 实例化注册器
func NewRegister() *Register {
	var register = &Register{}
	return register
}

// BindRouteMap 绑定自动生成的路由配置文件
// 入参m: proto同级的 autogen_router_module.go 文件中的 GroupRouterMap
func (r *Register) BindRouteMap(m map[string]*GroupRouter) *Register {
	r.routeMap = m
	return r
}

// RegisterStruct 按照 struct 的方法进行路由注册
// rout: gin路由 建议传入group 将公共的中间件传递入group中
// igs: 需要注册的API组的struct ptr
// 对于错误或异常零容忍 直接panic
func (r *Register) RegisterStruct(rout gin.IRouter, igs ...interface{}) {
	if r.routeMap == nil {
		panic("route map nil, run *Register.BindRouteMapConfig() to bind rout map")
	}

	if len(igs) == 0 {
		panic("group struct empty")
	}

	for _, ig := range igs {
		bind := ig.(BindGroupRouteSrv) // 你需要实现 BindGroupRouteSrv
		// 输出一下 rg
		refVal := reflect.ValueOf(ig)
		refTyp := reflect.TypeOf(ig)

		routConfig := r.routeMap[bind.Bind()]
		if routConfig == nil {
			panic("no func to register")
		}
		if routConfig.Apis == nil {
			panic("no func to register")
		}
		// 注册路由公共中间件
		if len(routConfig.Middlewares) != 0 {
			r.registerMiddleware(rout, routConfig.Middlewares)
		}
		routMap := routConfig.Apis
		for m := 0; m < refTyp.NumMethod(); m++ {
			// 这里取出方法
			method := refTyp.Method(m)
			if method.Name == "Bind" {
				continue
			}

			var routc *GroupRouterNode
			var exist bool
			if routc, exist = routMap[method.Name]; !exist {
				continue
			}
			if routConfig.RouterPrefix != "" {
				routc.API = routConfig.RouterPrefix + routc.API
			}
			// 注册路由
			if err := r.registerHandle(rout, routc, method.Func, refVal); err != nil {
				logrus.Errorf("err: %+v", err)
				panic("err: " + err.Error())
			}
		}
	}
}

// registerMiddleware 注册中间件
func (r *Register) registerMiddleware(router gin.IRouter, mws []gin.HandlerFunc) {
	router.Use(mws...)
}

// registerHandle 注册Handle
func (r *Register) registerHandle(router gin.IRouter, rc *GroupRouterNode, rFunc, rGroup reflect.Value) error {
	call, err := r.getCallFunc(rFunc, rGroup)
	if err != nil {
		return err
	}
	if call == nil {
		return nil
	}

	var hfs []gin.HandlerFunc
	if len(rc.Middlewares) != 0 {
		hfs = append(hfs, rc.Middlewares...)
		hfs = append(hfs, call)
	} else {
		hfs = append(hfs, call)
	}

	switch rc.Method {
	case http.MethodPost:
		router.POST(rc.API, hfs...)
	case http.MethodGet:
		router.GET(rc.API, hfs...)
	case http.MethodDelete:
		router.DELETE(rc.API, hfs...)
	case http.MethodPatch:
		router.PATCH(rc.API, hfs...)
	case http.MethodPut:
		router.PUT(rc.API, hfs...)
	case http.MethodOptions:
		router.OPTIONS(rc.API, hfs...)
	case http.MethodHead:
		router.HEAD(rc.API, hfs...)
	case "ANY":
		router.Any(rc.API, hfs...)
	default:
		return fmt.Errorf("method:[%v] not support", rc.Method)
	}
	return nil
}

// getCallFunc 获取运行函数入口
func (r *Register) getCallFunc(rFunc, rGroup reflect.Value) (gin.HandlerFunc, error) {
	typ := rFunc.Type() // 获取函数的类型

	// 参数检查
	if typ.NumIn() != 3 {
		return nil, fmt.Errorf("func need two request param, (ctx, req)")
	}

	// 响应检查
	if typ.NumOut() != 2 {
		return nil, fmt.Errorf("func need two response param, (resp, error)")
	}

	// 第二返回参数是否是error
	if returnType := typ.Out(1); returnType != reflect.TypeOf((*error)(nil)).Elem() {
		return nil, errors.Errorf("method : %v , returns[1] %v not error",
			runtime.FuncForPC(rFunc.Pointer()).Name(), returnType.String())
	}

	ctxType, reqType := typ.In(1), typ.In(2)
	if ctxType != reflect.TypeOf(&Context{}) {
		return nil, fmt.Errorf("first param must *core.Context")
	}

	if reqType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("req type not ptr")
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("err: %+v\nstack: %+v", err, string(debug.Stack()))
				return
			}
		}()

		req := reflect.New(reqType.Elem())
		// 参数校验
		err := r.bindAndValidate(c, req.Interface())
		if err != nil {
			c.JSON(http.StatusOK, struct {
				ErrCode int    `json:"err_code"`
				ErrMsg  string `json:"err_msg"`
			}{
				ErrCode: ErrInvalidArg,
				ErrMsg:  err.Error(),
			})
			return
		}

		var returnValues = rFunc.Call([]reflect.Value{rGroup, reflect.ValueOf(&Context{Context: c}), req})

		// 重定向的情况
		if c.Writer.Status() == http.StatusFound || c.Writer.Status() == http.StatusMovedPermanently {
			c.Abort()
			return
		}

		// 传输文件直接下载的情况
		ct := c.Writer.Header().Get("Content-Type")
		if ct == "application/octet-stream" {
			c.Abort()
			return
		}

		if returnValues != nil {
			resp := returnValues[0].Interface()
			rerr := returnValues[1].Interface()

			if rerr == nil {
				c.JSON(http.StatusOK, Result{
					ErrCode: ErrNil,
					ErrMsg:  "ok",
					Data:    ResponseCompatible(resp),
				})
				return
			}

			var err error
			var errCode int
			var errMsg string

			var isAutonomy bool
			if reflect.TypeOf(rerr).String() == "*core.ErrMsg" {
				e := rerr.(*ErrMsg)
				if e.Autonomy {
					err = e
					errCode = int(e.ErrCode)
					errMsg = e.ErrMsg
					isAutonomy = true
				}
			}

			if !isAutonomy {
				err = rerr.(error)
				errCode = GetErrCode(err)
				errMsg = GetErrMsg(int32(errCode))
			}

			c.JSON(http.StatusOK, Result{
				ErrCode: errCode,
				ErrMsg:  errMsg,
				Data:    ResponseCompatible(resp),
			})
			return
		}
	}, nil
}

// bindAndValidate 绑定并校验参数
func (r *Register) bindAndValidate(c *gin.Context, req interface{}) error {
	if err := c.ShouldBind(req); err != nil {
		return err
	}
	return nil
}
