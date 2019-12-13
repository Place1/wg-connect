module github.com/place1/wg-connect

go 1.13

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/gosuri/uilive v0.0.3
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.8.1
	github.com/place1/wg-embed v0.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/vishvananda/netlink v1.0.0
	golang.zx2c4.com/wireguard v0.0.20191012
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191205174707-786493d6718c
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.51.0
)

replace github.com/place1/wg-embed => ../wg-embed
