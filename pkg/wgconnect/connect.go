package wgconnect

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConnectOpts struct {
	Interface struct {
		PrivateKey string
		Address    string
		DNS        []string
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
		if err := startWgIface(ctx); err != nil {
			logrus.Error(errors.Wrap(err, "failed to start wireguard"))
		}
		waitgroup.Done()
	}()

	waitgroup.Add(1)
	go func() {
		if err := connectToWg(ctx, opts); err != nil {
			logrus.Error(errors.Wrap(err, "failed to connect to wireguard"))
		}
		waitgroup.Done()
	}()

	waitgroup.Wait()

	return err
}
