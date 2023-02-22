package sparse

import (
	"sort"
)

// CSMatrix is a compressed sparse matrix.
// Used as the base of CSRMatrix and CSCMatrix.
//
// (Shallow-)copying CSMatrix is lightweight.
type CSMatrix struct {
	MajorDim, MinorDim int
	Entries            [][]Entry
}

// Dim asserts the receiver is a square matrix and returns the dimension.
func (m *CSMatrix) Dim() (int, error) {
	if m.MajorDim != m.MinorDim {
		return 0, ErrDimensionMismatch
	}
	return m.MajorDim, nil
}

// SetMajorDim grows/shrinks the receiver in-place,
// so it matches the given major dimension.
func (m *CSMatrix) SetMajorDim(dim int) {
	if cap(m.Entries) < dim {
		m.Entries = append(make([][]Entry, 0, dim), m.Entries...)
	}
	m.Entries = m.Entries[:dim]
	m.MajorDim = dim
}

// SetMinorDim grows/shrinks the receiver in-place,
// so it matches the given minor dimension.
func (m *CSMatrix) SetMinorDim(dim int) {
	if dim < m.MinorDim {
		for maj, entries := range m.Entries {
			end := sort.Search(len(entries),
				func(i int) bool { return entries[i].Index >= dim })
			m.Entries[maj] = entries[:end]
		}
	}
	m.MinorDim = dim
}

// NNZ counts nonzero entries.
func (m *CSMatrix) NNZ() (nnz int) {
	for _, row := range m.Entries {
		nnz += len(row)
	}
	return
}

// Transpose transposes the sparse matrix.
func (m *CSMatrix) Transpose() *CSMatrix {
	nnzs := make([]int, m.MinorDim) // indexed by column
	for _, rowEntries := range m.Entries {
		for _, e := range rowEntries {
			nnzs[e.Index]++
		}
	}
	transposedEntries := make([][]Entry, m.MinorDim)
	for col, nnz := range nnzs {
		transposedEntries[col] = make([]Entry, 0, nnz)
	}
	for row, rowEntries := range m.Entries {
		for _, e := range rowEntries {
			col := e.Index
			transposedEntries[col] = append(transposedEntries[col],
				Entry{Index: row, Value: e.Value})
		}
	}
	return &CSMatrix{
		MajorDim: m.MinorDim,
		MinorDim: m.MajorDim,
		Entries:  transposedEntries,
	}
}

// CSRMatrix is a compressed sparse row matrix.
type CSRMatrix struct {
	CSMatrix
}

// NewCSRMatrix creates a new compressed sparse row matrix
// with the given dimension and entries.
//
// The given entries are sorted in row-column order.
func NewCSRMatrix(
	rows, columns int,
	entries []CooEntry,
) *CSRMatrix {
	var entries2 [][]Entry
	if rows != 0 {
		entries2 = make([][]Entry, rows)
	}
	for _, e := range entries {
		if e.Value == 0 {
			continue
		}
		entries2[e.Row] = append(entries2[e.Row], Entry{
			Index: e.Column,
			Value: e.Value,
		})
	}
	for _, row := range entries2 {
		sort.Sort(entrySort(row))
	}
	return &CSRMatrix{
		CSMatrix{
			MajorDim: rows,
			MinorDim: columns,
			Entries:  entries2,
		},
	}
}

// Rows and Columns return the number of rows/columns.
func (m *CSRMatrix) Rows() int    { return m.MajorDim }
func (m *CSRMatrix) Columns() int { return m.MinorDim }

// SetDim grows/shrinks the receiver in-place,
// so it contains the specified number of rows/columns.
func (m *CSRMatrix) SetDim(rows, cols int) {
	m.SetMajorDim(rows)
	m.SetMinorDim(cols)
}

// RowVector returns the given row as a sparse vector.
// The returned vector shares the same slice of entry objects.
func (m *CSRMatrix) RowVector(index int) *Vector {
	return &Vector{
		Dim:     m.MinorDim,
		Entries: m.Entries[index],
	}
}

// SetRowVector replaces the given row.
// The receiver shares the same slice of entry objects.
func (m *CSRMatrix) SetRowVector(index int, vector *Vector) {
	m.Entries[index] = vector.Entries
}

// Transpose transposes the matrix.
func (m *CSRMatrix) Transpose() *CSRMatrix {
	return &CSRMatrix{
		*m.CSMatrix.Transpose(),
	}
}

// TransposeToCSC transposes the matrix.
// The returned matrix shares the same entry objects.
func (m *CSRMatrix) TransposeToCSC() *CSCMatrix {
	return &CSCMatrix{
		CSMatrix{
			MajorDim: m.MinorDim,
			MinorDim: m.MajorDim,
			Entries:  m.Entries,
		},
	}
}

// CSCMatrix is a compressed sparse column matrix.
type CSCMatrix struct {
	CSMatrix
}

// Rows and Columns return the number of rows/columns
func (m *CSCMatrix) Rows() int    { return m.MinorDim }
func (m *CSCMatrix) Columns() int { return m.MajorDim }

// SetDim grows/shrinks the receiver in-place,
// so it contains the specified number of rows/columns.
func (m *CSCMatrix) SetDim(rows, cols int) {
	m.SetMajorDim(cols)
	m.SetMinorDim(rows)
}

// ColumnVector returns the given row as a sparse vector.
// The returned vector shares the same entry objects.
func (m *CSCMatrix) ColumnVector(index int) *Vector {
	return &Vector{
		Dim:     m.MinorDim,
		Entries: m.Entries[index],
	}
}

// Transpose transposes the matrix.
func (m *CSCMatrix) Transpose() *CSCMatrix {
	return &CSCMatrix{
		*m.CSMatrix.Transpose(),
	}
}

// TransposeToCSR transposes the matrix.
// The returned matrix shares the same entry objects.
func (m *CSCMatrix) TransposeToCSR() *CSRMatrix {
	return &CSRMatrix{
		CSMatrix{
			MajorDim: m.MinorDim,
			MinorDim: m.MajorDim,
			Entries:  m.Entries,
		},
	}
}

// Matrix is just an alias of CSRMatrix,
// which is the more popular of the two compressed sparse variants.
type Matrix = CSRMatrix
