package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/place1/wg-connect/pkg/wgconnect"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
)

var (
	app = kingpin.New("wg-connect", "a cli app to connect to a wireguard VPN server using a userspace wireguard implementation")

	config = app.Arg("config", "a wireguard configuration file").Required().File()

	debug = app.Flag("debug", "enable verbose logging").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	configFile, err := ini.Load((*config).Name())
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed to read wireguard config file"))
	}

	opts := wgconnect.ConnectOpts{}
	if err := configFile.StrictMapTo(&opts); err != nil {
		logrus.Fatal(errors.Wrap(err, "failed to parse config"))
	}

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, os.Interrupt)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		if err := wgconnect.Connect(ctx, opts); err != nil {
			logrus.Fatal(errors.Wrap(err, "wireguard client crash"))
		}
		wg.Done()
	}()

	select {
	case <-term:
	}

	logrus.Debug("shutting down...")

	cancel()

	c := make(chan struct{})
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()

	select {
	case <-c:
	case <-time.After(1 * time.Second):
		logrus.Error("shutdown timeout exceeded. exiting.")
	}
}
