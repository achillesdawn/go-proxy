package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/fatih/color"
)

//lint:ignore U1000 Ignore unused function temporarily for debugging
func debugRequest(buf *bufio.Reader) *bufio.Reader {

	result := make([]byte, 0, 1000)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		result = append(result, line...)

		if slices.Compare(line, []byte{'\r', '\n'}) == 0 {
			break
		}
	}

	fmt.Printf("RECEIVED:\n%s", string(result))

	return bufio.NewReader(bytes.NewReader(result))
}

func handleGetRequest(conn net.Conn, r *http.Request) {

	if r.RequestURI == "/" {
		r.RequestURI = "http://www.google.com"
	}

	newRequest, err := http.NewRequest(r.Method, r.RequestURI, r.Body)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(newRequest)
	if err != nil {
		panic(err)
	}

	res.Write(conn)

}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func readData(buf bufio.Reader) {

	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				color.Red("Conn closed")
				break
			} else {
				panic(err)
			}
		}
		fmt.Println("┋", string(line))
	}
}

func handleConnect(conn net.Conn, r *http.Request) {

	color.Blue("DIALING %s", r.Host)

	upstream, err := net.Dial("tcp", r.Host)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
	}

	okResponse := "HTTP/1.1 200 OK\r\n\r\n"
	conn.Write([]byte(okResponse))

	defer upstream.Close()
	defer conn.Close()

	go io.Copy(conn, upstream)
	io.Copy(upstream, conn)
	// go func() {
	// 	buf := bufio.NewReader(upstream)

	// 	var buffer bytes.Buffer

	// 	for {
	// 		n, err := buf.Read(buffer.AvailableBuffer())
	// 		if err != nil {
	// 			color.Red("Error reading from %s : %s", addr, err.Error())
	// 			return
	// 		}

	// 		color.Blue("[%s]: %s\n\n", addr, buffer.String())
	// 		if n == 0 {
	// 			break
	// 		}
	// 	}

	// }()

}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func printHeaders(req *http.Request) {
	for key, value := range req.Header {
		values := strings.Join(value, ";")
		fmt.Printf("├─ %s : %s\n", color.BlueString(key), values)
	}
	fmt.Println("└───────────")
}

func handleConn(conn net.Conn) {

	defer conn.Close()

	fmt.Printf("[%s]\n", conn.RemoteAddr().String())

	buf := bufio.NewReader(conn)

	// byteRequest := debugRequest(buf)

	req, err := http.ReadRequest(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			color.Red("client closed")
			return
		}
	}

	switch req.Method {
	case http.MethodConnect:
		color.Cyan("CONNECT: %s", req.RequestURI)
		handleConnect(conn, req)
		// color.Yellow("CONNECT Proxied")

	case http.MethodGet:
		fmt.Printf("%s %s\n", color.GreenString("GET"), req.RequestURI)
		handleGetRequest(conn, req)
	}

}

func main() {
	listener, err := net.Listen("tcp", ":9988")
	if err != nil {
		panic(err)
	}

	color.Yellow("Listening on port %d", 9988)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go handleConn(conn)
	}
}
