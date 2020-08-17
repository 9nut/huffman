package huffman

import (
	"container/heap"
	"errors"
	"io"

	bs "github.com/dgryski/go-bitstream"
)

// HuffmanEncoder holds the information for a Huffman
// encoder, and provides a io.Writer interface.
type HuffmanEncoder struct {
	root *HTree
	dict map[byte]Huffcode
	bw   *bs.BitWriter
}

// NewHuffmanEncoder creates an encoder from the input io.ReadSeeker and
// prepares it for writing to output io.Writer.  It calculates the dictionary
// by doing frequency counting on the input bytes.
func NewHuffmanEncoder(inp io.ReadSeeker, wc io.Writer) *HuffmanEncoder {
	he := new(HuffmanEncoder)
	freq := make(map[byte]int)

	var b [1]byte
	// using the reader, count the frequency of bytes
	for {
		_, err := inp.Read(b[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		_, ok := freq[b[0]]
		if !ok {
			freq[b[0]] = 0
		}
		freq[b[0]]++
	}
	_, err := inp.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	pQ := make(PriorityQueue, len(freq))
	i := 0
	for v, f := range freq {
		pQ[i] = NewHNode(v, f)
		i++
	}

	heap.Init(&pQ)

	for pQ.Len() > 1 {
		zero := pQ.Pop()
		l := zero.(Item)
		one := pQ.Pop()
		r := one.(Item)
		ht := NewHTree(l, r)
		heap.Push(&pQ, ht)
	}

	htree := pQ.Pop()
	root, ok := htree.(*HTree)
	if !ok {
		panic("Huffman Tree")
	}
	he.root = root
	he.dict = make(map[byte]Huffcode)
	filldict(he.root, "", he.dict)
	he.bw = bs.NewWriter(wc)
	return he
}


// NewHuffmanEncoder creates a new encoder given an existing dictionary.
// It prepares the encoder to write to the io.Writer that is provided.
// The order of the dictionary slice determines its priority.
func NewHuffmanEncoderWithDict(wc io.Writer, dict []byte) *HuffmanEncoder {
	he := new(HuffmanEncoder)

	pQ := make(PriorityQueue, len(dict))
	MaxPri := len(dict)
	for i, v := range dict {
		pQ[i] = NewHNode(v, MaxPri - i)	// prioritize in order of dict
	}

	heap.Init(&pQ)

	for pQ.Len() > 1 {
		zero := pQ.Pop()
		l := zero.(Item)
		one := pQ.Pop()
		r := one.(Item)
		ht := NewHTree(l, r)
		heap.Push(&pQ, ht)
	}

	htree := pQ.Pop()
	root, ok := htree.(*HTree)
	if !ok {
		panic("Huffman Tree")
	}
	he.root = root
	he.dict = make(map[byte]Huffcode)
	filldict(he.root, "", he.dict)
	he.bw = bs.NewWriter(wc)
	return he
}

// Writes the Huffman tree that will be used for decoding
//
// XXX - should probably save the uncompressed file size (in bytes)
// in the header; this would allow the reader to know how big the file
// should be. It might make sense to also have the compressed size
// as well, to be able to perform a sanity check that junk isn't being
// added to the end of the file somehow.
func (enc *HuffmanEncoder) WriteHeader() error {
	// for iterative tree walking use savedict
	// for recursive, use rsavedict

	// if err := savedict(enc.bw, enc.root); err != nil {
	if err := rsavedict(enc.bw, enc.root); err != nil {		// recursive version
		return err
	}
	return enc.bw.WriteBit(bs.Zero) // end of dictionary indicator
}

// Satisfies the io.Writer interface
func (enc *HuffmanEncoder) Write(p []byte) (n int, err error) {
	for _, v := range p {
		code, ok := enc.dict[v]
		if !ok {
			panic(errors.New("non-existant uncompressed code " + string(v)))
		}

		err = enc.bw.WriteBits(code.hcode, code.nbits)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Flushes incomplete bytes to output
func (enc *HuffmanEncoder) Flush() error {
	return enc.bw.Flush(bs.Zero)
}

// Print the constructed Huffman tree
func (enc *HuffmanEncoder) ShowHuffTree() {
	traverse(enc.root, "")
}

