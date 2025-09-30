package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/mventrelli/telemetry-relay/internal/observability"
	"github.com/mventrelli/telemetry-relay/internal/telemetry"
)

func main() {
	udpAddr := getEnv("UDP_ADDR", ":9000")
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	forwardURL := os.Getenv("FORWARD_URL") // optional
	workers := atoi(getEnv("WORKERS", "4"))
	queueSz := atoi(getEnv("QUEUE_SIZE", "1024"))

	// HTTP: /healthz + /metrics
	httpSrv := &http.Server{Addr: httpAddr, Handler: observability.Router()}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	// UDP listener
	conn, err := net.ListenPacket("udp", udpAddr)
	if err != nil {
		log.Fatalf("udp listen %s: %v", udpAddr, err)
	}
	defer conn.Close()
	log.Printf("udp listening on %s", udpAddr)

	// Forwarding worker pool
	type job struct{ pkt telemetry.Packet }
	jobs := make(chan job, queueSz)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if forwardURL == "" {
				for range jobs {
				} // drain if no forward target set
				return
			}
			client := &http.Client{Timeout: 3 * time.Second}
			for j := range jobs {
				start := time.Now()
				payload, _ := json.Marshal(j.pkt)
				req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, forwardURL, bytes.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					observability.ForwardErrs.Inc()
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					observability.Forwarded.Inc()
				} else {
					observability.ForwardErrs.Inc()
				}
				observability.ForwardSec.Observe(time.Since(start).Seconds())
			}
		}()
	}

	// UDP read loop
	buf := make([]byte, 64*1024)
readLoop:
	for {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := conn.ReadFrom(buf)
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			select {
			case <-ctx.Done():
				break readLoop
			default:
				continue
			}
		}
		if err != nil {
			log.Printf("udp read: %v", err)
			continue
		}
		observability.UDPBytes.Add(float64(n))
		observability.Ingested.Inc()

		var pkt telemetry.Packet
		if err := json.Unmarshal(buf[:n], &pkt); err != nil {
			observability.ParseErrors.Inc()
			continue
		}
		log.Printf("received packet seq=%d source=%s values=%v", pkt.Seq, pkt.Source, pkt.Values)
		select {
		case jobs <- job{pkt: pkt}:
		default:
			// drop when queue is full (UDP semantics)
		}
	}

	// Shutdown
	close(jobs)
	wg.Wait()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	log.Println("graceful shutdown complete")
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
