package api

import (
	"encoding/xml"
	"time"
)

// TimeFormat the format to be used to marshal/unmarshal date from strings
const TimeFormat = "2006-01-02T15:04:05.000-07:00"

// Timestamp an object to seamlessly manage times in multiple formats
// Expected text format: 2006-01-02T15:04:05.000-07:00
type Timestamp struct {
	time.Time
}

// MarshalXML converts time object into time as string
func (t Timestamp) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if t.IsZero() {
		return e.EncodeElement("", start)
	}
	return e.EncodeElement(t.Format(TimeFormat), start)
}

// UnmarshalXML converts time string into time object
func (t *Timestamp) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	var err error
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	t.Time, err = time.Parse(TimeFormat, s)
	if err != nil {
		return err
	}
	return nil
}
