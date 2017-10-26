package bdb

import (
	"fmt"
	"strings"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

// StringContains checks if the given field value contains any of the text in contains
func StringContains(field, contains string) q.Matcher {
	return q.NewFieldMatcher(field, &strContainsMatcher{contains: contains})
}

type strContainsMatcher struct {
	contains string
}

func (s *strContainsMatcher) MatchField(v interface{}) (bool, error) {
	switch fieldValue := v.(type) {
	case string:
		return strings.Contains(fieldValue, s.contains), nil
	case *string:
		if fieldValue == nil {
			return false, storm.ErrNilParam
		}
		return strings.Contains(*fieldValue, s.contains), nil
	case storage.TaskStatus:
		return strings.Contains(string(fieldValue), s.contains), nil
	default:
		return false, fmt.Errorf("Only string is supported for StringContains matcher, got %T", fieldValue)
	}
}
