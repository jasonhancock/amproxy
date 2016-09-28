package main

import (
	"fmt"
	"github.com/spf13/viper"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jasonhancock/amproxy"
)

var cServerAddr *net.TCPAddr
var skew float64
var authMap map[string]amproxy.Creds
var authMapLoadTime time.Time

func main() {
	var err error

	viper.AutomaticEnv()
	viper.SetDefault("bind_interface", "127.0.0.1")
	viper.SetDefault("port", 2005)
	viper.SetDefault("carbon_server", "localhost")
	viper.SetDefault("carbon_port", "2003")
	viper.SetDefault("auth_file", "")
	viper.SetDefault("skew", 300)

	// Read config from the environment
	var authFile = viper.GetString("auth_file")
	skew = viper.GetFloat64("skew")

	if authFile == "" {
		println("No auth file passed")
		os.Exit(1)
	}

	authMap, authMapLoadTime = amproxy.LoadUserConfigFile(authFile)

	carbon_server := viper.GetString("carbon_server") + ":" + strconv.Itoa(viper.GetInt("carbon_port"))
	cServerAddr, err = net.ResolveTCPAddr("tcp", carbon_server)
	if err != nil {
		println("Unable to resolve carbon server: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Carbon server: " + carbon_server)

	// Set up the metrics http server
	go func() {
		http.HandleFunc("/", amproxy.Metrics_http_handler)
		http.ListenAndServe(":8080", nil)
	}()

	go amproxy.ShipMetrics(cServerAddr, &amproxy.Counters)

	go reloadAuth(authFile)

	// Listen for incoming connections.
	listen_string := viper.GetString("bind_interface") + ":" + strconv.Itoa(viper.GetInt("port"))
	l, err := net.Listen("tcp", listen_string)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + listen_string)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		atomic.AddUint64(&amproxy.Counters.Connections, 1)
		go handleRequest(conn)
	}
}

func reloadAuth(authFile string) {
	ticker := time.NewTicker(time.Second * 60)
	for range ticker.C {
		info, err := os.Stat(authFile)
		if err != nil {
			fmt.Println("Error stating authFile:", err.Error())
			continue
		}

		ts := info.ModTime()
		if ts != authMapLoadTime {
			fmt.Println("Reloading auth file configuration")
			authMap, authMapLoadTime = amproxy.LoadUserConfigFile(authFile)
		}
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Connection from: ", conn.RemoteAddr())

	// connect to carbon server
	carbon_conn, err := net.DialTCP("tcp", nil, cServerAddr)
	if err != nil {
		atomic.AddUint64(&amproxy.Counters.BadCarbonconn, 1)
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
		if string(buf[n-1]) == "\n" {
			lines := strings.Split(buffer, "\n")

			for i := 0; i < len(lines); i++ {
				if len(strings.TrimSpace(lines[i])) > 0 {
					processMessage(carbon_conn, lines[i])
				}
			}

			buffer = ""
		}
	}
}

func processMessage(conn *net.TCPConn, line string) {

	msg := new(amproxy.Message)
	e := msg.Decompose(line)
	if e != nil {
		atomic.AddUint64(&amproxy.Counters.BadDecompose, 1)
		fmt.Printf("Error decomposing message %q - %s\n", line, e.Error())
		return
	}

	creds, ok := authMap[msg.Public_key]

	if !ok {
		atomic.AddUint64(&amproxy.Counters.BadKeyundef, 1)
		fmt.Printf("key not defined for %s\n", msg.Public_key)
		return
	}

	sig := msg.ComputeSignature(creds.SecretKey)

	if sig != msg.Signature {
		atomic.AddUint64(&amproxy.Counters.BadSig, 1)
		fmt.Printf("Computed signature %s doesn't match provided signature %s\n", sig, msg.Signature)
		return
	}

	delta := math.Abs(float64(time.Now().Unix() - int64(msg.Timestamp)))
	if delta > skew {
		atomic.AddUint64(&amproxy.Counters.BadSkew, 1)
		fmt.Printf("delta = %.0f, max skew set to %.0f\n", delta, skew)
		return
	}

	// validate the metric is on the approved list
	_, ok = creds.Metrics[msg.Name]
	if !ok {
		atomic.AddUint64(&amproxy.Counters.BadMetric, 1)
		fmt.Printf("not an approved metric: %s\n", msg.Name)
		return
	}

	_, err := conn.Write([]byte(msg.MetricStr() + "\n"))
	if err != nil {
		atomic.AddUint64(&amproxy.Counters.BadCarbonwrite, 1)
		println("Write to carbon server failed:", err.Error())
		return
	}
	atomic.AddUint64(&amproxy.Counters.GoodMetric, 1)
}
