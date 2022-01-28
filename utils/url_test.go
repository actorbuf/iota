package utils

import (
	"fmt"
	"testing"
)

func TestStruct2UrlValues(t *testing.T) {
	type A struct {
		AccountType int32  `protobuf:"varint,1,opt,name=accountType,proto3" json:"accountType,omitempty"` // 用户账号类型 2：微信开放账号
		Uid         string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`                  // 微信用户则填入对应的 openid/unionid
		UserIp      string `protobuf:"bytes,3,opt,name=userIp,proto3" json:"userIp,omitempty"`            // 用户领取奖励时的真实外网 IP
		PostTime    int64  `protobuf:"varint,4,opt,name=postTime,proto3" json:"postTime,omitempty"`       // 用户操作时间戳，单位秒（格林威治时间精确到秒，如1501590972）。
		WxSubType   int32  `protobuf:"varint,5,opt,name=wxSubType,proto3" json:"wxSubType,omitempty"`     // 1：微信公众号 2：微信小程序。
		WxToken     string `protobuf:"bytes,6,opt,name=wxToken,proto3" json:"wxToken,omitempty"`          // wxSubType = 1：微信公众号或第三方登录，则为授权的 access_token
		AppId       string `protobuf:"bytes,7,opt,name=appId,proto3" json:"appId,omitempty"`              // 天御appId
		RandNum     string `protobuf:"bytes,8,opt,name=randNum,proto3" json:"randNum,omitempty"`          // 随机数
	}

	var a = A{}
	u := Struct2UrlValues(a)
	fmt.Println(u.Encode())
}

func TestStruct2UrlValuesOmitEmpty(t *testing.T) {
	type A struct {
		AccountType int32  `protobuf:"varint,1,opt,name=accountType,proto3" json:"accountType,omitempty"` // 用户账号类型 2：微信开放账号
		Uid         string `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`                  // 微信用户则填入对应的 openid/unionid
		UserIp      string `protobuf:"bytes,3,opt,name=userIp,proto3" json:"userIp,omitempty"`            // 用户领取奖励时的真实外网 IP
		PostTime    int64  `protobuf:"varint,4,opt,name=postTime,proto3" json:"postTime,omitempty"`       // 用户操作时间戳，单位秒（格林威治时间精确到秒，如1501590972）。
		WxSubType   int32  `protobuf:"varint,5,opt,name=wxSubType,proto3" json:"wxSubType,omitempty"`     // 1：微信公众号 2：微信小程序。
		WxToken     string `protobuf:"bytes,6,opt,name=wxToken,proto3" json:"wxToken,omitempty"`          // wxSubType = 1：微信公众号或第三方登录，则为授权的 access_token
		AppId       string `protobuf:"bytes,7,opt,name=appId,proto3" json:"appId,omitempty"`              // 天御appId
		RandNum     string `protobuf:"bytes,8,opt,name=randNum,proto3" json:"randNum,omitempty"`          // 随机数
	}

	var a = A{}
	u := Struct2UrlValuesOmitEmpty(a)
	fmt.Println(u.Encode())
}
