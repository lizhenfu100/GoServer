package test

import (
	"bytes"
	"common"
	"strings"
	_ "svr_client/test/init"
	"testing"
)

// go test -v ./src/svr_client/test/ringbuf_test.go

func TestRingBuffer_Write(t *testing.T) {
	rb := common.NewRingBuf(64)
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 64 {
		t.Fatalf("expect free 64 bytes but got %d. %v", rb.WritableBytes(), rb)
	}

	rb.Write([]byte(strings.Repeat("abcd", 4))) //16
	if rb.ReadableBytes() != 16 {
		t.Fatalf("expect len 16 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 48 {
		t.Fatalf("expect free 48 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte(strings.Repeat("abcd", 4))) != 0 {
		t.Fatalf("expect 4 abcd but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	rb.Write([]byte(strings.Repeat("abcd", 12))) //48
	if rb.ReadableBytes() != 64 {
		t.Fatalf("expect len 64 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 0 {
		t.Fatalf("expect free 0 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	if rb.Cap() != 64 {
		t.Fatalf("%v", rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte(strings.Repeat("abcd", 16))) != 0 {
		t.Fatalf("expect 16 abcd but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// write more 4 bytes
	rb.Write([]byte(strings.Repeat("abcd", 1)))
	if rb.ReadableBytes() != 68 {
		t.Fatalf("expect len 64 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 60 {
		t.Fatalf("expect free 0 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	rb.Clear() //128
	rb.Write([]byte(strings.Repeat("abcd", 32)))
	if rb.ReadableBytes() != 128 {
		t.Fatalf("expect len 128 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 0 {
		t.Fatalf("expect free 0 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	if rb.Cap() != 128 {
		t.Fatalf("%v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte(strings.Repeat("abcd", 32))) != 0 {
		t.Fatalf("expect 32 abcd but got %v", rb)
	}

	rb.Clear()                                  //128
	rb.Write([]byte(strings.Repeat("abcd", 2))) //8
	if rb.ReadableBytes() != 8 {
		t.Fatalf("expect len 16 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 120 {
		t.Fatalf("expect free 48 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	buf := make([]byte, 5)
	rb.Read(buf)
	if rb.ReadableBytes() != 3 {
		t.Fatalf("expect len 3 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	rb.Write([]byte(strings.Repeat("abcd", 15))) //60
	if head, _ := rb.Peek(); bytes.Compare(head, []byte("bcd"+strings.Repeat("abcd", 15))) != 0 {
		t.Fatalf("expect 63 ... but got %v", rb)
	}

	rb.Clear()                                    //128
	rb.Write([]byte(strings.Repeat("abcde", 25))) //125
	if rb.WritableBytes() != 3 {
		t.Fatalf("expect free 3 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	buf = make([]byte, 16)
	rb.Read(buf)
	rb.Write([]byte(strings.Repeat("1234", 4))) //16
	if rb.WritableBytes() != 3 {
		t.Fatalf("expect free 0 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	head, tail := rb.Peek()
	buf2 := append(buf, head...)
	buf2 = append(buf2, tail...)
	if !bytes.Equal(buf2, []byte(strings.Repeat("abcde", 25)+strings.Repeat("1234", 4))) {
		t.Fatalf("expect 25 abcde and 4 1234 but got %v", rb)
	}
}

func TestRingBuffer_Read(t *testing.T) {
	rb := common.NewRingBuf(64)
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 64 {
		t.Fatalf("expect free 64 bytes but got %d. %v", rb.WritableBytes(), rb)
	}

	// read empty
	buf := make([]byte, 1024)
	n := rb.Read(buf)
	if n != 0 {
		t.Fatalf("expect read 0 bytes but got %d", n)
	}
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 64 {
		t.Fatalf("expect free 64 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	if rb.Cap() != 64 {
		t.Fatalf("%v", rb)
	}

	rb.Write([]byte(strings.Repeat("abcd", 4))) //16
	n = rb.Read(buf)
	if n != 16 {
		t.Fatalf("expect read 16 bytes but got %d", n)
	}
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 64 {
		t.Fatalf("expect free 64 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	t.Logf("%v", rb)

	rb.Write([]byte(strings.Repeat("abcd", 16))) //64
	n = rb.Read(buf)
	if n != 64 {
		t.Fatalf("expect read 64 bytes but got %d", n)
	}
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 64 {
		t.Fatalf("expect free 64 bytes but got %d. %v", rb.WritableBytes(), rb)
	}
	t.Logf("%v", rb)
}

func TestRingBuffer_ByteInterface(t *testing.T) {
	rb := common.NewRingBuf(2)
	// write one
	rb.Write([]byte{'a'})
	if rb.ReadableBytes() != 1 {
		t.Fatalf("expect len 1 byte but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 1 {
		t.Fatalf("expect free 1 byte but got %d. %v", rb.WritableBytes(), rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte{'a'}) != 0 {
		t.Fatalf("expect a but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// write two, isFull
	rb.Write([]byte{'b'})
	if rb.ReadableBytes() != 2 {
		t.Fatalf("expect len 2 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 0 {
		t.Fatalf("expect free 0 byte but got %d. %v", rb.WritableBytes(), rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte{'a', 'b'}) != 0 {
		t.Fatalf("expect a but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	rb.Write([]byte{'c'}) //4
	if rb.ReadableBytes() != 3 {
		t.Fatalf("expect len 3 bytes but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 1 {
		t.Fatalf("expect free 1 byte but got %d. %v", rb.WritableBytes(), rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte{'a', 'b', 'c'}) != 0 {
		t.Fatalf("expect a but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false %v", rb)
	}

	buf := make([]byte, 1)
	n := rb.Read(buf)
	if n != 1 {
		t.Fatalf("ReadByte failed")
	}
	if buf[0] != 'a' {
		t.Fatalf("expect a but got %c. %v", buf[0], rb)
	}
	if rb.ReadableBytes() != 2 {
		t.Fatalf("expect len 2 byte but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 2 {
		t.Fatalf("expect free 2 byte but got %d. %v", rb.WritableBytes(), rb)
	}
	if head, _ := rb.Peek(); bytes.Compare(head, []byte{'b', 'c'}) != 0 {
		t.Fatalf("expect a but got %v", rb)
	}
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	rb.Read(buf)
	if buf[0] != 'b' {
		t.Fatalf("expect b but got %c. %v", buf[0], rb)
	}
	if rb.ReadableBytes() != 1 {
		t.Fatalf("expect len 1 byte but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 3 {
		t.Fatalf("expect free 3 byte but got %d. %v", rb.WritableBytes(), rb)
	}

	rb.Read(buf)
	if rb.ReadableBytes() != 0 {
		t.Fatalf("expect len 0 byte but got %d. %v", rb.ReadableBytes(), rb)
	}
	if rb.WritableBytes() != 4 {
		t.Fatalf("expect free 4 byte but got %d. %v", rb.WritableBytes(), rb)
	}
}
