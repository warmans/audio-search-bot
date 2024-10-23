package bluge_query

import (
	"fmt"
	"github.com/blugelabs/bluge"
	"github.com/warmans/audio-search-bot/internal/search/mapping"
	"github.com/warmans/audio-search-bot/internal/search/model"
	"github.com/warmans/audio-search-bot/internal/searchterms"
	"math"
	"strings"
	"time"
)

func NewBlugeQuery(terms []searchterms.Term) (bluge.Query, *int64, error) {

	// the paging/offset is included in the filter string but is not a filter so it needs to be
	// extracted.
	filteredTerms, offset := searchterms.ExtractOffset(terms)

	q := &BlugeQuery{q: bluge.NewBooleanQuery()}
	for _, v := range filteredTerms {
		if err := q.And(v); err != nil {
			return nil, nil, err
		}
	}
	return q.q, offset, nil
}

type BlugeQuery struct {
	q *bluge.BooleanQuery
}

func (j *BlugeQuery) And(term searchterms.Term) error {
	cond, err := j.condition(term.Field, term.Op, term.Value)
	if err != nil {
		return err
	}
	j.q.AddMust(cond)
	return nil
}

func (j *BlugeQuery) condition(field string, op searchterms.CompOp, value searchterms.Value) (bluge.Query, error) {

	switch op {
	case searchterms.CompOpEq:
		return j.eqFilter(field, value)
	case searchterms.CompOpNeq:
		q, err := j.eqFilter(field, value)
		if err != nil {
			return nil, err
		}
		return bluge.NewBooleanQuery().AddMustNot(q), nil
	case searchterms.CompOpLike:
		q := bluge.NewMatchQuery(stripQuotes(value.String()))
		q.SetField(field)
		q.SetFuzziness(0)
		return q, nil
	case searchterms.CompOpFuzzyLike:
		q := bluge.NewMatchQuery(stripQuotes(value.String()))
		q.SetField(field)
		q.SetFuzziness(1)
		return q, nil
	case searchterms.CompOpGt:
		switch value.Type() {
		case searchterms.IntType:
			// is max always required?
			q := bluge.NewNumericRangeQuery(float64(value.Value().(int64)), math.MaxFloat64)
			q.SetField(field)
			return q, nil
		case searchterms.DurationType:
			q := bluge.NewNumericRangeQuery(float64(value.Value().(time.Duration).Milliseconds()), math.MaxFloat64)
			q.SetField(field)
			return q, nil
		case searchterms.StringType:
			// todo: how to handle dates? they don't have a special type so we would need to look
			// at the document mapping
			q := bluge.NewTermRangeQuery(stripQuotes(value.String()), "")
			q.SetField(field)
			return q, nil
		default:
			return nil, fmt.Errorf("value type %s is not applicable to %s operation", string(value.Type()), string(op))
		}
	case searchterms.CompOpLt:
		switch value.Type() {
		case searchterms.IntType:
			q := bluge.NewNumericRangeQuery(0-math.MaxFloat64, float64(value.Value().(int64)))
			q.SetField(field)
			return q, nil
		case searchterms.DurationType:
			q := bluge.NewNumericRangeQuery(0-math.MaxFloat64, float64(value.Value().(time.Duration).Milliseconds()))
			q.SetField(field)
			return q, nil
		case searchterms.StringType:
			// todo: how to handle dates? they don't have a special type so we would need to look
			// at the bleve mapping
			q := bluge.NewTermRangeQuery("", stripQuotes(value.String()))
			q.SetField(field)
			return q, nil
		default:
			return nil, fmt.Errorf("value type %s is not applicable to %s operation", string(value.Type()), string(op))
		}
	case searchterms.CompOpGe:
		switch value.Type() {
		case searchterms.IntType:
			q := bluge.NewNumericRangeInclusiveQuery(float64(value.Value().(int64)), math.MaxFloat64, true, true)
			q.SetField(field)
			return q, nil
		case searchterms.DurationType:
			q := bluge.NewNumericRangeInclusiveQuery(float64(value.Value().(time.Duration).Milliseconds()), math.MaxFloat64, true, true)
			q.SetField(field)
			return q, nil
		case searchterms.StringType:
			// todo: how to handle dates? they don't have a special type so we would need to look
			// at the mapping
			q := bluge.NewTermRangeInclusiveQuery(stripQuotes(value.String()), "", true, true)
			q.SetField(field)
			return q, nil
		default:
			return nil, fmt.Errorf("value type %s is not applicable to %s operation", string(value.Type()), string(op))
		}
	case searchterms.CompOpLe:
		switch value.Type() {
		case searchterms.IntType:
			q := bluge.NewNumericRangeInclusiveQuery(0-math.MaxFloat64, float64(value.Value().(int64)), true, true)
			q.SetField(field)
			return q, nil
		case searchterms.DurationType:
			q := bluge.NewNumericRangeInclusiveQuery(0-math.MaxFloat64, float64(value.Value().(time.Duration).Milliseconds()), true, true)
			q.SetField(field)
			return q, nil
		case searchterms.StringType:
			// todo: how to handle dates? they don't have a special type so we would need to look
			// at the bleve mapping
			q := bluge.NewTermRangeInclusiveQuery("", stripQuotes(value.String()), true, true)
			q.SetField(field)
			return q, nil
		default:
			return nil, fmt.Errorf("value type %s is not applicable to %s operation", string(value.Type()), string(op))
		}
	default:
		return nil, fmt.Errorf("operation %s was not implemented", string(op))
	}
}

func (j *BlugeQuery) eqFilter(field string, value searchterms.Value) (bluge.Query, error) {
	fieldMap := (&model.DialogDocument{}).FieldMapping()
	t, ok := fieldMap[field]
	if ok {
		switch t {
		case mapping.FieldTypeText:
			if value.Type() != searchterms.StringType {
				return nil, fmt.Errorf("could not compare text field %s with %s", field, value.Type())
			}
			q := bluge.NewMatchPhraseQuery(stripQuotes(value.String()))
			q.SetField(field)
			return q, nil
		case mapping.FieldTypeKeyword, mapping.FieldTypeShingles:
			if value.Type() != searchterms.StringType {
				return nil, fmt.Errorf("could not compare keyword field %s with %s", field, value.Type())
			}
			q := bluge.NewTermQuery(stripQuotes(value.String()))
			q.SetField(field)
			return q, nil
		case mapping.FieldTypeNumber:
			switch value.Type() {
			case searchterms.IntType:
				q := bluge.NewNumericRangeInclusiveQuery(float64(value.Value().(int64)), float64(value.Value().(int64)), true, true)
				q.SetField(field)
				return q, nil
			case searchterms.DurationType:
				q := bluge.NewNumericRangeInclusiveQuery(float64(value.Value().(time.Duration).Milliseconds()), float64(value.Value().(time.Duration).Milliseconds()), true, true)
				q.SetField(field)
				return q, nil
			default:
				return nil, fmt.Errorf("cannot compare number to %s", value.Type())
			}
		case mapping.FieldTypeDate:
			if v, ok := value.Value().(string); ok {
				ts, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %s as date: %s", field, err.Error())
				}
				q := bluge.NewDateRangeQuery(ts, ts)
				q.SetField(field)
				return q, nil
			}
			return nil, fmt.Errorf("non-string value given as date")
		}
	}
	return nil, fmt.Errorf("unknown field type %v", t)
}

func stripQuotes(v string) string {
	return strings.Trim(v, `"`)
}
