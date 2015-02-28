package main

import (
    "flag"
    "fmt"
    "net"
    "os"
    "strconv"

    "amproxy/msgparser"
)

var fInterface = flag.String("interface", "127.0.0.1", "Interface to listen on.")
var fPort = flag.Int("port", 2003, "Port to listen on.")

func main() {
    flag.Parse()

    // Listen for incoming connections.
    l, err := net.Listen("tcp", *fInterface + ":" + strconv.Itoa(*fPort))
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes.
    defer l.Close()
    fmt.Println("Listening on " + *fInterface + ":" + strconv.Itoa(*fPort))
    for {
        // Listen for an incoming connection.
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
        // Handle connections in a new goroutine.
        go handleRequest(conn)
    }
}

func handleRequest(conn net.Conn) {
    defer conn.Close()

    fmt.Println("Connection from: ", conn.RemoteAddr())

    var buf [1024]byte
    for {
        // read upto 512 bytes
        n, err := conn.Read(buf[0:])
        if err != nil {
            return
        }

        fmt.Println(string(buf[:n]))
        msg, e := msgparser.Decompose(string(buf[:n]))
        if e != nil {
            fmt.Printf("Error decomposing message %q - %s", string(buf[:n]), e.Error())
            return
        }

        fmt.Println(msg.Public_key)

        // write the n bytes read
        _, err2 := conn.Write(buf[0:n])
        if err2 != nil {
            return
        }
    }
}
