package huffman

import (
	"io"

	bs "github.com/dgryski/go-bitstream"
)

// HuffmanDecoder holds the information for a Huffman
// decoder and provides an io.Reader interface.
type HuffmanDecoder struct {
	root Item
	br   *bs.BitReader
}

// Prepare a new Huffman decoderfor the given io.Reader
func NewHuffmanDecoder(rc io.Reader) *HuffmanDecoder {
	hd := new(HuffmanDecoder)
	hd.br = bs.NewReader(rc)
	return hd
}

// Read and extract the Huffman tree from the file
func (dec *HuffmanDecoder) ReadHeader() (err error) {
	// for iterative tree loading use loaddict
	// for recursive, use rrloaddict

	// dec.root, err = loaddict(dec.br)
	dec.root, err = rloaddict(dec.br, nil) // recursive versione
	return err
}

// Huffman decoder that satisfies the io.Reader interface
func (dec *HuffmanDecoder) Read(buff []byte) (n int, err error) {
	for i := range buff {
		b, err := readNextSymbol(dec.br, dec.root)
		if err != nil {
			return len(buff) - i, err
		}
		buff[i] = b
	}
	return len(buff), nil
}

// Show the Huffman tree
func (dec *HuffmanDecoder) ShowHuffTree() {
	traverse(dec.root, "")
}

// read the next huffman symbole and return its bytecode value
func readNextSymbol(br *bs.BitReader, root interface{}) (val byte, err error) {
	t, ok := root.(*HTree)
	for ok {
		b, err := br.ReadBit()
		if err != nil {
			return 0, err
		}

		switch b {
		case bs.Zero:
			root = t.left
		case bs.One:
			root = t.right
		}
		t, ok = root.(*HTree)
	}
	node, ok := root.(*HNode)
	if ok {
		return node.value, nil
	}
	return 0, io.EOF
}

