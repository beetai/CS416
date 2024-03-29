// These tests follow the methodology discussed here https://golang.org/doc/tutorial/add-a-test
// Specific testing API functions are described in more detail here https://golang.org/pkg/testing/
package biguint

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func prettyPrintUInt8Slice(slice []uint8) string {
	result := "["
	for i, b := range slice {
		if i != 0 {
			if i%4 == 0 {
				result += ";"
			}
			result += " "
		}
		result += fmt.Sprintf("%x", b)
	}
	result += "]"
	return result
}

func TestReadFromInt64(t *testing.T) {
	type Test struct {
		input    uint64
		expected []uint8
	}
	tests := []Test{
		{0xffffffff_ffffffff, []uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0x00ff00ff_00ff00ff, []uint8{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}},
		{0xff00ff00_ff00ff00, []uint8{0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}},
		{0x12345678_87654321, []uint8{0x21, 0x43, 0x65, 0x87, 0x78, 0x56, 0x34, 0x12}},
	}
	doIt := func(pfx string, fn func(uint64) []uint8) {
		for _, test := range tests {
			t.Run(fmt.Sprintf("0x%x", test.input), func(t *testing.T) {
				result := fn(test.input)
				if !reflect.DeepEqual(test.expected, result) {
					t.Fatalf("%s: %s does not equal expected value %s", pfx, prettyPrintUInt8Slice(result), prettyPrintUInt8Slice(test.expected))
				}
			})
		}
	}
	doIt("private interface", bytesFromUInt64)
	doIt("public interface", func(i uint64) []uint8 {
		return NewBigUInt(i).Bytes()
	})
}

func TestString(t *testing.T) {
	type Test struct {
		input    uint64
		expected string
	}
	tests := []Test{
		{0x12345678_87654321, "0x12345678_87654321"},
		{0xf, "0xf"},
		{0x100, "0x100"},
		{0x1, "0x1"},
		{0x0, "0x0"},
		{0x1_12345678, "0x1_12345678"},
		{0x12_12345678, "0x12_12345678"},
		{0x123_12345678, "0x123_12345678"},
		{0x1234_12345678, "0x1234_12345678"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%x", test.input), func(t *testing.T) {
			result := NewBigUInt(test.input).String()
			if test.expected != result {
				t.Fatalf("%s, does not equal expected value %s", result, test.expected)
			}
		})
	}
}

func TestCopy(t *testing.T) {

	type Test struct {
		input []uint8
	}

	tests := []Test{
		{[]uint8{0xff}},
		{[]uint8{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}},
		{[]uint8{0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff}},
		{[]uint8{0x21, 0x43, 0x65, 0x87, 0x78, 0x56, 0x34, 0x12}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%x", test.input), func(t *testing.T) {
			var source BigUInt
			source.data = test.input
			dest := source.Copy()

			if len(dest.data) != len(source.data) {
				t.Fatalf("Copy Failed, copied bytes: %d; Expected: %d ", len(dest.data), len(source.data))
			}

			if source.String() != dest.String() {
				t.Fatalf("%s, does not equal expected value %s", dest.String(), source.String())
			}

			/*
			* Check if both piont to the same slice object
			* Now modify dest, and check if source and dest still match
			 */
			rindex := rand.Intn(len(dest.data))
			dest.data[rindex]++

			if source.data[rindex] == dest.data[rindex] {

				t.Fatal("Both Source and Destination point to the same object")
			}
		})
	}
}

