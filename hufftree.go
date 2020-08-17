package huffman

import (
	"errors"
	"fmt"
	"strconv"

	bs "github.com/dgryski/go-bitstream"
)

// Huffman compression utility functions to build a
// huffman tree. Recursive and iterative versions of
// functions are provided as examples.

// Item interface is used to abstract away
// the details of Huffman tree from PriorityQueue
type Item interface {
	Freq() int
}

// HNode satisfies the Item interface; it holds
// the uncompressed code value of a symbol and its
// frequency in the source file
type HNode struct {
	freq  int
	value byte
}

// HTree satisfies the Item interface; its
// frequency is the sum of all the frequencies of
// its subtrees
type HTree struct {
	freq  int
	left  Item
	right Item
}

// NewHNode creates a new HNode for the given
// value, set to the given frequency
func NewHNode(val byte, frq int) *HNode {
	return &HNode{freq: frq, value: val}
}

// NewHTree creates a binary HTree with the
// provided left and right subtrees set to
// their combined frequencies
func NewHTree(ll, rr Item) *HTree {
	return &HTree{freq: ll.Freq()+rr.Freq(), left: ll, right: rr}
}

// Freq is used by PriorityQueue to order the heap
func (hn *HNode) Freq() int {
	return hn.freq
}

// Freq is used to order the heap in PriorityQueue
func (ht *HTree) Freq() int {
	return ht.freq
}

// Huffcode holds the compressed code for
// each symbol in nbits of the lowest bits in hcode
type Huffcode struct {
	freq  int
	hcode uint64
	nbits int
}

// create a lookup map of bytecode->huffcode
func filldict(r interface{}, code string, lkup map[byte]Huffcode) {
	t, ok := r.(*HTree)
	if ok {
		filldict(t.left, code+"0", lkup)
		filldict(t.right, code+"1", lkup)
		return
	}
	n, ok := r.(*HNode)
	if ok {
		if res, err := strconv.ParseInt(code, 2, 32); err == nil {
			lkup[n.value] = Huffcode{freq: n.freq, hcode: uint64(res), nbits: len(code)}
		}
	}
}

// recursive post order traversal
func rsavedict(bw *bs.BitWriter, r interface{}) error {
	t, ok := r.(*HTree)
	if ok {
		rsavedict(bw, t.left)
		rsavedict(bw, t.right)
		return bw.WriteBit(bs.Zero)
	}
	n, ok := r.(*HNode)
	if ok {
		bw.WriteBit(bs.One)
		return bw.WriteByte(n.value)
	}
	return errors.New("Internal error saving dictionary")
}

// iterative post order traversal using two stacks
func savedict(bw *bs.BitWriter, root interface{}) error {
	stack1 := &symstack{}
	stack2 := &symstack{}
	stack1.Push(root.(*HTree))
	for stack1.Len() > 0 {
		r := stack1.Pop()
		stack2.Push(r)
		t, ok := r.(*HTree)
		if ok {
			stack1.Push(t.left)
			stack1.Push(t.right)
		}
	}
	for stack2.Len() > 0 {
		r := stack2.Pop()
		_, ok := r.(*HTree)
		if ok {
			fmt.Print("0")
			bw.WriteBit(bs.Zero)
			continue
		}
		n, ok := r.(*HNode)
		if ok {
			fmt.Printf("1%q", n.value)
			bw.WriteBit(bs.One)
			if err := bw.WriteByte(n.value); err != nil {
				return err
			}
		}
	}
	fmt.Println()
	return nil
}

// one could easily use a slice as a LIFO directly
// in the function, but the Stack abstraction is
// easier to follow.
type symstack []Item

func (stk *symstack) Push(v Item) {
	*stk = append(*stk, v)
}

func (stk *symstack) Pop() Item {
	v := (*stk)[len(*stk)-1]
	*stk = (*stk)[:len(*stk)-1]
	return v
}

func (stk *symstack) Len() int {
	return len(*stk)
}

// iterative load dictionary
// load the huffman tree in post order
// 1 bit is followed by an uncompressed byte value for a symbol
// 0 bit is preceeded by two edges (nodes or subtrees), else if
// only one edge or subtree remaining, that's the root of the tree
func loaddict(br *bs.BitReader) (tree Item, err error) {
	stack := &symstack{}

	for {
		b, err := br.ReadBit()
		if err != nil {
			return nil, err
		}
		// log.Println("loaddict")
		if b == bs.One {
			c, err := br.ReadBits(8)
			if err != nil {
				return nil, err
			}
			stack.Push(&HNode{value: byte(c & 0xff)})
		} else {
			if stack.Len() < 2 {
				return stack.Pop(), nil
			}
			r := stack.Pop()
			l := stack.Pop()
			stack.Push(&HTree{left: l, right: r})
		}
	}
}

// recursive load dictionary
// if the first bit is a ONE, build a node with the 8-bit literal value of
// the symbol and add it to the LIFO lifo; passing it down the recursion path.
// when a ZERO bit is encountered, if the LIFO has at least 2 elements,
// remove 2 elements and build a subtree node with them, then add subtree
// to the LIFO; otherwise (less than 2 elements in LIFO), return the remaining
// element as the root.
//
// Using a Stack instead of the lifo slice would make the code more readable.
func rloaddict(br *bs.BitReader, lifo []Item) (tree Item, err error) {
	b, err := br.ReadBit()
	if err != nil {
		return nil, err
	}
	// log.Println("loaddict")
	if b == bs.One {
		c, err := br.ReadBits(8)
		if err != nil {
			return nil, err
		}
		lifo = append(lifo, &HNode{value: byte(c & 0xff)})
		return rloaddict(br, lifo)
	} else {
		// b == bs.Zero
		if len(lifo) > 1 {
			rr := lifo[len(lifo)-1]
			ll := lifo[len(lifo)-2]
			lifo = lifo[:len(lifo)-2]
			lifo = append(lifo, &HTree{left: ll, right: rr})
			return rloaddict(br, lifo)
		}
		return lifo[0], nil
	}
}

// recursive in-order traversal to print the huffman tree
func traverse(r interface{}, code string) {
	t, ok := r.(*HTree)
	if ok {
		traverse(t.left, code+"0")
		traverse(t.right, code+"1")
		return
	}
	n, ok := r.(*HNode)
	if ok {
		fmt.Printf("%d\t%q\t%s\n", n.freq, n.value, code)
	}
}

