// Copyright 2017 Roman Kachanovsky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*

	Modified by @antisnatchor & @marver

*/

/*
	Package binary_pack performs conversions between some Go values represented as byte slices.
	This can be used in handling binary data stored in files or from network connections,
	among other sources. It uses format slices of strings as compact descriptions of the layout
	of the Go structs.

	Format characters (some characters like H have been reserved for future implementation of unsigned numbers):
		? - bool, packed size 1 byte
		h, H - int, packed size 2 bytes (in future it will support pack/unpack of int8, uint8 values)
		i, I, l, L - int, packed size 4 bytes (in future it will support pack/unpack of int16, uint16, int32, uint32 values)
		q, Q - int, packed size 8 bytes (in future it will support pack/unpack of int64, uint64 values)
		f - float32, packed size 4 bytes
		d - float64, packed size 8 bytes
		Ns - string, packed size N bytes, N is a number of runes to pack/unpack

 */
package binary_pack

import (
	"strings"
	"strconv"
	"errors"
	"encoding/binary"
	"bytes"
	"fmt"
)

type BinaryPack struct {}

// Return a byte slice containing the values of msg slice packed according to the given format.
// The items of msg slice must match the values required by the format exactly.
func (bp *BinaryPack) Pack(format []string, msg []interface{}) ([]byte, error) {
	if len(format) > len(msg) {
		return nil, errors.New(fmt.Sprintf("Format (%d) is longer than values (%d) to pack", len(format), len(msg)))
	}

	var endianess binary.ByteOrder = binary.BigEndian // defaults to big endian

	res := []byte{}

	for i, f := range format {
		if f[0] == '<' {
			// little endian
			endianess = binary.LittleEndian
			f = f[1:]
		} else if f[0] == '>' {
			endianess = binary.BigEndian
			f = f[1:]
		}
		switch f {
		case "?":
			casted_value, ok := msg[i].(bool)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (bool)")
			}
			res = append(res, boolToBytes(endianess, casted_value)...)
		case "B":
			casted_value, ok := msg[i].(uint8)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (int, 1 bytes)")
			}
			res = append(res, uint8ToBytes(endianess, casted_value, 1)...)
		case "h", "H":
			casted_value, ok := msg[i].(uint16)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (int, 2 bytes)")
			}
			res = append(res, uint16ToBytes(endianess, casted_value, 2)...)
		case "i", "I", "l", "L":
			casted_value, ok := msg[i].(uint32)
			if !ok {
				return nil, errors.New("Type of passed value " + string(msg[i].(uint32)) + " doesn't match to expected '" + f + "' (int, 4 bytes)")
			}
			res = append(res, uint32ToBytes(endianess, casted_value, 4)...)
		case "q", "Q":
			casted_value, ok := msg[i].(uint64)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (int, 8 bytes)")
			}
			res = append(res, uint64ToBytes(endianess, casted_value, 8)...)
		case "f":
			casted_value, ok := msg[i].(float32)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (float32)")
			}
			res = append(res, float32ToBytes(endianess, casted_value, 4)...)
		case "d":
			casted_value, ok := msg[i].(float64)
			if !ok {
				return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (float64)")
			}
			res = append(res, float64ToBytes(endianess, casted_value, 8)...)
		default:
			if strings.Contains(f, "s") {
				casted_value, ok := msg[i].(string)
				if !ok {
					return nil, errors.New("Type of passed value doesn't match to expected '" + f + "' (string)")
				}
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				res = append(res, []byte(fmt.Sprintf("%s%s",
					casted_value, strings.Repeat("\x00", n-len(casted_value))))...)
			} else {
				return nil, errors.New("Unexpected format token: '" + f + "'")
			}
		}
	}

	return res, nil
}

