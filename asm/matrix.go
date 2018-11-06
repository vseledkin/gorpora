package asm

import (
	"fmt"
	"math/rand"
	"time"
)

//Matrix32 - dence matrix type
type Matrix32 struct {
	m             []float32
	t             *Matrix32
	rows, columns uint32
}

//At get element at i-th row j-th column
func (m *Matrix32) At(i, j uint32) float32 {
	return m.m[m.columns*i+j]
}

//Mult matrix to vector product
func (m *Matrix32) Mult(v, target []float32) {
	if m.columns != uint32(len(v)) {
		e := fmt.Errorf("Cannot multiply matrix (%d,%d) to vector (%d,1)", m.rows, m.columns, len(v))
		panic(e)
	}
	if m.rows != uint32(len(target)) {
		e := fmt.Errorf("Cannot put result of multiply matrix (%d,%d) to vector (%d,1) into vector (%d,1)", m.rows, m.columns, len(v), len(target))
		panic(e)
	}

	var i uint32

	for ; i < m.rows; i++ {
		target[i] = Sdot(m.m[i*m.columns:i*m.columns+m.columns], v)
	}
}

//Shift matrix to vector*vector product result
func (m *Matrix32) Shift(rate float32, x, y []float32) {
	if m.columns != uint32(len(y)) || m.rows != uint32(len(x)) {
		e := fmt.Errorf("Cannot shift matrix (%d,%d) by vector*vector product (%d,1)*T(%d,1)", m.rows, m.columns, len(x), len(y))
		panic(e)
	}
	// columns of the shift matrix are rows of transposed matrix
	// so m.t is a new transposed matrix
	var i, mtc, mtr, mtshift uint32
	mt := m.T()
	mtc = mt.columns
	mtr = mt.rows
	for ; i < mtr; i++ {
		mtshift = i * mtc
		Saxpy(rate*y[i], x, mt.m[mtshift:mtshift+mtc])
	}
	// we have new updated weight matrix in its transposed form
	// now restore original matrix visit tm column by column
	var row, column, index uint32
	for ; column < mtc; column++ {
		for row = 0; row < mtr; row++ {
			m.m[index] = mt.m[mtc*row+column]
			index++
		}

	}
}

//T construct transposed matrix
func (m *Matrix32) T() *Matrix32 {
	if m.t == nil {

		fmt.Println(m.rows)
		mt := NewMatrix32(m.columns, m.rows)
		mt.t = m
		m.t = mt
		var index uint32
		m.VisitColumnByColumnElements(func(elp *float32) {
			mt.m[index] = *elp
			index++
		})
		return mt
	}
	return m.t
}

//NewMatrix32 new dence matrix
func NewMatrix32(rows, columns uint32) *Matrix32 {
	return &Matrix32{m: make([]float32, rows*columns),
		t:    nil,
		rows: rows, columns: columns,
	}
}

//NewMatrix32FromArray new dence matrix
func NewMatrix32FromArray(array [][]float32) *Matrix32 {
	rows := uint32(len(array))
	columns := uint32(len(array[0]))
	m := &Matrix32{
		make([]float32, rows*columns),
		nil,
		rows, columns,
	}
	var i uint32
	for ; i < rows; i++ {
		copy(m.m[i*columns:], array[i])
	}
	return m
}

//NewRandomMatrix32 new dence matrix initlialized with random values from [-1.0,1.0]
func NewRandomMatrix32(rows, columns uint32, upper, lower float32) *Matrix32 {
	m := NewMatrix32(rows, columns)
	m.VisitRowByRowElements(func(elp *float32) {
		*elp = rand.Float32()*(upper-lower) + lower
		//*elp = rand.Float32()*0.2 - 0.1
	})
	return m
}

//VisitRowByRowElements - visit all matrix elements row by row
func (m *Matrix32) VisitRowByRowElements(visitor func(elp *float32)) {
	var row, column uint32
	for ; row < m.rows; row++ {
		for column = 0; column < m.columns; column++ {
			var f float32
			visitor(&f)
			m.m[m.columns*row+column] = f
		}
	}
}

//VisitColumnByColumnElements - visit all matrix elements column by column
func (m *Matrix32) VisitColumnByColumnElements(visitor func(elp *float32)) {
	var row, column uint32
	for ; column < m.columns; column++ {
		for row = 0; row < m.rows; row++ {
			visitor(&m.m[m.columns*row+column])
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
