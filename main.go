package main

import (
	"fmt"
	"http_server/internal/request"
	"log"
	"net"
)

// func getLinesFromChannel(f io.ReadCloser) <-chan string {
// 	out := make(chan string, 1)

// 	go func() {
// 		defer f.Close()
// 		defer close(out)

// 		str := ""

// 		for {
// 			data := make([]byte, 10)
// 			n, err := f.Read(data)
// 			if err != nil {
// 				break
// 			}
// 			data = data[:n]
// 			if i := bytes.IndexByte(data, '\n'); i != -1 {
// 				str += string(data[:i])
// 				out <- str
// 				data = data[i+1:]
// 				str = ""
// 			}

// 			str += string(data)
// 		}

// 		if str != "" {
// 			out <- str
// 		}
// 	}()

// 	return out
// }

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
		// lines := getLinesFromChannel(conn)
		// for line := range lines {
		// 	fmt.Println(line)
		// }

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error : ", err)
		}

		fmt.Println("Request Line:")
		fmt.Println("- METHOD:", r.RequestLine.Method)
		fmt.Println("- PATH:", r.RequestLine.RequestTarget)
		fmt.Println("- HTTP VERSION:", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")

		r.Headers.ForEach(func(k, v string) {
			fmt.Printf("- %s: %s\n", k, v)
		})

		fmt.Println("Body:")
		fmt.Printf("%s\n", r.Body)

	}

}
