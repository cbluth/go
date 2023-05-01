package ping

import (
	"flag"
	"fmt"
	"time"

	"github.com/cbluth/go/pkg/ping"
)

func ExecuteCommand() error {
	args := struct {
		ipv6    bool
		help    bool
		sleep   time.Duration
		timeout time.Duration
	}{}
	flag.DurationVar(&args.sleep, "s", 1*time.Second, "sleep")
	flag.DurationVar(&args.timeout, "t", 1*time.Second, "timeout")
	flag.BoolVar(&args.help, "h", false, "help")
	flag.BoolVar(&args.ipv6, "6", false, "ipv6")
	flag.Parse()
	flag.Usage = func() {
		fmt.Println("ping [OPTION] [host]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if args.help {
		flag.Usage()
		return nil
	}
	proto := "ip4"
	host := flag.Args()[0]
	if args.ipv6 {
		proto = "ip6"
	}
	return ping.Ping(proto, host, args.sleep, args.timeout)
}
