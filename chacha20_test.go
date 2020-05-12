// Copryright (C) 2019 Yawning Angel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package chacha20

import (
	"crypto/rand"
	"crypto/sha512"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fengxuway/chacha20/internal/api"
)

// Test vectors taken from:
// https://tools.ietf.org/html/draft-strombergson-chacha-test-vectors-01
var draftTestVectors = []struct {
	name       string
	key        []byte
	iv         []byte
	stream     []byte
	seekOffset uint64
}{
	{
		name: "IETF Draft: TC1: All zero key and IV.",
		key: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		iv: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		stream: []byte{
			0x76, 0xb8, 0xe0, 0xad, 0xa0, 0xf1, 0x3d, 0x90,
			0x40, 0x5d, 0x6a, 0xe5, 0x53, 0x86, 0xbd, 0x28,
			0xbd, 0xd2, 0x19, 0xb8, 0xa0, 0x8d, 0xed, 0x1a,
			0xa8, 0x36, 0xef, 0xcc, 0x8b, 0x77, 0x0d, 0xc7,
			0xda, 0x41, 0x59, 0x7c, 0x51, 0x57, 0x48, 0x8d,
			0x77, 0x24, 0xe0, 0x3f, 0xb8, 0xd8, 0x4a, 0x37,
			0x6a, 0x43, 0xb8, 0xf4, 0x15, 0x18, 0xa1, 0x1c,
			0xc3, 0x87, 0xb6, 0x69, 0xb2, 0xee, 0x65, 0x86,
			0x9f, 0x07, 0xe7, 0xbe, 0x55, 0x51, 0x38, 0x7a,
			0x98, 0xba, 0x97, 0x7c, 0x73, 0x2d, 0x08, 0x0d,
			0xcb, 0x0f, 0x29, 0xa0, 0x48, 0xe3, 0x65, 0x69,
			0x12, 0xc6, 0x53, 0x3e, 0x32, 0xee, 0x7a, 0xed,
			0x29, 0xb7, 0x21, 0x76, 0x9c, 0xe6, 0x4e, 0x43,
			0xd5, 0x71, 0x33, 0xb0, 0x74, 0xd8, 0x39, 0xd5,
			0x31, 0xed, 0x1f, 0x28, 0x51, 0x0a, 0xfb, 0x45,
			0xac, 0xe1, 0x0a, 0x1f, 0x4b, 0x79, 0x4d, 0x6f,
		},
	},
	{
		name: "IETF Draft: TC2: Single bit in key set. All zero IV.",
		key: []byte{
			0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		iv: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		stream: []byte{
			0xc5, 0xd3, 0x0a, 0x7c, 0xe1, 0xec, 0x11, 0x93,
			0x78, 0xc8, 0x4f, 0x48, 0x7d, 0x77, 0x5a, 0x85,
			0x42, 0xf1, 0x3e, 0xce, 0x23, 0x8a, 0x94, 0x55,
			0xe8, 0x22, 0x9e, 0x88, 0x8d, 0xe8, 0x5b, 0xbd,
			0x29, 0xeb, 0x63, 0xd0, 0xa1, 0x7a, 0x5b, 0x99,
			0x9b, 0x52, 0xda, 0x22, 0xbe, 0x40, 0x23, 0xeb,
			0x07, 0x62, 0x0a, 0x54, 0xf6, 0xfa, 0x6a, 0xd8,
			0x73, 0x7b, 0x71, 0xeb, 0x04, 0x64, 0xda, 0xc0,
			0x10, 0xf6, 0x56, 0xe6, 0xd1, 0xfd, 0x55, 0x05,
			0x3e, 0x50, 0xc4, 0x87, 0x5c, 0x99, 0x30, 0xa3,
			0x3f, 0x6d, 0x02, 0x63, 0xbd, 0x14, 0xdf, 0xd6,
			0xab, 0x8c, 0x70, 0x52, 0x1c, 0x19, 0x33, 0x8b,
			0x23, 0x08, 0xb9, 0x5c, 0xf8, 0xd0, 0xbb, 0x7d,
			0x20, 0x2d, 0x21, 0x02, 0x78, 0x0e, 0xa3, 0x52,
			0x8f, 0x1c, 0xb4, 0x85, 0x60, 0xf7, 0x6b, 0x20,
			0xf3, 0x82, 0xb9, 0x42, 0x50, 0x0f, 0xce, 0xac,
		},
	},
	{
		name: "IETF Draft: TC3: Single bit in IV set. All zero key.",
		key: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		iv: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		stream: []byte{
			0xef, 0x3f, 0xdf, 0xd6, 0xc6, 0x15, 0x78, 0xfb,
			0xf5, 0xcf, 0x35, 0xbd, 0x3d, 0xd3, 0x3b, 0x80,
			0x09, 0x63, 0x16, 0x34, 0xd2, 0x1e, 0x42, 0xac,
			0x33, 0x96, 0x0b, 0xd1, 0x38, 0xe5, 0x0d, 0x32,
			0x11, 0x1e, 0x4c, 0xaf, 0x23, 0x7e, 0xe5, 0x3c,
			0xa8, 0xad, 0x64, 0x26, 0x19, 0x4a, 0x88, 0x54,
			0x5d, 0xdc, 0x49, 0x7a, 0x0b, 0x46, 0x6e, 0x7d,
			0x6b, 0xbd, 0xb0, 0x04, 0x1b, 0x2f, 0x58, 0x6b,
			0x53, 0x05, 0xe5, 0xe4, 0x4a, 0xff, 0x19, 0xb2,
			0x35, 0x93, 0x61, 0x44, 0x67, 0x5e, 0xfb, 0xe4,
			0x40, 0x9e, 0xb7, 0xe8, 0xe5, 0xf1, 0x43, 0x0f,
			0x5f, 0x58, 0x36, 0xae, 0xb4, 0x9b, 0xb5, 0x32,
			0x8b, 0x01, 0x7c, 0x4b, 0x9d, 0xc1, 0x1f, 0x8a,
			0x03, 0x86, 0x3f, 0xa8, 0x03, 0xdc, 0x71, 0xd5,
			0x72, 0x6b, 0x2b, 0x6b, 0x31, 0xaa, 0x32, 0x70,
			0x8a, 0xfe, 0x5a, 0xf1, 0xd6, 0xb6, 0x90, 0x58,
		},
	},
	{
		name: "IETF Draft: TC4: All bits in key and IV are set.",
		key: []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		iv: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		stream: []byte{
			0xd9, 0xbf, 0x3f, 0x6b, 0xce, 0x6e, 0xd0, 0xb5,
			0x42, 0x54, 0x55, 0x77, 0x67, 0xfb, 0x57, 0x44,
			0x3d, 0xd4, 0x77, 0x89, 0x11, 0xb6, 0x06, 0x05,
			0x5c, 0x39, 0xcc, 0x25, 0xe6, 0x74, 0xb8, 0x36,
			0x3f, 0xea, 0xbc, 0x57, 0xfd, 0xe5, 0x4f, 0x79,
			0x0c, 0x52, 0xc8, 0xae, 0x43, 0x24, 0x0b, 0x79,
			0xd4, 0x90, 0x42, 0xb7, 0x77, 0xbf, 0xd6, 0xcb,
			0x80, 0xe9, 0x31, 0x27, 0x0b, 0x7f, 0x50, 0xeb,
			0x5b, 0xac, 0x2a, 0xcd, 0x86, 0xa8, 0x36, 0xc5,
			0xdc, 0x98, 0xc1, 0x16, 0xc1, 0x21, 0x7e, 0xc3,
			0x1d, 0x3a, 0x63, 0xa9, 0x45, 0x13, 0x19, 0xf0,
			0x97, 0xf3, 0xb4, 0xd6, 0xda, 0xb0, 0x77, 0x87,
			0x19, 0x47, 0x7d, 0x24, 0xd2, 0x4b, 0x40, 0x3a,
			0x12, 0x24, 0x1d, 0x7c, 0xca, 0x06, 0x4f, 0x79,
			0x0f, 0x1d, 0x51, 0xcc, 0xaf, 0xf6, 0xb1, 0x66,
			0x7d, 0x4b, 0xbc, 0xa1, 0x95, 0x8c, 0x43, 0x06,
		},
	},
	{
		name: "IETF Draft: TC5: Every even bit set in key and IV.",
		key: []byte{
			0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
			0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
			0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
			0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
		},
		iv: []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55},
		stream: []byte{
			0xbe, 0xa9, 0x41, 0x1a, 0xa4, 0x53, 0xc5, 0x43,
			0x4a, 0x5a, 0xe8, 0xc9, 0x28, 0x62, 0xf5, 0x64,
			0x39, 0x68, 0x55, 0xa9, 0xea, 0x6e, 0x22, 0xd6,
			0xd3, 0xb5, 0x0a, 0xe1, 0xb3, 0x66, 0x33, 0x11,
			0xa4, 0xa3, 0x60, 0x6c, 0x67, 0x1d, 0x60, 0x5c,
			0xe1, 0x6c, 0x3a, 0xec, 0xe8, 0xe6, 0x1e, 0xa1,
			0x45, 0xc5, 0x97, 0x75, 0x01, 0x7b, 0xee, 0x2f,
			0xa6, 0xf8, 0x8a, 0xfc, 0x75, 0x80, 0x69, 0xf7,
			0xe0, 0xb8, 0xf6, 0x76, 0xe6, 0x44, 0x21, 0x6f,
			0x4d, 0x2a, 0x34, 0x22, 0xd7, 0xfa, 0x36, 0xc6,
			0xc4, 0x93, 0x1a, 0xca, 0x95, 0x0e, 0x9d, 0xa4,
			0x27, 0x88, 0xe6, 0xd0, 0xb6, 0xd1, 0xcd, 0x83,
			0x8e, 0xf6, 0x52, 0xe9, 0x7b, 0x14, 0x5b, 0x14,
			0x87, 0x1e, 0xae, 0x6c, 0x68, 0x04, 0xc7, 0x00,
			0x4d, 0xb5, 0xac, 0x2f, 0xce, 0x4c, 0x68, 0xc7,
			0x26, 0xd0, 0x04, 0xb1, 0x0f, 0xca, 0xba, 0x86,
		},
	},
	{
		name: "IETF Draft: TC6: Every odd bit set in key and IV.",
		key: []byte{
			0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
			0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
			0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
			0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
		},
		iv: []byte{0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa},
		stream: []byte{
			0x9a, 0xa2, 0xa9, 0xf6, 0x56, 0xef, 0xde, 0x5a,
			0xa7, 0x59, 0x1c, 0x5f, 0xed, 0x4b, 0x35, 0xae,
			0xa2, 0x89, 0x5d, 0xec, 0x7c, 0xb4, 0x54, 0x3b,
			0x9e, 0x9f, 0x21, 0xf5, 0xe7, 0xbc, 0xbc, 0xf3,
			0xc4, 0x3c, 0x74, 0x8a, 0x97, 0x08, 0x88, 0xf8,
			0x24, 0x83, 0x93, 0xa0, 0x9d, 0x43, 0xe0, 0xb7,
			0xe1, 0x64, 0xbc, 0x4d, 0x0b, 0x0f, 0xb2, 0x40,
			0xa2, 0xd7, 0x21, 0x15, 0xc4, 0x80, 0x89, 0x06,
			0x72, 0x18, 0x44, 0x89, 0x44, 0x05, 0x45, 0xd0,
			0x21, 0xd9, 0x7e, 0xf6, 0xb6, 0x93, 0xdf, 0xe5,
			0xb2, 0xc1, 0x32, 0xd4, 0x7e, 0x6f, 0x04, 0x1c,
			0x90, 0x63, 0x65, 0x1f, 0x96, 0xb6, 0x23, 0xe6,
			0x2a, 0x11, 0x99, 0x9a, 0x23, 0xb6, 0xf7, 0xc4,
			0x61, 0xb2, 0x15, 0x30, 0x26, 0xad, 0x5e, 0x86,
			0x6a, 0x2e, 0x59, 0x7e, 0xd0, 0x7b, 0x84, 0x01,
			0xde, 0xc6, 0x3a, 0x09, 0x34, 0xc6, 0xb2, 0xa9,
		},
	},
	{
		name: "IETF Draft: TC7: Sequence patterns in key and IV.",
		key: []byte{
			0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
			0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
			0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88,
			0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00,
		},
		iv: []byte{0x0f, 0x1e, 0x2d, 0x3c, 0x4b, 0x5a, 0x69, 0x78},
		stream: []byte{
			0x9f, 0xad, 0xf4, 0x09, 0xc0, 0x08, 0x11, 0xd0,
			0x04, 0x31, 0xd6, 0x7e, 0xfb, 0xd8, 0x8f, 0xba,
			0x59, 0x21, 0x8d, 0x5d, 0x67, 0x08, 0xb1, 0xd6,
			0x85, 0x86, 0x3f, 0xab, 0xbb, 0x0e, 0x96, 0x1e,
			0xea, 0x48, 0x0f, 0xd6, 0xfb, 0x53, 0x2b, 0xfd,
			0x49, 0x4b, 0x21, 0x51, 0x01, 0x50, 0x57, 0x42,
			0x3a, 0xb6, 0x0a, 0x63, 0xfe, 0x4f, 0x55, 0xf7,
			0xa2, 0x12, 0xe2, 0x16, 0x7c, 0xca, 0xb9, 0x31,
			0xfb, 0xfd, 0x29, 0xcf, 0x7b, 0xc1, 0xd2, 0x79,
			0xed, 0xdf, 0x25, 0xdd, 0x31, 0x6b, 0xb8, 0x84,
			0x3d, 0x6e, 0xde, 0xe0, 0xbd, 0x1e, 0xf1, 0x21,
			0xd1, 0x2f, 0xa1, 0x7c, 0xbc, 0x2c, 0x57, 0x4c,
			0xcc, 0xab, 0x5e, 0x27, 0x51, 0x67, 0xb0, 0x8b,
			0xd6, 0x86, 0xf8, 0xa0, 0x9d, 0xf8, 0x7e, 0xc3,
			0xff, 0xb3, 0x53, 0x61, 0xb9, 0x4e, 0xbf, 0xa1,
			0x3f, 0xec, 0x0e, 0x48, 0x89, 0xd1, 0x8d, 0xa5,
		},
	},
	{
		name: "IETF Draft: TC8: key: 'All your base are belong to us!, IV: 'IETF2013'",
		key: []byte{
			0xc4, 0x6e, 0xc1, 0xb1, 0x8c, 0xe8, 0xa8, 0x78,
			0x72, 0x5a, 0x37, 0xe7, 0x80, 0xdf, 0xb7, 0x35,
			0x1f, 0x68, 0xed, 0x2e, 0x19, 0x4c, 0x79, 0xfb,
			0xc6, 0xae, 0xbe, 0xe1, 0xa6, 0x67, 0x97, 0x5d,
		},
		iv: []byte{0x1a, 0xda, 0x31, 0xd5, 0xcf, 0x68, 0x82, 0x21},
		stream: []byte{
			0xf6, 0x3a, 0x89, 0xb7, 0x5c, 0x22, 0x71, 0xf9,
			0x36, 0x88, 0x16, 0x54, 0x2b, 0xa5, 0x2f, 0x06,
			0xed, 0x49, 0x24, 0x17, 0x92, 0x30, 0x2b, 0x00,
			0xb5, 0xe8, 0xf8, 0x0a, 0xe9, 0xa4, 0x73, 0xaf,
			0xc2, 0x5b, 0x21, 0x8f, 0x51, 0x9a, 0xf0, 0xfd,
			0xd4, 0x06, 0x36, 0x2e, 0x8d, 0x69, 0xde, 0x7f,
			0x54, 0xc6, 0x04, 0xa6, 0xe0, 0x0f, 0x35, 0x3f,
			0x11, 0x0f, 0x77, 0x1b, 0xdc, 0xa8, 0xab, 0x92,
			0xe5, 0xfb, 0xc3, 0x4e, 0x60, 0xa1, 0xd9, 0xa9,
			0xdb, 0x17, 0x34, 0x5b, 0x0a, 0x40, 0x27, 0x36,
			0x85, 0x3b, 0xf9, 0x10, 0xb0, 0x60, 0xbd, 0xf1,
			0xf8, 0x97, 0xb6, 0x29, 0x0f, 0x01, 0xd1, 0x38,
			0xae, 0x2c, 0x4c, 0x90, 0x22, 0x5b, 0xa9, 0xea,
			0x14, 0xd5, 0x18, 0xf5, 0x59, 0x29, 0xde, 0xa0,
			0x98, 0xca, 0x7a, 0x6c, 0xcf, 0xe6, 0x12, 0x27,
			0x05, 0x3c, 0x84, 0xe4, 0x9a, 0x4a, 0x33, 0x32,
		},
	},
	{
		name: "XChaCha20 Test",
		key: []byte{
			0x54, 0x68, 0x65, 0x20, 0x70, 0x75, 0x72, 0x70,
			0x6f, 0x73, 0x65, 0x20, 0x6f, 0x66, 0x20, 0x74,
			0x68, 0x65, 0x20, 0x70, 0x69, 0x73, 0x74, 0x6f,
			0x6c, 0x20, 0x69, 0x73, 0x20, 0x74, 0x6f, 0x20,
		},
		iv: []byte{
			0x73, 0x74, 0x6f, 0x70, 0x20, 0x61, 0x20, 0x66,
			0x69, 0x67, 0x68, 0x74, 0x20, 0x74, 0x68, 0x61,
			0x74, 0x20, 0x73, 0x6f, 0x6d, 0x65, 0x62, 0x6f,
		},
		stream: []byte{
			0xfa, 0x27, 0xe8, 0x54, 0xba, 0x9f, 0x74, 0xf0,
			0x7f, 0xa3, 0x85, 0x03, 0x65, 0x1b, 0xc9, 0xbf,
			0xe8, 0x80, 0xed, 0x2b, 0xf7, 0x20, 0xed, 0xc5,
			0xb4, 0x82, 0xc9, 0xea, 0x95, 0x32, 0x59, 0x20,
			0xb9, 0xf5, 0x86, 0x93, 0xf1, 0xf7, 0xa6, 0x67,
			0x4e, 0xc4, 0xf3, 0xfa, 0x93, 0x14, 0x4d, 0xe2,
			0xe9, 0x6d, 0x0d, 0x4a, 0xe7, 0x70, 0x32, 0xae,
			0x88, 0xb7, 0x58, 0x3b, 0xd5, 0x5f, 0x16, 0xa6,
			0xd5, 0x0e, 0x8d, 0x53, 0x0c, 0x60, 0x8e, 0x41,
			0x5c, 0xda, 0x0c, 0xe9, 0xc7, 0xde, 0xc6, 0x71,
			0x9d, 0x8a, 0xe5, 0xcb, 0xde, 0x45, 0x5b, 0x98,
			0xf6, 0x57, 0xca, 0x9d, 0xd1, 0xa1, 0xe3, 0x5c,
			0x01, 0xe6, 0x29, 0x38, 0x9e, 0xf5, 0x97, 0x38,
			0x75, 0x61, 0x0e, 0x71, 0xfc, 0xd1, 0xa1, 0x11,
			0x29, 0x47, 0x37, 0x6b, 0xac, 0x09, 0xd6, 0xee,
			0xf7, 0x61, 0x62, 0x24, 0x00, 0x58, 0xac, 0x1d,
			0xcf, 0x42, 0xef, 0x50, 0xfa, 0x54, 0x14, 0x49,
			0x31, 0xd2, 0x62,
		},
	},
	{
		name: "RFC 7539 Test Vector (96 bit nonce)",
		key: []byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		},
		iv: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4a,
			0x00, 0x00, 0x00, 0x00,
		},
		stream: []byte{
			0x22, 0x4f, 0x51, 0xf3, 0x40, 0x1b, 0xd9, 0xe1,
			0x2f, 0xde, 0x27, 0x6f, 0xb8, 0x63, 0x1d, 0xed,
			0x8c, 0x13, 0x1f, 0x82, 0x3d, 0x2c, 0x06, 0xe2,
			0x7e, 0x4f, 0xca, 0xec, 0x9e, 0xf3, 0xcf, 0x78,
			0x8a, 0x3b, 0x0a, 0xa3, 0x72, 0x60, 0x0a, 0x92,
			0xb5, 0x79, 0x74, 0xcd, 0xed, 0x2b, 0x93, 0x34,
			0x79, 0x4c, 0xba, 0x40, 0xc6, 0x3e, 0x34, 0xcd,
			0xea, 0x21, 0x2c, 0x4c, 0xf0, 0x7d, 0x41, 0xb7,
			0x69, 0xa6, 0x74, 0x9f, 0x3f, 0x63, 0x0f, 0x41,
			0x22, 0xca, 0xfe, 0x28, 0xec, 0x4d, 0xc4, 0x7e,
			0x26, 0xd4, 0x34, 0x6d, 0x70, 0xb9, 0x8c, 0x73,
			0xf3, 0xe9, 0xc5, 0x3a, 0xc4, 0x0c, 0x59, 0x45,
			0x39, 0x8b, 0x6e, 0xda, 0x1a, 0x83, 0x2c, 0x89,
			0xc1, 0x67, 0xea, 0xcd, 0x90, 0x1d, 0x7e, 0x2b,
			0xf3, 0x63,
		},
		seekOffset: 1,
	},
}

