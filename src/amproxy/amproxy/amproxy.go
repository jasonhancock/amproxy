package main

import (
    "fmt"
    "math"
    "net"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync/atomic"
    "time"

    "amproxy/envparse"
    "amproxy/message"
)

var cServerAddr *net.TCPAddr
var skew float64
var authMap map[string]Creds
var authMapLoadTime time.Time
var c counters

func main() {
    var err error

    c.init()

    // Read config from the environment
    var bInterface = envparse.GetSettingStr("BIND_INTERFACE", "127.0.0.1")
    var bPort      = envparse.GetSettingInt("BIND_PORT", 2005)
    var cServer    = envparse.GetSettingStr("CARBON_SERVER", "localhost")
    var cPort      = envparse.GetSettingInt("CARBON_PORT", 2003)
    var authFile   = envparse.GetSettingStr("AUTH_FILE", "")
    skew           = float64(envparse.GetSettingInt("SKEW", 300))

    if(authFile == "") {
        println("No auth file passed")
        os.Exit(1)
    }

    authMap, authMapLoadTime = loadUserConfigFile(authFile)

    cServerAddr, err = net.ResolveTCPAddr("tcp", cServer + ":" + strconv.Itoa(cPort))
    if err != nil {
        println("Unable to resolve carbon server: ", err.Error())
        os.Exit(1)
    }
    fmt.Printf("Carbon server: %s:%d\n", cServer, cPort)

    // Set up the metrics http server
    go func() {
        http.HandleFunc("/", metrics_http_handler)
        http.ListenAndServe(":8080", nil)
    }()

    go shipMetrics(cServerAddr, &c)

    go reloadAuth(authFile)

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
        atomic.AddUint64(&c.Connections, 1)
        go handleRequest(conn)
    }
}

func reloadAuth(authFile string) {
    ticker := time.NewTicker(time.Second * 60)
    for _ = range ticker.C {
        info, err := os.Stat(authFile)
        if err != nil {
             fmt.Println("Error stating authFile:", err.Error())
             continue
        }

        ts := info.ModTime()
        if ts != authMapLoadTime {
            fmt.Println("Reloading auth file configuration")
            authMap, authMapLoadTime = loadUserConfigFile(authFile)
        }
    }
}

func handleRequest(conn net.Conn) {
    defer conn.Close()

    fmt.Println("Connection from: ", conn.RemoteAddr())

    // connect to carbon server
    carbon_conn, err := net.DialTCP("tcp", nil, cServerAddr)
    if err != nil {
        atomic.AddUint64(&c.BadCarbonconn, 1)
        println("Connection to carbon server failed:", err.Error())
        return
    }
    defer carbon_conn.Close()

    var buf [1024]byte
    buffer := ""
    for {
        n, err := conn.Read(buf[0:])
        if err != nil {
            return
        }
        buffer = buffer + string(buf[:n])

        // If the buffer ends in a newline, process the metrics
        if string(buf[n - 1]) == "\n" {
            lines := strings.Split(buffer, "\n")

            for i := 0; i < len(lines) ; i++ {
                if len(strings.TrimSpace(lines[i])) > 0 {
                    processMessage(carbon_conn, lines[i])
                }
            }

            buffer = ""
        }
    }
}

func processMessage(conn *net.TCPConn, line string) {

    msg := new(message.Message)
    e := msg.Decompose(line)
    if e != nil {
        atomic.AddUint64(&c.BadDecompose, 1)
        fmt.Printf("Error decomposing message %q - %s\n", line, e.Error())
        return
    }

    creds, ok := authMap[msg.Public_key]

    if !ok {
        atomic.AddUint64(&c.BadKeyundef, 1)
        fmt.Printf("key not defined for %s\n", msg.Public_key)
        return
    }

    sig := msg.ComputeSignature(creds.SecretKey)

    if sig != msg.Signature {
        atomic.AddUint64(&c.BadSig, 1)
        fmt.Printf("Computed signature %s doesn't match provided signature %s\n", sig, msg.Signature)
        return
    }

    delta := math.Abs(float64(time.Now().Unix() - int64(msg.Timestamp)))
    if delta > skew {
        atomic.AddUint64(&c.BadSkew, 1)
        fmt.Printf("delta = %.0f, max skew set to %.0f\n", delta, skew)
        return
    }

    // validate the metric is on the approved list
    _, ok = creds.Metrics[msg.Name]
    if !ok {
        atomic.AddUint64(&c.BadMetric, 1)
        fmt.Printf("not an approved metric: %s\n", msg.Name)
        return
    }

    _, err := conn.Write([]byte(msg.MetricStr() + "\n"))
    if err != nil {
        atomic.AddUint64(&c.BadCarbonwrite, 1)
        println("Write to carbon server failed:", err.Error())
        return
    }
    atomic.AddUint64(&c.GoodMetric, 1)

    // write the n bytes read
    _, err2 := conn.Write([]byte(line))
    if err2 != nil {
        return
    }
}
