package shared

import "time"

type Message struct {
    Text      string    `json:"text"`
    Sender    string    `json:"sender"`
    Receiver  string    `json:"receiver"`
    Type      string    `json:"type"`
    Timestamp time.Time `json:"timestamp"`
}