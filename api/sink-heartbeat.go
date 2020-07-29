package api

import (
	"encoding/xml"
)

// MinionIdentityDTO represents the Minion Identity
type MinionIdentityDTO struct {
	XMLName   xml.Name   `xml:"minion"`
	ID        string     `xml:"id"`
	Location  string     `xml:"location"`
	Timestamp *Timestamp `xml:"timestamp"`
}
