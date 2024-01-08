package main

import (
	"log"
	"nat/pkg/nat"
)

func main() {
	client := nat.NewClient()
	err := client.DialUDP("192.168.110.18", 53771)
	if err != nil {
		log.Fatal("dail server failed")
	}

	client.HasNatProtection()
}
