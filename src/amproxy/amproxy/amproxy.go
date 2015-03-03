package main

import (
    "fmt"
    "net"
    "os"
    "strconv"

    "amproxy/msgparser"
)

var cServerAddr *net.TCPAddr

func main() {
    var err error

    // Read config from the environment
    var bInterface = getSettingStr("BIND_INTERFACE", "127.0.0.1")
    var bPort      = getSettingInt("BIND_PORT", 2005)
    var cServer    = getSettingStr("CARBON_SERVER", "localhost")
    var cPort      = getSettingInt("CARBON_PORT", 2003)

    cServerAddr, err = net.ResolveTCPAddr("tcp", cServer + ":" + strconv.Itoa(cPort))
    if err != nil {
        println("Unable to resolve carbon server: ", err.Error())
        os.Exit(1)
    }
    fmt.Printf("Carbon server: %s:%d\n", cServer, cPort)

    // Listen for incoming connections.
    l, err := net.Listen("tcp", bInterface + ":" + strconv.Itoa(bPort))
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    // Close the listener when the application closes.
    defer l.Close()
    fmt.Println("Listening on " + bInterface + ":" + strconv.Itoa(bPort))
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

func getSettingStr(key string, def string) string {
    val := os.Getenv(key)
    if val == "" {
        val = def
    }
    return val
}

func getSettingInt(key string, def int) int {
    var valInt int
    val := os.Getenv(key)
    if val == "" {
        valInt = def
    } else {
        valParsed, err := strconv.Atoi(val)
        if err != nil {
            fmt.Printf("Expecting an integer for key %s", key)
            os.Exit(1)
        }
        valInt = valParsed
    }
    return valInt
}

func handleRequest(conn net.Conn) {
    defer conn.Close()

    fmt.Println("Connection from: ", conn.RemoteAddr())

    // connect to carbon server
    carbon_conn, err := net.DialTCP("tcp", nil, cServerAddr)
    if err != nil {
        println("Connection to carbon server failed:", err.Error())
        return
    }
    defer carbon_conn.Close()

    var buf [1024]byte
    for {
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

        metricstr := msg.Name + " " + strconv.Itoa(msg.Value) + " " + strconv.Itoa(msg.Timestamp)

        _, err = carbon_conn.Write([]byte(metricstr + "\n"))
        //_, err = conn.Write([]byte(msg.Name + " " + strconv.Itoa(msg.Value) + " " + strconv.Itoa(msg.Timestamp)))
        if err != nil {
            println("Write to carbon server failed:", err.Error())
            return
        }
        fmt.Println("Wrote to server: ", metricstr)

        // write the n bytes read
        _, err2 := conn.Write(buf[0:n])
        if err2 != nil {
            return
        }
    }
}
