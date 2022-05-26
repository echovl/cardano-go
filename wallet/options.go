package wallet

import "github.com/echovl/cardano-go/node"

type Options interface {
	apply(*Client)
}

type optionFunc func(*Client)

func (f optionFunc) apply(client *Client) {
	f(client)
}

func WithDB(db DB) Options {
	return optionFunc(func(client *Client) {
		client.db = db
	})
}

func WithSocket(socketPath string) Options {
	return optionFunc(func(client *Client) {
		client.socketPath = socketPath
	})
}

func WithNode(node node.Node) Options {
	return optionFunc(func(client *Client) {
		client.node = node
	})
}
