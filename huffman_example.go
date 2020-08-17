// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	hf "github.com/9nut/huffman"
)

func main() {
	str := `this is a long string of many words that can then wrap
	around and keep on going, etc. etc. how now brown cow?`

	fmt.Println("Input:", str)
	input := strings.NewReader(str)

	var compressed bytes.Buffer
	encoder := hf.NewHuffmanEncoder(input, &compressed)
	fmt.Println("Dictionary")
	encoder.ShowHuffTree()
	fmt.Println()
	if err := encoder.WriteHeader(); err != nil {
		fmt.Println("Can't write header", err)
		os.Exit(1)
	}
	io.Copy(encoder, input)
	encoder.Flush()

	fmt.Println("Compressed: ", compressed)
	decoder := hf.NewHuffmanDecoder(&compressed)
	if err := decoder.ReadHeader(); err != nil {
		fmt.Println("Can't read header", err)
		os.Exit(2)
	}
	fmt.Println("Loaded dictionary")
	decoder.ShowHuffTree()
	fmt.Println()
	var decompressed bytes.Buffer
	io.Copy(&decompressed, decoder)
	fmt.Println("Uncompressed:", decompressed.String())
	if str == decompressed.String()[:len(str)] {
		fmt.Println("Success")
	}
}
