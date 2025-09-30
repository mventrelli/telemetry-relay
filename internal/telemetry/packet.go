package telemetry

import "time"

// Packet is the normalized structure we accept over UDP (JSON).
type Packet struct {
    Timestamp time.Time          `json:"timestamp"`            // RFC3339/RFC3339Nano
    Source    string             `json:"source"`
    Signal    string             `json:"signal_type,omitempty"`
    Seq       uint64             `json:"seq"`
    CRCOk     bool               `json:"crc_ok"`
    Values    map[string]float64 `json:"values"`
}
