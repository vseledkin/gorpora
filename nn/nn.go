package nn

import (
	"github.com/vseledkin/gorpora/asm"
	"math"
)

type Graph struct {
	NeedsBackprop bool
	backprop      []func()
}

func (g *Graph) Backward() {
	for i := len(g.backprop) - 1; i >= 0; i-- {
		g.backprop[i]()
	}
}

type Matrix struct {
	Rows    int //number of rows
	Columns int // number of columns
	W       []float32
	DW      []float32 `json:"-"`
}

func Zeros(n int) []float32 {
	return make([]float32, n)
}

func Mat(rows, columns int) *Matrix {
	M := new(Matrix)
	M.Rows = rows
	M.Columns = columns
	M.W = Zeros(rows * columns)
	M.DW = Zeros(rows * columns)
	return M
}

func InputVector(v []float32) *Matrix {
	M := new(Matrix)
	M.Rows = len(v)
	M.Columns = 1
	M.W = v
	M.DW = make([]float32, M.Rows)
	return M
}

func (m *Matrix) SameAs() (mm *Matrix) {
	mm = Mat(m.Rows, m.Columns)
	return
}

func (m *Matrix) CopyAs() (mm *Matrix) {
	mm = Mat(m.Rows, m.Columns)
	copy(mm.W, m.W)
	return
}

func (g *Graph) Add(m ... *Matrix) *Matrix {
	out := m[0].CopyAs() // copy only weights not gradients
	for i := range m[1:] {
		asm.Sxpy(m[i].W, out.W)
	}

	if g.NeedsBackprop {
		g.backprop = append(g.backprop, func() {
			for i := range m {
				asm.Sxpy(out.DW, m[i].DW)
			}
		})
	}
	return out
}

// EMul elementwise matrix matrix multiplication
func (g *Graph) EMul(m1, m2 *Matrix) *Matrix {
	out := m1.CopyAs()
	asm.Sxmuley(m2.W, out.W)

	if g.NeedsBackprop {
		g.backprop = append(g.backprop, func() {
			asm.Sxmuleyplusz(m2.W, out.DW, m1.DW)
			asm.Sxmuleyplusz(m1.W, out.DW, m2.DW)
		})
	}
	return out
}

func (g *Graph) Dot(m1, m2 *Matrix) *Matrix {
	out := m1.CopyAs()
	asm.Sxmuley(m2.W, out.W)
	out.W[0] = asm.Ssum(out.W)
	out.Rows = 1
	out.Columns = 1
	out.W = out.W[:1]
	out.DW = out.DW[:1]
	if g.NeedsBackprop {
		g.backprop = append(g.backprop, func() {
			asm.Saxpy(out.DW[0], m1.W, m2.DW)
			asm.Saxpy(out.DW[0], m2.W, m1.DW)
		})
	}
	return out
}

// sigmoid nonlinearity
func (g *Graph) Sigmoid(m *Matrix) *Matrix {
	out := m.SameAs()

	for i := range m.W {
		out.W[i] = 1.0 / (1.0 + float32(math.Exp(float64(-m.W[i])))) // Sigmoid
	}

	if g.NeedsBackprop {
		g.backprop = append(g.backprop, func() {
			// grad for z = sigmoid(x) is sigmoid(x)(1 - sigmoid(x))
			for i := range m.W {
				m.DW[i] += out.W[i] * (1.0 - out.W[i]) * out.DW[i]
			}
		})
	}
	return out
}