func TestChaCha20(t *testing.T) {
	for _, v := range supportedImpls {
		testWithImpl := func(t *testing.T, impl api.Implementation) {
			oldImpl := activeImpl
			defer func() {
				activeImpl = oldImpl
			}()

			activeImpl = impl

			t.Run("Basic", doTestBasic)
			t.Run("TestVectors", doTestVectors)
		}

		t.Run(v.Name(), func(t *testing.T) {
			testWithImpl(t, v)
		})
	}
}

func doTestBasic(t *testing.T) {
	t.Run("RoundTrip", doTestBasicRoundTrip)
	t.Run("Counter", doTestBasicCounter)
	t.Run("IETFCounter", doTestBasicIETFCounter)
	t.Run("Incremental", doTestBasicIncremental)
}

func doTestBasicRoundTrip(t *testing.T) {
	require := require.New(t)

	var (
		key   [KeySize]byte
		nonce [NonceSize]byte
	)

	c, err := New(key[:], nonce[:])
	require.NoError(err, "New")

	plaintext := []byte("The smallest minority on earth is the individual.  Those who deny individual rights cannot claim to be defenders of minorities.")
	ciphertext := make([]byte, len(plaintext))
	c.XORKeyStream(ciphertext, plaintext)

	require.NotEqual(plaintext, ciphertext, "XORKeyStream - output")

	err = c.Seek(0)
	require.NoError(err, "Seek")
	c.XORKeyStream(ciphertext, ciphertext)

	require.Equal(plaintext, ciphertext, "XORKeyStream - round trip output")
}

