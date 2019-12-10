package wgconnect

import (
	"context"
	"sync"
)

type ConnectOpts struct {
	Interface struct {
		PrivateKey string
		Address    string
		DNS        string
	}
	Peer struct {
		PublicKey  string
		AllowedIPs []string
		Endpoint   string
	}
}

// Connect will start a userspace wireguard interface
// and then connect to a peer using the given connect opts
func Connect(ctx context.Context, opts ConnectOpts) (err error) {

	waitgroup := sync.WaitGroup{}

	waitgroup.Add(1)
	go func() {
		err = startWgIface(ctx)
		waitgroup.Done()
	}()

	waitgroup.Add(1)
	go func() {
		err = connectToWg(ctx, opts)
		waitgroup.Done()
	}()

	waitgroup.Wait()

	return err
}
