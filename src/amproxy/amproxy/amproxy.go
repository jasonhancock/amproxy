package main

import (
    "encoding/json"
    "fmt"
    "math"
    "net"
    "net/http"
    "os"
    "strconv"
    "sync/atomic"
    "time"

    "amproxy/envparse"
    "amproxy/message"
)

type counters struct {
    Connections uint64 `json:"connections"`
    BadSkew uint64 `json:"badskew"`
    BadSig uint64 `json:"badsig"`
    BadKeyundef uint64 `json:"badkeyundef"`
    BadDecompose uint64 `json:"baddecompose"`
    BadMetric uint64 `json:"badmetric"`
    BadCarbonconn uint64 `json:"badcarbonconn"`
    BadCarbonwrite uint64 `json:"badcarbonwrite"`
    GoodMetric uint64 `json:"goodmetric"`
}

var cServerAddr *net.TCPAddr
var skew float64
var authMap map[string]Creds
var c counters

func handler(w http.ResponseWriter, r *http.Request) {
    b, err := json.Marshal(c)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(b)
}

func main() {
    var err error

    c.Connections = 0
    c.BadSkew = 0
    c.BadSig = 0
    c.BadKeyundef = 0
    c.BadDecompose = 0
    c.BadMetric = 0
    c.BadCarbonwrite = 0
    c.BadCarbonconn = 0
    c.GoodMetric = 0

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

    authMap = loadUserConfigFile(authFile)

    cServerAddr, err = net.ResolveTCPAddr("tcp", cServer + ":" + strconv.Itoa(cPort))
    if err != nil {
        println("Unable to resolve carbon server: ", err.Error())
        os.Exit(1)
    }
    fmt.Printf("Carbon server: %s:%d\n", cServer, cPort)

    // Set up the metrics http server
    go func() {
        http.HandleFunc("/", handler)
        http.ListenAndServe(":8080", nil)
    }()

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
    for {
        n, err := conn.Read(buf[0:])
        if err != nil {
            return
        }

        fmt.Println(string(buf[:n]))
        msg := new(message.Message)
        e := msg.Decompose(string(buf[:n]))
        if e != nil {
            atomic.AddUint64(&c.BadDecompose, 1)
            fmt.Printf("Error decomposing message %q - %s\n", string(buf[:n]), e.Error())
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
            fmt.Printf("not an approved metric: %s", msg.Name)
            return
        }

        fmt.Println(msg.Public_key)
        fmt.Println(sig)

        _, err = carbon_conn.Write([]byte(msg.MetricStr() + "\n"))
        if err != nil {
            atomic.AddUint64(&c.BadCarbonwrite, 1)
            println("Write to carbon server failed:", err.Error())
            return
        }
        atomic.AddUint64(&c.GoodMetric, 1)

        // write the n bytes read
        _, err2 := conn.Write(buf[0:n])
        if err2 != nil {
            return
        }
    }
}
