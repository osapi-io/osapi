// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package types

import "slices"

// FieldIn is where a field of a request goes.
type FieldIn string

// The values of FieldIn: the places of a field.
const (
	// FieldInPath is in the path.
	FieldInPath FieldIn = "path"
	// FieldInQuery is in the query.
	FieldInQuery FieldIn = "query"
	// FieldInBody is in the body.
	FieldInBody FieldIn = "body"
)

// Values are the places of a field.
func (FieldIn) Values() []FieldIn {
	return []FieldIn{FieldInPath, FieldInQuery, FieldInBody}
}

// Valid reports whether f is one of the values.
func (f FieldIn) Valid() bool {
	return slices.Contains(f.Values(), f)
}

// FieldKind is how a form takes the value of a field.
type FieldKind string

// The values of FieldKind: the kinds of an input field.
const (
	// FieldString is text.
	FieldString FieldKind = "string"
	// FieldNumber is a number.
	FieldNumber FieldKind = "number"
	// FieldBool is true or false.
	FieldBool FieldKind = "bool"
	// FieldList is a list of values.
	FieldList FieldKind = "list"
	// FieldObject is an object.
	FieldObject FieldKind = "object"
)

// Values are the kinds of an input field.
func (FieldKind) Values() []FieldKind {
	return []FieldKind{FieldString, FieldNumber, FieldBool, FieldList, FieldObject}
}

// Valid reports whether f is one of the values.
func (f FieldKind) Valid() bool {
	return slices.Contains(f.Values(), f)
}

// ColumnKind is what a field of an answer holds.
type ColumnKind string

// The values of ColumnKind: the kinds of an output field.
const (
	// ColumnValue holds one value.
	ColumnValue ColumnKind = "value"
	// ColumnValues holds a list of values.
	ColumnValues ColumnKind = "values"
	// ColumnObject holds an object.
	ColumnObject ColumnKind = "object"
	// ColumnList holds a list of objects.
	ColumnList ColumnKind = "list"
)

// Values are the kinds of an output field.
func (ColumnKind) Values() []ColumnKind {
	return []ColumnKind{ColumnValue, ColumnValues, ColumnObject, ColumnList}
}

// Valid reports whether c is one of the values.
func (c ColumnKind) Valid() bool {
	return slices.Contains(c.Values(), c)
}
