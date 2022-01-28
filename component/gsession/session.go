package gsession

/**
 * TODO 存在的版本问题：
 * 1. 当有两个路由同时使用到session时，都是先读取库里的内容，初始化session的数据，然后进行增删改查，导致最后入库的时候可能不一致
 * 解决方式：增加old数据，new数据，对比两次数据与记录修改，加入版本控制，read版本跟write版本，加上读取版本锁？相对比较复杂，正常只要都是通用一个session中间件的话，没有问题。
 * 2. 当前有startSession的情况下，如果本身没有session，没有进行session更新操作（增删改），那么认为没有必要返回与保存这个session，不返回session。好处是，当session中间件被当成全局中间件（滥用）使用时，不会强行设置到库里去，而是不返回session。缺点是：对于规范来说，这里startSession但是没有创建与返回session。（通过变量 SetSession 控制）
 */

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/actorbuf/iota/component/gsession/driver"
	"github.com/gin-gonic/gin"
)

const SessionName = "IOTAGSESSION"
const SessionContextName string = "SessionContext"

var ErrSessionHasStarted = errors.New("session has start")

// Session session的部分
type Session struct {
	started   bool // 是否已经开启
	SessionId string
	Data      map[string]string // 具体数据
	Attribute Attribute         // 属性
	Driver    driver.Driver     // 存储处理
	// TODO 加一lock限制（文件类型的话session在不同项目下有可能有问题，就大家读的都是自己的项目，其他情况下的话也有可能有问题）
	SetSession bool // 是否有设置session, 记录是否有修改过session，如果没有的话，底层也需要写入session
}

type Attribute struct {
	Domain   string        `json:"domain" yaml:"domain"`
	Path     string        `json:"path" yaml:"path"`
	MaxAge   int           `json:"max-age" yaml:"max-age"`
	HttpOnly bool          `json:"http-only" yaml:"http-only"`
	Secure   bool          `json:"secure" yaml:"secure"`
	SameSite http.SameSite `json:"same-site" yaml:"same-site"`
}

func newSessionId() string {
	return RandomString()
}

// 获取SessionName
func getSessionName() string {
	return SessionName
}

// StartSession 初始化方法
func StartSession(c *gin.Context, drive driver.Driver, attribute Attribute) error {
	// 创建或者生成一个Session
	session := GetSession(c)
	if session != nil {
		if session.started {
			return ErrSessionHasStarted
		}
		return nil
	}
	sessionName := getSessionName()
	setSession := false

	sessionId, err := c.Cookie(sessionName)
	if err != nil && err != http.ErrNoCookie {
		return err
	}
	// 如果当前sessionId没有的话，重新创建一个
	if err == http.ErrNoCookie || len(sessionId) == 0 {
		sessionId = newSessionId()
	} else {
		// 再到库里查一把是否真实存在
		has, err := drive.HasSession(c, sessionId)
		if err != nil || !has {
			sessionId = newSessionId()
		} else {
			setSession = true // 如果以前有session的话，需要设置进去
		}
	}

	// 默认属性植入
	session = &Session{
		started:    true,
		Driver:     drive,
		SessionId:  sessionId,
		Data:       map[string]string{},
		Attribute:  attribute,
		SetSession: setSession,
	}

	// loading一把初始数据
	err = session.loadSession(c)
	if err != nil {
		return err
	}
	// 设置到上下文
	c.Set(SessionContextName, session)

	path := session.Attribute.Path
	if len(path) == 0 {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     getSessionName(),
		Value:    url.QueryEscape(session.SessionId),
		MaxAge:   session.Attribute.MaxAge,
		Path:     path,
		Domain:   session.Attribute.Domain,
		Secure:   session.Attribute.Secure,
		HttpOnly: session.Attribute.HttpOnly,
		SameSite: session.Attribute.SameSite,
	})
	return nil
}

// GetSession 从上下文获取session
func GetSession(c *gin.Context) *Session {
	session, exists := c.Get(SessionContextName)
	if !exists {
		return nil
	}
	return session.(*Session)
}

func (session *Session) loadSession(c *gin.Context) error {
	dataStr, err := session.Driver.Read(c, session.SessionId)
	if err != nil {
		return err
	}
	if len(dataStr) == 0 {
		session.Data = map[string]string{}
		return nil
	}

	// 先存一下数据
	err = json.Unmarshal([]byte(dataStr), &session.Data)
	if err != nil {
		return err
	}
	return nil
}

func (session *Session) setSession(c *gin.Context) {
	session.SetSession = true
}

func (session *Session) GetSessionId(c *gin.Context) string {
	return session.SessionId
}

// ============= 设置属性 =============

func (session *Session) SetDomain(domain string) *Session {
	session.Attribute.Domain = domain
	return session
}

func (session *Session) SetPath(path string) *Session {
	session.Attribute.Path = path
	return session
}

func (session *Session) SetMaxAge(maxAge int) *Session {
	session.Attribute.MaxAge = maxAge
	return session
}

func (session *Session) SetHttpOnly(httpOnly bool) *Session {
	session.Attribute.HttpOnly = httpOnly
	return session
}

func (session *Session) SetSecure(secure bool) *Session {
	session.Attribute.Secure = secure
	return session
}

func (session *Session) SetSameSite(sameSite http.SameSite) *Session {
	session.Attribute.SameSite = sameSite
	return session
}

// ========== 实现session的增删改查 ========

func (session *Session) Set(c *gin.Context, name string, value string) (bool, error) {
	// set 操作需要设置session
	session.setSession(c)
	session.Data[name] = value
	return true, nil
}

func (session *Session) Get(c *gin.Context, name string) (string, error) {
	return session.Data[name], nil
}

func (session *Session) GetByNames(c *gin.Context, names []string) (map[string]string, error) {
	resMap := map[string]string{}
	for _, name := range names {
		resMap[name] = session.Data[name]
	}
	return resMap, nil
}

func (session *Session) All(c *gin.Context) (map[string]string, error) {
	return session.Data, nil
}

func (session *Session) Remove(c *gin.Context, name string) (bool, error) {
	// Remove 操作需要设置session
	session.setSession(c)
	delete(session.Data, name)
	return true, nil
}

// Clear 增多一个清除session的操作，把key也给删掉
func (session *Session) Clear(c *gin.Context) (bool, error) {
	// Clear 操作需要设置session
	session.setSession(c)
	session.Data = map[string]string{}
	return true, nil
}

func (session *Session) Has(c *gin.Context, name string) (bool, error) {
	_, has := session.Data[name]
	return has, nil
}

func (session *Session) Save(c *gin.Context) error {
	// 如果不需要设置session的话，直接返回，防止session堆积
	if !session.SetSession {
		return nil
	}

	jsonData, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}
	err = session.Driver.Write(c, session.SessionId, string(jsonData), int64(session.Attribute.MaxAge))
	if err != nil {
		return err
	}
	return nil
}
