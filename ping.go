package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <host>")
		return
	}

	host := os.Args[1]

	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		fmt.Println("Resolve error:", err)
		return
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		fmt.Println("Listen error (need sudo?):", err)
		return
	}
	defer conn.Close()

	fmt.Printf("PING %s (%s)\n\n", host, ipAddr.String())

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var sent, received int
	seq := 1
	id := os.Getpid() & 0xffff

loop:
	for {
		select {
		case <-sigChan:
			break loop
		default:
		}

		sent++

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  seq,
				Data: []byte("PING"),
			},
		}

		msgBytes, _ := msg.Marshal(nil)

		start := time.Now()
		_, err := conn.WriteTo(msgBytes, ipAddr)
		if err != nil {
			fmt.Println("Write error:", err)
			continue
		}

		reply := make([]byte, 1500)
		_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, peer, err := conn.ReadFrom(reply)
		if err == nil {
			duration := time.Since(start)

			rm, _ := icmp.ParseMessage(1, reply[:n])
			if rm.Type == ipv4.ICMPTypeEchoReply {
				received++
				fmt.Printf("%d bytes from %v: icmp_seq=%d time=%v\n",
					n, peer, seq, duration)
			}
		} else {
			fmt.Printf("Request timeout for icmp_seq %d\n", seq)
		}

		seq++
		time.Sleep(1 * time.Second)
	}

	// Print statistics
	fmt.Println("\n--- ping statistics ---")
	packetLoss := float64(sent-received) / float64(sent) * 100
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n",
		sent, received, packetLoss)
}
