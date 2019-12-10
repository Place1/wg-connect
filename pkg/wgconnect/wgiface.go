// +build !windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2019 WireGuard LLC. All Rights Reserved.
 */

// modified from https://git.zx2c4.com/wireguard-go

package wgconnect

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

func startWgIface(ctx context.Context) error {

	interfaceName := "wg0"

	tun, err := tun.CreateTUN("wg0", device.DefaultMTU)
	if err != nil {
		return errors.Wrap(err, "failed to create TUN device")
	}

	// open UAPI file (or use supplied fd)
	fileUAPI, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		return errors.Wrap(err, "UAPI listen error")
	}

	device := device.NewDevice(tun, device.NewLogger(device.LogLevelError, "wg0"))
	logrus.Debug("wg0 interface started")

	errs := make(chan error)

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		return errors.Wrap(err, "failed to listen on uapi socket")
	}

	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go device.IpcHandle(conn)
		}
	}()

	logrus.Debug("UAPI listener started")

	select {
	case <-ctx.Done():
	case <-device.Wait():
	}

	// clean up
	uapi.Close()
	device.Close()

	logrus.Debug("shutting down wg0")

	return nil
}
