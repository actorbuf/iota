package etcd

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	client3 "go.etcd.io/etcd/client/v3"
)

var client *Client

type Client struct {
	client *client3.Client
}

type ClientStarter struct {
	Servers []string
}

func (e *ClientStarter) Init() error {
	var err error
	client, err = NewEtcdClient(e.Servers)
	if err != nil {
		return err
	}
	return nil
}

func GetClient() *Client {
	return client
}

func NewEtcdClient(servers []string) (*Client, error) {
	cli, err := client3.New(client3.Config{
		Endpoints:          servers,
		DialTimeout:        timeOut,
		MaxCallSendMsgSize: 11 * (1 << 20),
	})
	if err != nil {
		return nil, err
	}
	var std = &Client{client: cli}
	return std, nil
}

func (c *Client) GetClient() *client3.Client {
	return c.client
}

func (c *Client) Put(ctx context.Context, key, value string) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	_, err := c.client.Put(ctx, key, value)
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	rsp, err := c.client.Get(ctx, key)
	cancel()
	if err != nil {
		return nil, err
	}

	return rsp.Kvs, nil
}

func (c *Client) Del(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	_, err := c.client.Delete(ctx, key)
	cancel()
	if err != nil {
		return err
	}

	return nil
}
