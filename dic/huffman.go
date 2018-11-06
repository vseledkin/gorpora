package dic

import (
	"time"
	"sort"
	"log"
	"container/heap"
)

/*
HuffmanNode - huffman tree node
*/
type HuffmanNode struct {
	*Word
	left, right *HuffmanNode
}

/*
TreeHeap - huffman tree nodes slice
*/
type TreeHeap []HuffmanNode

func (th TreeHeap) Len() int { return len(th) }
func (th TreeHeap) Less(i, j int) bool {
	if th[i].Count == th[j].Count {
		return th[i].Word.Word < th[j].Word.Word
	}
	return th[i].Count < th[j].Count
}

/*
Push heap interface implementation
*/
func (th *TreeHeap) Push(e interface{}) {
	*th = append(*th, e.(HuffmanNode))
}

/*
Pop heap interface implementation
*/
func (th *TreeHeap) Pop() (popped interface{}) {
	popped = (*th)[len(*th)-1]
	*th = (*th)[:len(*th)-1]
	return
}

func (th TreeHeap) Swap(i, j int) { th[i], th[j] = th[j], th[i] }

func BuildHuffmanTreeFromDictionary(d *Dictionary) (codes [][]byte, points [][]uint32, root HuffmanNode) {
	_time := time.Now()
	L := len(d.Words)
	th := make(TreeHeap, L)
	var i int
	for k, wc := range d.Words {
		th[i].Word = &Word{Count: wc.Count, Word: k}
		i++
	}
	sort.Sort(sort.Reverse(th))
	root = buildTree(th)
	var maxDepth int
	points = make([][]uint32, L)
	codes = make([][]byte, L)
	assignCodes(root, codes, []byte{}, points, []uint32{}, &maxDepth, uint32(L))
	log.Printf("built huffman tree with maximum node depth %d for %v\n", maxDepth, time.Now().Sub(_time))
	return
}

func buildTree(trees TreeHeap) HuffmanNode {
	words := len(trees)
	heap.Init(&trees)
	var i int
	for trees.Len() > 1 {
		// two trees with least frequency
		a := heap.Pop(&trees).(HuffmanNode)
		b := heap.Pop(&trees).(HuffmanNode)
		// put into new node and re-insert into queue
		word := Word{ID: uint32(words + i), Count: uint32(a.Count + b.Count)}
		heap.Push(&trees, HuffmanNode{&word, &a, &b})
		i++
	}
	return heap.Pop(&trees).(HuffmanNode)
}

func assignCodes(tree HuffmanNode, codes [][]byte, prefix []byte, points [][]uint32, point []uint32, maxdepth *int, wordCount uint32) {
	if tree.ID < wordCount {
		points[tree.ID] = append(points[tree.ID], point...)
		codes[tree.ID] = append(codes[tree.ID], prefix...)
		// print out symbol, frequency, and code for this
		// leaf (which is just the prefix)
		codelength := len(prefix)
		if codelength > *maxdepth {
			*maxdepth = codelength
		}
	} else {
		point = append(point, tree.ID-wordCount)
		// traverse left branch
		assignCodes(*tree.left, codes, append(prefix, 0), points, point, maxdepth, wordCount)

		// traverse right branch
		assignCodes(*tree.right, codes, append(prefix, 1), points, point, maxdepth, wordCount)
	}
}
