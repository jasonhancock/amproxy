package amproxy

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type counters struct {
	Connections    uint64 `json:"connections"`
	BadSkew        uint64 `json:"badskew"`
	BadSig         uint64 `json:"badsig"`
	BadKeyundef    uint64 `json:"badkeyundef"`
	BadDecompose   uint64 `json:"baddecompose"`
	BadMetric      uint64 `json:"badmetric"`
	BadCarbonconn  uint64 `json:"badcarbonconn"`
	BadCarbonwrite uint64 `json:"badcarbonwrite"`
	GoodMetric     uint64 `json:"goodmetric"`
}

var Counters counters

func ShipMetrics(cServerAddr *net.TCPAddr, c *counters) {
	ticker := time.NewTicker(time.Second * 60)
	hostname, _ := os.Hostname()
	pre := reverseDelimitedString(hostname, ".") + ".amproxy"
	format := "%s.%s %d %d\n"
	for range ticker.C {
		b, _ := json.Marshal(c)
		fmt.Printf("%s\n", b)

		// connect to carbon server
		carbon_conn, err := net.DialTCP("tcp", nil, cServerAddr)
		if err != nil {
			println("Connection to carbon server failed:", err.Error())
			continue
		}
		ts := time.Now().Unix()

		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "connections", c.Connections, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "badskew", c.BadSkew, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "badsig", c.BadSig, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "badkeyundef", c.BadKeyundef, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "baddecompose", c.BadDecompose, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "badmetric", c.BadMetric, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "badcarbonwrite", c.BadCarbonwrite, ts))
		writeMetric(carbon_conn, fmt.Sprintf(format, pre, "goodmetric", c.GoodMetric, ts))

		carbon_conn.Close()
	}
}

func writeMetric(conn *net.TCPConn, str string) {
	_, err := conn.Write([]byte(str))
	if err != nil {
		println("Write to carbon server failed:", err.Error())
	}
}

func reverseDelimitedString(str, delimiter string) string {
	pieces := strings.Split(str, delimiter)

	var rev []string

	for i := len(pieces) - 1; i >= 0; i-- {
		rev = append(rev, pieces[i])
	}

	return strings.Join(rev, delimiter)
}

func Metrics_http_handler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(Counters)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
