package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesFromChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""

		for {
			data := make([]byte, 10)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				out <- str
				data = data[i+1:]
				str = ""
			}

			str += string(data)
		}

		if str != "" {
			out <- str
		}
	}()

	return out
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("error : ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error : ", err)
		}
		lines := getLinesFromChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
	}

}
