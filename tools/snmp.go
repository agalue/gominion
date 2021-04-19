package tools

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/log"
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
	log.Debugf("Processing PDU: %v", pdu)
	var valueBytes []byte
	switch pdu.Type {
	case gosnmp.ObjectIdentifier:
		fallthrough
	case gosnmp.IPAddress:
		if data, ok := pdu.Value.(string); ok {
			valueBytes = []byte(data)
		} else {
			log.Warnf("Cannot parse PDU %v", pdu)
		}
	case gosnmp.OctetString:
		if data, ok := pdu.Value.([]byte); ok {
			valueBytes = data
		} else {
			log.Warnf("Cannot parse PDU %v", pdu)
		}
	default:
		valueBytes = BytesToJavaBigIntegerBytes(gosnmp.ToBigInt(pdu.Value).Bytes())
	}
	return api.SNMPResultDTO{
		Base:     base,
		Instance: pdu.Name[len(base):],
		Value: api.SNMPValueDTO{
			Type:  int(pdu.Type),
			Value: base64.StdEncoding.EncodeToString(valueBytes),
		},
	}
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
