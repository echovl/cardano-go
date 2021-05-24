package cardano

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

func WithNode(node cardanoNode) Options {
	return optionFunc(func(client *Client) {
		client.node = node
	})
}