func doTestBasicCounter(t *testing.T) {
	require := require.New(t)

	var (
		key   [KeySize]byte
		nonce [NonceSize]byte

		block, block2 [api.BlockSize]byte
	)

	c, err := New(key[:], nonce[:])
	require.NoError(err, "New")

	c.KeyStream(block[:]) // Block: 0

	err = c.Seek(math.MaxUint32 + 1)
	require.NoError(err, "Seek")
	c.KeyStream(block2[:]) // Block: 0x100000000

	require.NotEqual(block, block2, "KeyStream - 32 bit counter would wrap")
}

func doTestBasicIETFCounter(t *testing.T) {
	require := require.New(t)

	var (
		key   [KeySize]byte
		nonce [INonceSize]byte

		block [api.BlockSize]byte
	)

	c, err := New(key[:], nonce[:])
	require.NoError(err, "New")

	err = c.Seek(math.MaxUint32 - 1)
	require.NoError(err, "Seek")
	c.KeyStream(block[:])

	require.Panics(func() {
		c.KeyStream(block[:])
	}, "KeyStream - counter would wrap")
}

func doTestBasicIncremental(t *testing.T) {
	require := require.New(t)

	var (
		key   [KeySize]byte
		nonce [NonceSize]byte

		buf [2048]byte

		expectedDigest = []byte{
			0xcf, 0xd6, 0xe9, 0x49, 0x22, 0x5b, 0x85, 0x4f,
			0xe0, 0x49, 0x46, 0x49, 0x1e, 0x69, 0x35, 0xff,
			0x05, 0xff, 0x98, 0x3d, 0x15, 0x54, 0xbc, 0x88,
			0x5b, 0xca, 0x0e, 0xc8, 0x08, 0x2d, 0xd5, 0xb8,
		}
	)

	c, err := New(key[:], nonce[:])
	require.NoError(err, "New")

	h := sha512.New512_256()
	for i := 1; i <= 2048; i++ {
		c.KeyStream(buf[:i])
		_, _ = h.Write(buf[:i])
	}

	digest := h.Sum(nil)

	require.Equal(expectedDigest, digest, "KeyStream digest matches")
}

