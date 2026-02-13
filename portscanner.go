package main

import (
	"fmt"
	"log"

	gonet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	fmt.Println("Gathering detailed network service report...\n")

	connections, err := gonet.Connections("all")
	if err != nil {
		log.Fatal(err)
	}

	for _, conn := range connections {

		// Only show listening or established
		if conn.Status != "LISTEN" && conn.Status != "ESTABLISHED" {
			continue
		}

		var processName string
		if conn.Pid != 0 {
			proc, err := process.NewProcess(conn.Pid)
			if err == nil {
				name, err := proc.Name()
				if err == nil {
					processName = name
				}
			}
		}

		fmt.Println("===================================")
		fmt.Printf("Protocol     : %s\n", conn.Type)
		fmt.Printf("Status       : %s\n", conn.Status)
		fmt.Printf("Local Addr   : %s:%d\n", conn.Laddr.IP, conn.Laddr.Port)

		if conn.Raddr.IP != "" {
			fmt.Printf("Remote Addr  : %s:%d\n", conn.Raddr.IP, conn.Raddr.Port)
		}

		fmt.Printf("PID          : %d\n", conn.Pid)
		fmt.Printf("Process Name : %s\n", processName)
		fmt.Println("===================================")
	}

	fmt.Println("\nReport completed.")
}
