package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	timeout = 5
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	args := os.Args
	argc := len(args)

	if argc == 1 {
		help()
		os.Exit(1)
	}

	switch args[1] {
	case "-server":
		if argc < 4 {
			fmt.Println(`-server needs two arguments, like "-server 3389 3390"`)
			os.Exit(1)
		}
		port1 := args[2]
		port2 := args[3]
		if isPort(port1) == false || isPort(port2) == false {
			fmt.Println("port should be a number and the range is [1,65535]")
			os.Exit(1)
		}
		log.Println("[√]", "start to listen port:", port1, "and port:", port2)
		port2port(port1, port2)
	case "-client":
		if argc < 4 {
			fmt.Println(`-client needs two arguments, like "-client 3389 8.8.8.8:3390"`)
			os.Exit(1)
		}
		port := args[2]
		address := args[3]
		if isPort(port) == false {
			fmt.Println("port should be a number and the range is [1,65535]")
			os.Exit(1)
		}
		if isDialString(address) == false {
			fmt.Println("address should be a string like [domain|ip:port]")
			os.Exit(1)
		}
		log.Println("[√]", "start to connect address: 127.0.0.1:", port, "and address:", address)
		host2host("127.0.0.1:"+port, address)
	default:
		help()
	}
}

func help() {
	fmt.Println("Usage:")
	fmt.Println("  -server port1 port2")
	fmt.Println("        forward data between port1 and port2 on server")
	fmt.Println("  -client port1 ip:port2")
	fmt.Println("        forward data between 127.0.0.1:port1 and ip:port2")
}

func isDNSName(str string) bool {
	r := regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		return false
	}
	return !isIP(str) && r.MatchString(str)

}

func isIP(str string) bool {
	return net.ParseIP(str) != nil
}

func isPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

func isDialString(str string) bool {
	if h, p, err := net.SplitHostPort(str); err == nil && h != "" && p != "" && (isDNSName(h) || isIP(h)) && isPort(p) {
		return true
	}

	return false
}

func port2port(port1, port2 string) {
	listener1 := createServer("0.0.0.0:" + port1)
	listener2 := createServer("0.0.0.0:" + port2)
	log.Println("[√]", "listen port:", port1, "and", port2, "successfully, wait for client")
	for {
		conn1 := accept(listener1)
		conn2 := accept(listener2)
		if conn1 == nil || conn2 == nil {
			log.Println("[x]", "accept client faild, retry in ", timeout, " seconds")
			time.Sleep(timeout * time.Second)
			continue
		}
		forward(conn1, conn2)
	}
}

func host2host(address1, address2 string) {
	for {
		log.Println("[+]", "try to connect host:["+address1+"] and ["+address2+"]")
		var host1, host2 net.Conn
		var err error
		for {
			host1, err = net.Dial("tcp", address1)
			if err == nil {
				log.Println("[→]", "connect ["+address1+"] success.")
				break
			} else {
				log.Println("[x]", "connect target address ["+address1+"] faild. retry in ", timeout, " seconds")
				time.Sleep(timeout * time.Second)
			}
		}
		for {
			host2, err = net.Dial("tcp", address2)
			if err == nil {
				log.Println("[→]", "connect ["+address2+"] success")
				break
			} else {
				log.Println("[x]", "connect target address ["+address2+"] faild. retry in ", timeout, " seconds")
				time.Sleep(timeout * time.Second)
			}
		}
		forward(host1, host2)
	}
}

func createServer(address string) net.Listener {
	log.Println("[+]", "try to start server on:["+address+"]")
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("[x]", "listen address ["+address+"] faild")
	}
	log.Println("[√]", "start listen at address:["+address+"]")
	return server
}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalln("[x]", "accept connect ["+conn.RemoteAddr().String()+"] faild ", err.Error())
		return nil
	}
	log.Println("[√]", "accept a new client. remote address:["+conn.RemoteAddr().String()+"], local address:["+conn.LocalAddr().String()+"]")
	return conn
}

func forward(conn1, conn2 net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go copyConn(conn1, conn2, &wg)
	go copyConn(conn2, conn1, &wg)
	wg.Wait()
}

func copyConn(conn1, conn2 net.Conn, wg *sync.WaitGroup) {
	io.Copy(conn1, conn2)
	conn1.Close()
	log.Println("[←]", "close the connect at local:["+conn1.LocalAddr().String()+"] and remote:["+conn1.RemoteAddr().String()+"]")
	wg.Done()
}
