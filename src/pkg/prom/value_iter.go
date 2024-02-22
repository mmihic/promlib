package prom

import (
	"time"

	"github.com/prometheus/common/model"
)

// A ValueIter iterates over a set of values.
type ValueIter interface {
	Next() bool

	// Metric returns the metric of the current element.
	Metric() model.Metric

	// Timestamp returns the timestamp of the current element.
	Timestamp() time.Time

	// FloatValue returns the value of the current element (if it is a number).
	FloatValue() float64

	// StringValue returns the value of the current element as a string.
	StringValue() string
}

// NewValueIter returns a ValueIter around the given value.
func NewValueIter(value model.Value) ValueIter {
	switch val := value.(type) {
	case model.Matrix:
		return &matrixIter{m: val}
	case model.Vector:
		return &vectorIter{v: val}
	case *model.String:
		return &stringIter{s: val}
	case *model.Scalar:
		return &scalarIter{s: val}
	default:
		return &emptyIter{}
	}
}

type emptyIter struct{}

func (iter *emptyIter) Next() bool {
	return false
}

func (iter *emptyIter) Metric() model.Metric {
	return model.Metric{}
}

func (iter *emptyIter) Timestamp() time.Time {
	return time.Time{}
}

func (iter *emptyIter) FloatValue() float64 {
	return 0
}

func (iter *emptyIter) StringValue() string {
	return ""
}

type matrixIter struct {
	m               model.Matrix
	i, j            int
	curSampleStream *model.SampleStream
	curSamplePair   model.SamplePair
}

func (iter *matrixIter) Next() bool {
	for iter.i < len(iter.m) {
		iter.curSampleStream = iter.m[iter.i]
		if iter.j >= len(iter.curSampleStream.Values) {
			// We're past the end of the current sample stream, move to the next
			iter.i++
			iter.j = 0
			continue
		}

		iter.curSamplePair = iter.curSampleStream.Values[iter.j]
		iter.j++
		return true
	}

	// We're out of sample streams
	return false
}

func (iter *matrixIter) Metric() model.Metric {
	return iter.curSampleStream.Metric
}

func (iter *matrixIter) Timestamp() time.Time {
	return iter.curSamplePair.Timestamp.Time()
}

func (iter *matrixIter) FloatValue() float64 {
	return float64(iter.curSamplePair.Value)
}

func (iter *matrixIter) StringValue() string {
	return iter.curSamplePair.Value.String()
}

type vectorIter struct {
	v         model.Vector
	i         int
	curSample *model.Sample
}

func (iter *vectorIter) Next() bool {
	if iter.i >= len(iter.v) {
		return false
	}

	iter.curSample = iter.v[iter.i]
	iter.i++
	return true
}

func (iter *vectorIter) Metric() model.Metric {
	return iter.curSample.Metric
}

func (iter *vectorIter) Timestamp() time.Time {
	return iter.curSample.Timestamp.Time()
}

func (iter *vectorIter) FloatValue() float64 {
	return float64(iter.curSample.Value)
}

func (iter *vectorIter) StringValue() string {
	return iter.curSample.Value.String()
}

type scalarIter struct {
	s        *model.Scalar
	consumed bool
}

func (iter *scalarIter) Next() bool {
	if !iter.consumed {
		iter.consumed = true
		return true
	}

	return false
}

func (iter *scalarIter) Metric() model.Metric {
	return model.Metric{}
}

func (iter *scalarIter) Timestamp() time.Time {
	return iter.s.Timestamp.Time()
}

func (iter *scalarIter) FloatValue() float64 {
	return float64(iter.s.Value)
}

func (iter *scalarIter) StringValue() string {
	return iter.s.Value.String()
}

type stringIter struct {
	s        *model.String
	consumed bool
}

func (iter *stringIter) Next() bool {
	if !iter.consumed {
		iter.consumed = true
		return true
	}

	return false
}

func (iter *stringIter) Metric() model.Metric {
	return model.Metric{}
}

func (iter *stringIter) Timestamp() time.Time {
	return iter.s.Timestamp.Time()
}

func (iter *stringIter) FloatValue() float64 {
	return 0
}

func (iter *stringIter) StringValue() string {
	return iter.s.Value
}

var (
	_ ValueIter = &matrixIter{}
	_ ValueIter = &scalarIter{}
	_ ValueIter = &vectorIter{}
	_ ValueIter = &stringIter{}
	_ ValueIter = &emptyIter{}
)
