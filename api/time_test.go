package api

import (
	"encoding/xml"
	"testing"
	"time"

	"gotest.tools/assert"
)

type Sample struct {
	XMLName xml.Name   `xml:"sample"`
	Date    *Timestamp `xml:"time"`
}

func TestTime(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	date := &Sample{Date: &Timestamp{Time: time.Date(1974, 11, 25, 12, 0, 0, 0, location)}}
	bytes, err := xml.Marshal(date)
	assert.NilError(t, err)
	xmlstr := string(bytes)
	assert.Equal(t, "<sample><time>1974-11-25T12:00:00.000+00:00</time></sample>", xmlstr)
}
