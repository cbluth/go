package ping

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func Ping(proto, host string, sleep, timeout time.Duration) error {

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	ip, err := net.ResolveIPAddr(proto, host)
	if err != nil {
		return fmt.Errorf("ResolveIPAddr: %v", err)
	}
	c, err := icmp.ListenPacket(proto+":icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("ListenPacket: %v", err)
	}
	defer c.Close()
	seq := 0
	t := time.Tick(sleep)
	fmt.Printf("PING %s (%s)\n", host, ip.String())
	for {
		select {
		case <-s:
			return nil
		case <-t:
			m := icmp.Message{
				Code: 0,
				Body: &icmp.Echo{
					Seq:  seq,
					Data: []byte("ping"),
					ID:   os.Getpid() & 0xffff,
				},
				Type: ipv4.ICMPTypeEcho,
			}
			wb, err := m.Marshal(nil)
			if err != nil {
				return fmt.Errorf("Marshal: %v", err)
			}
			_, err = c.WriteTo(wb, &net.IPAddr{IP: ip.IP})
			if err != nil {
				return fmt.Errorf("WriteTo: %v", err)
			}
			c.SetReadDeadline(time.Now().Add(timeout))
			rb := make([]byte, 1500)
			n, _, err := c.ReadFrom(rb)
			if err != nil {
				continue
			} else {
				rm, err := icmp.ParseMessage(
					ipv4.ICMPTypeEcho.Protocol(),
					rb[:n],
				)
				if err == nil && rm.Type == ipv4.ICMPTypeEchoReply {
					log.Printf("%s :: PING\n", host)
				}
			}
		}
	}
	return nil
}
