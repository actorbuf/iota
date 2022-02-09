package rabbitmq

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type Alarm interface {
	SetMsg(map[string]string) error
	Do() error
	Validate() error
}

type DefaultAlarm struct {
	msg map[string]string
}

func (d *DefaultAlarm) SetMsg(m map[string]string) error {
	d.msg = m
	return nil
}

func (d *DefaultAlarm) Do() error {
	alarmMsg, err := jsoniter.Marshal(d.msg)
	if err != nil {
		return err
	}
	fmt.Println(string(alarmMsg))
	return nil
}

func (d *DefaultAlarm) Validate() error {
	return nil
}
