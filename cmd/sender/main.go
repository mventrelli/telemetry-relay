package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	addr := getEnv("UDP_ADDR", "127.0.0.1:9000")
	conn, err := net.Dial("udp", addr)
	if err != nil { log.Fatal(err) }
	defer conn.Close()

	rand.Seed(time.Now().UnixNano())
	seq := uint64(0)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		seq++
		msg := map[string]any{
			"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
			"source":      "stage1",
			"signal_type": "thermal",
			"seq":         seq,
			"crc_ok":      true,
			"values": map[string]float64{
				"tank_temp_c": 10 + rand.Float64()*5,
				"pump_rpm":    5000 + rand.Float64()*200,
			},
		}
		b, _ := json.Marshal(msg)
		_, _ = conn.Write(b)
	}
}

func getEnv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
