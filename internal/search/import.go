package search

import (
	"fmt"
	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/warmans/audio-search-bot/internal/model"
	"github.com/warmans/audio-search-bot/internal/search/mapping"
	searchModel "github.com/warmans/audio-search-bot/internal/search/model"
	"time"
)

func getMappedField(fieldName string, t mapping.FieldType, d searchModel.DialogDocument) (bluge.Field, bool) {
	switch t {
	case mapping.FieldTypeKeyword:
		return bluge.NewKeywordField(fieldName, d.GetNamedField(fieldName).(string)).StoreValue().Aggregatable().StoreValue(), true
	case mapping.FieldTypeDate:
		dateField := d.GetNamedField(fieldName).(*time.Time)
		if dateField == nil {
			return nil, false
		}
		return bluge.NewDateTimeField(fieldName, *dateField).Aggregatable().StoreValue(), true
	case mapping.FieldTypeNumber:
		switch typed := d.GetNamedField(fieldName).(type) {
		case float32:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case float64:
			return bluge.NewNumericField(fieldName, typed).StoreValue(), true
		case int32:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case int64:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case int:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case int8:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case uint8:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		case int16:
			return bluge.NewNumericField(fieldName, float64(typed)).StoreValue(), true
		default:
			panic("non-numeric type mapped as number")
		}

	case mapping.FieldTypeShingles:
		shingleAnalyzer := &analysis.Analyzer{
			Tokenizer: tokenizer.NewUnicodeTokenizer(),
			TokenFilters: []analysis.TokenFilter{
				token.NewNgramFilter(2, 16),
			},
		}
		return bluge.NewTextField(fieldName, fmt.Sprintf("%v", d.GetNamedField(fieldName))).WithAnalyzer(shingleAnalyzer).SearchTermPositions().StoreValue(), true
	}
	// just use text for everything else
	return bluge.NewTextField(fieldName, fmt.Sprintf("%v", d.GetNamedField(fieldName))).SearchTermPositions().StoreValue(), true
}

func DocumentsFromModel(episode *model.Audio) []searchModel.DialogDocument {
	docs := []searchModel.DialogDocument{}
	for _, v := range episode.Dialog {
		docs = append(docs, searchModel.DialogDocument{
			ID:             fmt.Sprintf("%s-%d", episode.ID(), v.Pos),
			Pos:            int32(v.Pos),
			MediaID:        episode.ID(),
			Publication:    episode.Publication,
			Series:         episode.Series,
			Episode:        episode.Episode,
			StartTimestamp: v.StartTimestamp.Milliseconds(),
			EndTimestamp:   v.EndTimestamp.Milliseconds(),
			MediaFileName:  episode.MediaFile,
			Content:        v.Content,
		})
	}
	return docs
}

func AddDocsToIndex(docs []searchModel.DialogDocument, writer *bluge.Writer) error {
	batch := bluge.NewBatch()
	for _, d := range docs {
		doc := bluge.NewDocument(d.ID)
		for k, t := range d.FieldMapping() {
			if mapped, ok := getMappedField(k, t, d); ok {
				doc.AddField(mapped)
			}
		}
		batch.Delete(doc.ID())
		batch.Update(doc.ID(), doc)
	}
	return writer.Batch(batch)
}
