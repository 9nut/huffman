// Huffman package test file

package huffman

import (
	"bytes"
	"io"
	"strings"

	"testing"
)

func TestHuffman(t *testing.T) {
	str := `this is a long string of many words that can then wrap
	around and keep on going, etc. etc. how now brown cow?`

	input := strings.NewReader(str)

	var compressed bytes.Buffer
	encoder := NewHuffmanEncoder(input, &compressed)
	if err := encoder.WriteHeader(); err != nil {
		t.Error(err)
	}
	io.Copy(encoder, input)
	encoder.Flush()

	decoder := NewHuffmanDecoder(&compressed)
	if err := decoder.ReadHeader(); err != nil {
		t.Error(err)
	}

	var decompressed bytes.Buffer
	io.Copy(&decompressed, decoder)
	if str != decompressed.String()[:len(str)] {
		t.Errorf("Expected %s: got %s\n", str, decompressed.String())
	}
}
