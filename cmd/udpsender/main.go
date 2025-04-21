package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {

	address, err := net.ResolveUDPAddr("udp", ":42069")

	if err != nil {
		log.Fatal(err.Error())
	}

	conn, conn_err := net.DialUDP("udp", nil, address)

	if conn_err != nil {
		log.Fatal(conn_err.Error())
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {

		fmt.Print(">")

		input, err := reader.ReadString('\n')

		if err != nil {
			log.Fatal(err.Error())
		}

		conn.Write([]byte(input))

	}

}
