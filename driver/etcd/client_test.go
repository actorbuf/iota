package etcd

import (
	"context"
	"testing"
)

func TestNewEtcdClient(t *testing.T) {
	var err error
	client, err = NewEtcdClient([]string{"10.0.0.94:32349"})
	if err != nil {
		panic(err)
	}
	var ctx = context.Background()
	//err = client.Put(ctx, "/test/hello", "hello1")
	//if err != nil {
	//	panic(err)
	//}
	//get, err := client.Get(ctx, "hello")
	//if err != nil {
	//	panic(err)
	//}
	//logrus.Infof("value: %+v", get[0].Value)
	client.Del(ctx, "/test/hello")
}
