package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"

	"github.com/agalue/gominion/api"
	"github.com/gosnmp/gosnmp"
)

// GetOidToWalk builds a walkable OID
func GetOidToWalk(base string, instance string) string {
	var effectiveOid string
	if len(instance) > 0 {
		// Append the instance to the OID
		effectiveOid = base + instance
		// And remove the last byte
		effectiveOid = effectiveOid[:len(effectiveOid)-2]
	} else {
		// Use the OID "as-is"
		effectiveOid = base
	}
	return effectiveOid
}

// GetResultForPDU get results from a given SNMP PDU
func GetResultForPDU(pdu gosnmp.SnmpPDU, base string) api.SNMPResultDTO {
	var valueBytes []byte
	switch pdu.Type {
	case gosnmp.OctetString:
		valueBytes = []byte(pdu.Value.(string))
	case gosnmp.IPAddress:
		valueBytes = []byte(pdu.Value.(string))
	case gosnmp.ObjectIdentifier:
		valueBytes = getBytes(pdu.Value)
	default:
		valueBytes = BytesToJavaBigIntegerBytes(gosnmp.ToBigInt(pdu.Value).Bytes())
	}
	result := api.SNMPResultDTO{
		Base:     base,
		Instance: pdu.Name[len(base):],
		Value: api.SNMPValueDTO{
			Type:  int(pdu.Type),
			Value: base64.StdEncoding.EncodeToString(valueBytes),
		},
	}
	return result
}

func getBytes(key interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(key); err != nil {
		return make([]byte, 0)
	}
	return buf.Bytes()
}

// ToJavaBigIntegerBytes converts the value to a byte-array that can be used to initialize a java.math.BigInteger via the (byte[]) constructor.
// source: https://github.com/j-white/underling/blob/master/underlinglib/snmp_helper.go
func ToJavaBigIntegerBytes(value uint32) []byte {
	// Convert the integer to a byte-array
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, value)
	return BytesToJavaBigIntegerBytes(bytes)
}

// BytesToJavaBigIntegerBytes performs the opposite of toJavaBigIntegerBytes
// source: https://github.com/j-white/underling/blob/master/underlinglib/snmp_helper.go
func BytesToJavaBigIntegerBytes(valueBytes []byte) []byte {
	var bytes []byte
	// Find the first byte with a non-zero value, and trim the slice
	offset := 0
	for ; offset < len(valueBytes)-1; offset++ {
		if valueBytes[offset] != 0 {
			break
		}
	}

	if len(valueBytes) < 1 {
		bytes = []byte{byte(0)}
	} else {
		bytes = valueBytes[offset:]
	}

	// If the left-most bit of the first byte is 1, prepend another byte for the sign
	if bytes[0]>>7 == 1 {
		bytes = append([]byte{byte(0)}, bytes...)
	}

	return bytes
}