func TestAdd(t *testing.T) {
	type Test struct {
		lhs      uint64
		rhs      uint64
		expected string
	}
	tests := []Test{
		{0x2, 0x2, "0x4"},
		{0xff, 1, "0x100"},
		{0x100000ff, 0x100000ff, "0x200001fe"},
		{0x0, 0xff, "0xff"},
		{0xff00, 0xff, "0xffff"},
		{0xff, 0x0, "0xff"},
		{0x0, 0x0, "0x0"},
		{0x100000ff, 0xff, "0x100001fe"},
		{0xff, 0x100000ff, "0x100001fe"},
		{0xff, 0x10000fff, "0x100010fe"},
		{0xfffffff_ffffffff, 0x1, "0x10000000_00000000"},
		{0x1, 0xfffffff_ffffffff, "0x10000000_00000000"},
		{0x00000000_00000000, 0x00000000_00000000, "0x0"},
		{0xff, 0x0000, "0xff"},
		{0x00, 0x0000, "0x0"},
		{0xc, 0xfffff, "0x10000b"},
		{0xfffff, 0xc, "0x10000b"},
		{0xff, 0xff, "0x1fe"},
		{0xffffffff_ffffffff, 0xffffffff_ffffffff, "0x1_ffffffff_fffffffe"},
		{0xffffffff_ffffffff, 0xff, "0x1_00000000_000000fe"},
		{0xffffffff_ffffffff, 0xf, "0x1_00000000_0000000e"},
	}
	//biggerThanMax64 := NewBigUInt(0)
	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%x + 0x%x", test.lhs, test.rhs), func(t *testing.T) {
			result := NewBigUInt(test.lhs).Add(NewBigUInt(test.rhs))
			//biggerThanMax64 = result
			resultStr := result.String()
			if test.expected != resultStr {
				t.Fatalf("%s, %s does not equal expected value %s", resultStr, prettyPrintUInt8Slice(result.Bytes()), test.expected)
			}
		})
	}

	//bigger := biggerThanMax64.Add(biggerThanMax64)
	//biggerStr := bigger.String()
	//if biggerStr != "0x3_ffffffff_fffffffc" {
	//	t.Fatalf("%s, %s does not equal expected value %s", biggerStr, prettyPrintUInt8Slice(bigger.Bytes()), "0x3_ffffffff_fffffffc")
	//}

}

func TestSubtract(t *testing.T) {
	type Test struct {
		lhs         uint64
		rhs         uint64
		expected    string
		errExpected error
	}
	tests := []Test{
		{0x2, 0x2, "0x0", nil},
		{0xff, 0x1, "0xfe", nil},
		{0xf0, 0xf, "0xe1", nil},
		{0xff00, 0x00f0, "0xfe10", nil},
		{0xfe10, 0x0, "0xfe10", nil},
		{0xf000, 0x1, "0xefff", nil},
		{0xf00000, 0x1, "0xefffff", nil},
		{0xefffff, 0x0, "0xefffff", nil},
		{0xff, 0xff, "0x0", nil},
		{0x0, 0x1, "", ErrUnderflow},
		{0x0, 0x0, "0x0", nil},
		{0x1, 0x0, "0x1", nil},
		{0x0, 0x1ff, "", ErrUnderflow},
		{0x101, 0x200, "", ErrUnderflow},
		{0xffffffff_ffffffff, 0xffffffff_ffffffff, "0x0", nil},
		{0xf0000000, 0xf0000001, "", ErrUnderflow},
		{0xffffff_ffffffff, 0xffffff_ffffffff, "0x0", nil},
		{0xffffffff_ffffffff, 0xffffffff_ffffffee, "0x11", nil},
		{0x0, 0x0, "0x0", nil},
		{0x10000000_00000000, 0x1, "0xfffffff_ffffffff", nil},
		{0xffffffff_ffffffff, 0xff, "0xffffffff_ffffff00", nil},
		{0xffffffff_ffffffff, 0xf, "0xffffffff_fffffff0", nil},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("0x%x - 0x%x", test.lhs, test.rhs), func(t *testing.T) {
			lhs := NewBigUInt(test.lhs)
			result, err := lhs.Subtract(NewBigUInt(test.rhs))
			if test.errExpected != nil {
				if test.errExpected != err {
					t.Fatalf("Expected error %v, got %v", test.errExpected, err)
				}
				if result != nil {
					t.Fatalf("Expected nil result in error case, got %v", result)
				}
				expectedBytes := bytesFromUInt64(test.lhs)
				if !reflect.DeepEqual(expectedBytes, lhs.Bytes()) {
					t.Fatalf(
						"Expected no change to lhs in error case, but expected %s was not equal to actual %s",
						prettyPrintUInt8Slice(expectedBytes), prettyPrintUInt8Slice(lhs.Bytes()))
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error %v, with result %v. expected %s", err, result, test.expected)
				}
				resultStr := result.String()
				if test.expected != resultStr {
					t.Fatalf("%s, %s does not equal expected value %s", resultStr, prettyPrintUInt8Slice(result.Bytes()), test.expected)
				}
			}
		})
	}
	// subtracting more then uint64
	biggerThanMax64 := NewBigUInt(0xffffffff_ffffffff).Add(NewBigUInt(0xffffffff_ffffffff))
	biggerThanMax64Twice := NewBigUInt(0xffffffff_ffffffff).Add(NewBigUInt(0xffffffff_ffffffff))
	biggerThanMax64 = biggerThanMax64.Add(biggerThanMax64)
	biggerThanMax64Twice = biggerThanMax64Twice.Add(biggerThanMax64Twice)
	biggerThanMax64Twice = biggerThanMax64Twice.Add(biggerThanMax64Twice)
	subtracted, err := biggerThanMax64Twice.Subtract(biggerThanMax64)
	if err != nil {
		t.Fatalf("unexpected error %v, with result %v. expected %s", err, subtracted, subtracted.String())
	}
	subtractedStr := subtracted.String()
	if subtractedStr != biggerThanMax64.String() {
		t.Fatalf("%s, %s does not equal expected value %s", subtractedStr, prettyPrintUInt8Slice(subtracted.Bytes()), biggerThanMax64.String())
	}


}