func doTestVectors(t *testing.T) {
	for _, v := range draftTestVectors {
		t.Run(v.name, func(t *testing.T) {
			require := require.New(t)
			c, err := New(v.key, v.iv)
			require.NoError(err, "New")

			if v.seekOffset != 0 {
				err = c.Seek(v.seekOffset)
				require.NoErrorf(err, "Seek(%d)", v.seekOffset)
			}

			out := make([]byte, len(v.stream))
			c.XORKeyStream(out, out)
			require.EqualValues(v.stream, out, "XORKeyStream")

			// Ensure that KeyStream tramples over dst.
			_, err = rand.Read(out)
			require.NoError(err, "rand.Read")

			err = c.Seek(v.seekOffset)
			require.NoErrorf(err, "Seek(%d) - resetting", v.seekOffset)
			c.KeyStream(out)
			require.EqualValues(v.stream, out, "KeyStream")
		})
	}
}

func BenchmarkChaCha20(b *testing.B) {
	for _, v := range supportedImpls {
		benchWithImpl := func(b *testing.B, impl api.Implementation) {
			oldImpl := activeImpl
			defer func() {
				activeImpl = oldImpl
			}()

			activeImpl = impl
			for _, vv := range []int{
				1, 8, 32, 64, 576, 1536, 4096, 1024768,
			} {
				b.Run(strconv.Itoa(vv), func(b *testing.B) {
					doBenchN(b, vv)
				})
			}
		}

		b.Run(v.Name(), func(b *testing.B) {
			benchWithImpl(b, v)
		})
	}
}

func doBenchN(b *testing.B, n int) {
	var (
		key   [KeySize]byte
		nonce [NonceSize]byte
	)

	s := make([]byte, n)
	c, err := New(key[:], nonce[:])
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(n))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.XORKeyStream(s, s)
	}
}
