package exp

import (
	"reflect"
	"sort"

	"github.com/kovetskiy/goqu/v9/internal/util"
)

// Alternative to writing map[string]interface{}. Can be used for Inserts, Updates or Deletes
type Record map[string]interface{}

func (r Record) Cols() []string {
	cols := make([]string, 0, len(r))
	for col := range r {
		cols = append(cols, col)
	}
	sort.Strings(cols)
	return cols
}

func NewRecordFromStruct(i interface{}, forInsert, forUpdate bool) (r Record, err error) {
	value := reflect.ValueOf(i)
	if value.IsValid() {
		cm, err := util.GetColumnMap(value.Interface())
		if err != nil {
			return nil, err
		}
		cols := cm.Cols()
		r = make(map[string]interface{})
		for _, col := range cols {
			f := cm[col]
			switch {
			case forInsert:
				if f.ShouldInsert {
					addFieldToRecord(r, value, f)
				}
			case forUpdate:
				if f.ShouldUpdate {
					addFieldToRecord(r, value, f)
				}
			default:
				addFieldToRecord(r, value, f)
			}
		}
	}
	return
}

func addFieldToRecord(r Record, val reflect.Value, f util.ColumnData) {
	v, isAvailable := util.SafeGetFieldByIndex(val, f.FieldIndex)
	if !isAvailable {
		return
	}
	switch {
	case f.DefaultIfEmpty && util.IsEmptyValue(v):
		r[f.ColumnName] = Default()
	case f.Omitempty && util.IsPointer(v.Kind()) && v.IsNil():
		// omit the value
		return
	case v.IsValid():
		r[f.ColumnName] = v.Interface()
	default:
		r[f.ColumnName] = reflect.Zero(f.GoType).Interface()
	}
}