func TestBigSubtract(t *testing.T) {
	type Test struct {
		lhs         []uint8
		rhs         []uint8
		expected    string
		errExpected error
	}
	tests := []Test{
		{[]uint8{}, []uint8{}, "0x0", nil},
		{[]uint8{0xff}, []uint8{0xfe}, "0x1", nil},
		{[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			[]uint8{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "0x1", nil},
		{[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, []uint8{0xff}, "0xffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffff00", nil},
		{[]uint8{0xff}, []uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "", ErrUnderflow},
	}

	for _, test := range tests {
		lhs := &BigUInt{data: test.lhs}
		rhs := &BigUInt{data: test.rhs}
		t.Run(fmt.Sprintf("%s - %s", lhs.String(), rhs.String()), func(t *testing.T) {
			result, err := lhs.Subtract(rhs)
			if test.errExpected != nil {
				if test.errExpected != err {
					t.Fatalf("Expected error %v, got %v", test.errExpected, err)
				}
				if result != nil {
					t.Fatalf("Expected nil result in error case, got %v", result)
				}
				expectedBytes := test.lhs
				if !reflect.DeepEqual(expectedBytes, lhs.Bytes()) {
					t.Fatalf(
						"Expected no change to lhs in error case, but expected %s was not equal to actual %s",
						prettyPrintUInt8Slice(expectedBytes), prettyPrintUInt8Slice(lhs.Bytes()))
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error %v, with result %v. expected %s", err, result, test.expected)
				}
				resultStr := result.String()
				if test.expected != resultStr {
					t.Fatalf("%s, %s does not equal expected value %s", resultStr, prettyPrintUInt8Slice(result.Bytes()), test.expected)
				}
			}
		})
	}
}
func TestBigAdd(t *testing.T) {
	type Test struct {
		lhs      []uint8
		rhs      []uint8
		expected string
	}
	tests := []Test{
		{[]uint8{}, []uint8{}, "0x0"},
		{[]uint8{0xff}, []uint8{0xff}, "0x1fe"},
		{[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, []uint8{0x0}, "0xff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff"},
		{[]uint8{0x0}, []uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "0xff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff_ffffffff"},
		{[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, []uint8{0xff}, "0x1_00000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000_000000fe"},
		{[]uint8{0xff}, []uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "0x1_00000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000_000000fe"},
	}

	for _, test := range tests {
		lhs := &BigUInt{data: test.lhs}
		rhs := &BigUInt{data: test.rhs}
		t.Run(fmt.Sprintf("%v + %v", lhs.String(), rhs.String()), func(t *testing.T) {

			result := lhs.Add(rhs)
			resultStr := result.String()
			if test.expected != resultStr {
				t.Fatalf("%s, %s does not equal expected value %s", resultStr, prettyPrintUInt8Slice(result.Bytes()), test.expected)
			}
		})
	}
}