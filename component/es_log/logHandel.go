package es_log

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/actorbuf/iota/driver/rabbitmq"
	log2 "log"
	"time"
)

const LevelInfo = "INFO"
const LevelWarning = "WARNING"
const LevelError = "ERROR"
const LevelDebug = "DEBUG"
const LevelNotice = "NOTICE"

type MqData struct {
	Project string `json:"project"`
	Data    struct {
		Msg       string                 `json:"msg"`
		Level     string                 `json:"level"`
		Channel   string                 `json:"channel"`
		CreatedAt int64                  `json:"created_at"`
		Define    map[string]interface{} `json:"define"`
	} `json:"data"`
}

type log struct {
	MqData      MqData `json:"mq_data"`
	ServerName  string
	Environment string
	Logger      Logger
	Alarm       Alarm
	MqLogger    rabbitmq.Logger
	MaAddress   string
}

type Alarm interface {
	Info(ctx context.Context, step string, str string)
	Error(ctx context.Context, step string, err error)
}

type Logger interface {
	Info(ctx context.Context, step string, str string)
	Error(ctx context.Context, step string, err error)
}

type FmtInfoFunc func(step string, str string)
type FmtErrFunc func(step string, err error)

func NewLog() *log {
	log := new(log)
	log.MqData.Data.Define = make(map[string]interface{}, 0)
	return log
}

func (log *log) SetProject(project string) *log {
	log.MqData.Project = project
	return log
}

func (log *log) SetChannel(channel string) *log {
	log.MqData.Data.Channel = channel
	return log
}

func (log *log) InfoByMap(infoMap map[string]interface{}) *log {
	infoJson, _ := json.Marshal(&infoMap)
	log.MqData.Data.Msg = string(infoJson)
	log.MqData.Data.Level = LevelInfo
	return log
}

func (log *log) Info(info string) *log {
	log.MqData.Data.Msg = info
	log.MqData.Data.Level = LevelInfo
	return log
}

func (log *log) ErrorMap(errMap map[string]interface{}) *log {
	errJson, _ := json.Marshal(&errMap)
	log.MqData.Data.Msg = string(errJson)
	log.MqData.Data.Level = LevelError
	return log
}

func (log *log) Error(errorStr string) *log {
	log.MqData.Data.Msg = errorStr
	log.MqData.Data.Level = LevelError
	return log
}

func (log *log) AddDefine(defineKey, defineValue string) *log {
	log.MqData.Data.Define[defineKey] = defineValue
	return log
}

func (log *log) AddArrDefine(defineArrKey string, defineArrValue []string) *log {
	log.MqData.Data.Define[defineArrKey] = defineArrValue
	return log
}

func (log *log) SetServerName(serverName string) *log {
	log.ServerName = serverName
	return log
}

func (log *log) SetEnvironment(environment string) *log {
	log.Environment = environment
	return log
}

func (log *log) SetMqAddress(address string) *log {
	log.MaAddress = address
	return log
}

func (log *log) SetAlarm(alarm Alarm) *log {
	log.Alarm = alarm
	return log
}

func (log *log) SetLogger(logger Logger) *log {
	log.Logger = logger
	return log
}

func (log *log) alarmError(ctx context.Context, step string, err error) {
	if log.Alarm == nil {
		log2.Printf("【alarm】 step %s: %s \n", step, err.Error())
		return
	}
	log.Alarm.Error(ctx, step, err)
}

func (log *log) info(ctx context.Context, step string, str string) {
	if log.Logger == nil {
		log2.Printf("【info】 step %s: %s \n", step, str)
		return
	}
	log.Logger.Info(ctx, step, str)
}

func (log *log) error(ctx context.Context, step string, err error) {
	if log.Logger == nil {
		log2.Printf("【error】 step %s: %s \n", step, err.Error())
		return
	}
	log.Logger.Error(ctx, step, err)
}

func (log *log) Send() error {
	if log.ServerName == "" {
		return fmt.Errorf("ServerName is null")
	}
	if log.Environment == "" {
		return fmt.Errorf("environment is null")
	}
	if log.MqData.Project == "" {
		log.MqData.Project = log.ServerName + log.Environment
	}
	// 设置时间
	if log.MqData.Data.CreatedAt == 0 {
		log.MqData.Data.CreatedAt = time.Now().UnixNano() / 1e6
	}
	return LogMqPush(log)
}