// Unpack the byte slice (presumably packed by Pack(format, msg)) according to the given format.
// The result is a []interface{} slice even if it contains exactly one item.
// The byte slice must contain not less the amount of data required by the format
// (len(msg) must more or equal CalcSize(format)).
func (bp *BinaryPack) UnPack(format []string, msg []byte) ([]interface{}, error) {
	expected_size, err := bp.CalcSize(format)

	if err != nil {
		return nil, err
	}

	if expected_size > len(msg) {
		return nil, errors.New("Expected size is bigger than actual size of message")
	}

	res := []interface{}{}

	var endianess binary.ByteOrder = binary.BigEndian // default big endian

	for _, f := range format {
		if f[0] == '<' {
			// little endian
			endianess = binary.LittleEndian
			f = f[1:]
		} else if f[0] == '>' {
			endianess = binary.BigEndian
			f = f[1:]
		}
		switch f {
		case "?":
			res = append(res, bytesToBool(endianess, msg[:1]))
			msg = msg[1:]
		case "B":
			res = append(res, bytesToInt8(endianess, msg[:1]))
			msg = msg[1:]
		case "h", "H":
			res = append(res, bytesToInt16(endianess, msg[:2]))
			msg = msg[2:]
		case "i", "I", "l", "L":
			res = append(res, bytesToInt32(endianess, msg[:4]))
			msg = msg[4:]
		case "q", "Q":
			res = append(res, bytesToInt64(endianess, msg[:8]))
			msg = msg[8:]
		case "f":
			res = append(res, bytesToFloat32(endianess, msg[:4]))
			msg = msg[4:]
		case "d":
			res = append(res, bytesToFloat64(endianess, msg[:8]))
			msg = msg[8:]
		default:
			if strings.Contains(f, "s") {
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				res = append(res, string(msg[:n]))
				msg = msg[n:]
			} else {
				return nil, errors.New("Unexpected format token: '" + f + "'")
			}
		}
	}

	return res, nil
}

// Return the size of the struct (and hence of the byte slice) corresponding to the given format.
func (bp *BinaryPack) CalcSize(format []string) (int, error) {
	var size int

	for _, f := range format {
		// skip endianess switches
		if f[0] == '<' || f[0] == '>' {
			f = f[1:]
		}
		switch f {
		case "?":
			size = size + 1
		case "B":
			size = size + 1
		case "h", "H":
			size = size + 2
		case "i", "I", "l", "L", "f":
			size = size + 4
		case "q", "Q", "d":
			size = size + 8
		default:
			if strings.Contains(f, "s") {
				n, _ := strconv.Atoi(strings.TrimRight(f, "s"))
				size = size + n
			} else {
				return 0, errors.New("Unexpected format token: '" + f + "'")
			}
		}
	}

	return size, nil
}

func boolToBytes(endianess binary.ByteOrder, x bool) []byte {
	if x {
		return uint32ToBytes(endianess, 1, 1)
	}
	return uint32ToBytes(endianess, 0, 1)
}

func bytesToBool(endianess binary.ByteOrder, b []byte) bool {
	return bytesToInt8(endianess, b) > 0
}

func uint32ToBytes(endianess binary.ByteOrder, n uint32, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, uint32(n))
	return buf.Bytes()[0:size]
}

func uint64ToBytes(endianess binary.ByteOrder, n uint64, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, uint64(n))
	return buf.Bytes()[0:size]
}

func uint16ToBytes(endianess binary.ByteOrder, n uint16, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, uint16(n))
	return buf.Bytes()[0:size]
}

func uint8ToBytes(endianess binary.ByteOrder, n uint8, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, uint8(n))
	return buf.Bytes()[0:size]
}

func bytesToInt8(endianess binary.ByteOrder, b []byte) uint8 {
	buf := bytes.NewBuffer(b)

	var x uint8
	binary.Read(buf, endianess, &x)
	return x
}

func bytesToInt16(endianess binary.ByteOrder, b []byte) uint16 {

	buf := bytes.NewBuffer(b)

	var x uint16
	binary.Read(buf, endianess, &x)
	return x
}
func bytesToInt32(endianess binary.ByteOrder, b []byte) uint32 {
	buf := bytes.NewBuffer(b)

	var x uint32
	binary.Read(buf, endianess, &x)
	return x
}
func bytesToInt64(endianess binary.ByteOrder, b []byte) uint64 {
	buf := bytes.NewBuffer(b)

	var x uint64
	binary.Read(buf, endianess, &x)
	return x
}

func float32ToBytes(endianess binary.ByteOrder, n float32, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, n)
	return buf.Bytes()[0:size]
}

func bytesToFloat32(endianess binary.ByteOrder, b []byte) float32 {
	var x float32
	buf := bytes.NewBuffer(b)
	binary.Read(buf, endianess, &x)
	return x
}

func float64ToBytes(endianess binary.ByteOrder, n float64, size int) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, endianess, n)
	return buf.Bytes()[0:size]
}

func bytesToFloat64(endianess binary.ByteOrder, b []byte) float64 {
	var x float64
	buf := bytes.NewBuffer(b)
	binary.Read(buf, endianess, &x)
	return x
}
