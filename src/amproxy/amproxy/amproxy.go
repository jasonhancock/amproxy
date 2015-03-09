package main

import (
    "fmt"
    "math"
    "net"
    "os"
    "strconv"
    "time"

    "amproxy/auth"
    "amproxy/envparse"
    "amproxy/message"
)

var cServerAddr *net.TCPAddr
var skew float64
var authMap map[string]string

func main() {
    var err error

    // Read config from the environment
    var bInterface = envparse.GetSettingStr("BIND_INTERFACE", "127.0.0.1")
    var bPort      = envparse.GetSettingInt("BIND_PORT", 2005)
    var cServer    = envparse.GetSettingStr("CARBON_SERVER", "localhost")
    var cPort      = envparse.GetSettingInt("CARBON_PORT", 2003)
    authMap        = auth.Parse(envparse.GetSettingStr("AUTH", ""))
    skew           = float64(envparse.GetSettingInt("SKEW", 300))

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
        msg := new(message.Message)
        e := msg.Decompose(string(buf[:n]))
        if e != nil {
            fmt.Printf("Error decomposing message %q - %s\n", string(buf[:n]), e.Error())
            return
        }

        key, ok := authMap[msg.Public_key]

        if !ok {
            fmt.Printf("key not defined for %s\n", msg.Public_key)
            return
        }

        sig := msg.ComputeSignature(key)

        if sig != msg.Signature {
            fmt.Printf("Computed signature %s doesn't match provided signature %s\n", sig, msg.Signature)
            return
        }

        delta := math.Abs(float64(time.Now().Unix() - int64(msg.Timestamp)))
        if delta > skew {
            fmt.Printf("delta = %.0f, max skew set to %.0f\n", delta, skew)
            return
        }

        fmt.Println(msg.Public_key)
        fmt.Println(sig)

        _, err = carbon_conn.Write([]byte(msg.MetricStr() + "\n"))
        if err != nil {
            println("Write to carbon server failed:", err.Error())
            return
        }

        // write the n bytes read
        _, err2 := conn.Write(buf[0:n])
        if err2 != nil {
            return
        }
    }
}
