package dlis

import (
	"encoding/binary"
	"math"
)

// http://w3.energistics.org/rp66/v1/rp66v1_appb.html

var RepCode = []struct {
	// Code is index
	Name        string
	Size        int // # of bytes
	Descirption string
	Read        func([]byte) (interface{}, int)
}{
	{}, // 0 is not present
	{"FSHORT", 2, "Low precision floating point", nil}, // 1

	{"FSINGL", 4, "IEEE single precision floating point",
		func(in []byte) (interface{}, int) {
			return math.Float32frombits(binary.BigEndian.Uint32(in[:4])), 4
		}}, // 2

	{"FSING1", 8, "Validated single precision floating point", nil},          // 3
	{"FSING2", 12, "Two-way validated single precision floating point", nil}, // 4
	{"ISINGL", 4, "IBM single precision floating point", nil},                // 5
	{"VSINGL", 4, "VAX single precision floating point", nil},                // 6

	{"FDOUBL", 8, "IEEE double precision floating point",
		func(in []byte) (interface{}, int) {
			return math.Float64frombits(binary.BigEndian.Uint64(in[:4])), 8
		}}, // 7

	{"FDOUB1", 16, "Validated double precision floating point", nil},         // 8
	{"FDOUB2", 24, "Two-way validated double precision floating point", nil}, // 9
	{"CSINGL", 8, "Single precision complex", nil},                           // 10
	{"CDOUBL", 16, "Double precision complex", nil},                          // 11
	{"SSHORT", 1, "Short signed integer", nil},                               // 12
	{"SNORM", 2, "Normal signed integer", nil},                               // 13
	{"SLONG", 4, "Long signed integer", nil},                                 // 14

	{"USHORT", 1, "Short unsigned integer",
		func(in []byte) (interface{}, int) {
			return int(in[0]), 1
		}}, // 15

	{"UNORM", 2, "Normal unsigned integer",
		func(in []byte) (interface{}, int) {
			return int(binary.BigEndian.Uint16(in[:2])), 2
		}}, // 16

	{"ULONG", 4, "Long unsigned integer",
		func(in []byte) (interface{}, int) {
			return int(binary.BigEndian.Uint32(in[:4])), 4
		}}, // 17

	{"UVARI", 0, "Variable-length unsigned integer 1, 2, or 4",
		func(in []byte) (interface{}, int) {
			b1 := in[0]
			if checkBit(7, uint(b1)) { //
				if checkBit(6, uint(b1)) { // 4 bytes
					tmp := [4]byte{b1 & 0x3F} // first byte with mask 0011_1111
					copy(tmp[1:], in[1:4])    // remaining 3 bytes
					return int(binary.BigEndian.Uint32(tmp[:])), 4
				} else { // 2 bytes
					tmp := [2]byte{b1 & 0x3F} // first byte with mask 0011_1111
					tmp[1] = in[1]
					return int(binary.BigEndian.Uint16(tmp[:])), 2
				}
			}
			// single byte
			return int(b1), 1 // bit 7 is 0
		}}, // 18

	{"IDENT", 0, "Variable-length identifier",
		func(in []byte) (interface{}, int) {
			ln := in[0]
			if ln == 0 {
				return "", 0
			}
			// only allowed 33-96, 123-126
			// TODO check for allowed
			return string(in[1 : 1+ln]), int(1 + ln)
		}}, // 19

	{"ASCII", 0, "Variable-length ASCII character string",
		func(in []byte) (interface{}, int) {
			b1 := in[0]

			if b1 == 0 {
				return "", 1
			}

			var idlen, asciilen int
			if checkBit(7, uint(b1)) { //
				if checkBit(6, uint(b1)) { // 4 bytes
					tmp := [4]byte{b1 & 0x3F} // first byte with mask 0011_1111
					copy(tmp[1:], in[1:4])    // remaining 3 bytes
					idlen = 4
					asciilen = int(binary.BigEndian.Uint32(tmp[:]))
				} else { // 2 bytes
					tmp := [2]byte{b1 & 0x3F} // first byte with mask 0011_1111
					tmp[1] = in[1]
					idlen = 2
					asciilen = int(binary.BigEndian.Uint16(tmp[:]))
				}
			} else {
				// single byte
				idlen = 1
				asciilen = int(b1) // bit 7 is 0
			}

			return string(in[idlen : idlen+asciilen]), (idlen + asciilen)
		}}, // 20

	{"DTIME", 8, "Date and time", nil}, // 21

	{"ORIGIN", 0, "Origin reference",
		func(in []byte) (interface{}, int) {
			b1 := in[0]
			if checkBit(7, uint(b1)) { //
				if checkBit(6, uint(b1)) { // 4 bytes
					tmp := [4]byte{b1 & 0x3F} // first byte with mask 0011_1111
					copy(tmp[1:], in[1:4])    // remaining 3 bytes
					return int(binary.BigEndian.Uint32(tmp[:])), 4
				} else { // 2 bytes
					tmp := [2]byte{b1 & 0x3F} // first byte with mask 0011_1111
					tmp[1] = in[1]
					return int(binary.BigEndian.Uint16(tmp[:])), 2
				}
			}
			// single byte
			return int(b1), 1 // bit 7 is 0
		}}, // 22

	{"OBNAME", 0, "Object name",
		func(in []byte) (interface{}, int) {

			// ORIGIN
			b1 := in[0]
			var olen, origin int
			if checkBit(7, uint(b1)) { //
				if checkBit(6, uint(b1)) { // 4 bytes
					tmp := [4]byte{b1 & 0x3F} // first byte with mask 0011_1111
					copy(tmp[1:], in[1:4])    // remaining 3 bytes
					olen = 4
					origin = int(binary.BigEndian.Uint32(tmp[:]))
				} else { // 2 bytes
					tmp := [2]byte{b1 & 0x3F} // first byte with mask 0011_1111
					tmp[1] = in[1]
					olen = 2
					origin = int(binary.BigEndian.Uint16(tmp[:]))
				}
			} else {
				// single byte
				olen = 1
				origin = int(b1) // bit 7 is 0
			}

			// COPY
			copy := int(in[olen])

			// IDENT
			ln := int(in[olen+1])
			ident := ""
			if ln == 0 {
				ident = ""
			} else {
				ident = string(in[olen+2 : olen+2+ln])
			}
			// only allowed 33-96, 123-126
			// TODO check for allowed

			return struct {
				Origin, Copy int
				Ident        string
			}{
				origin, copy, ident,
			}, int(olen + 2 + ln)
		}}, // 23

	{"OBJREF", 0, "Object reference", nil},    // 24
	{"ATTREF", 0, "Attribute reference", nil}, // 25
	{"STATUS", 1, "Boolean status", nil},      // 26
	{"UNITS", 0, "Units expression", nil},     // 27
}
