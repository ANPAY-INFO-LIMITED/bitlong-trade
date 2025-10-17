package btldb

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"trade/middleware"
)

type QueryParams map[string]interface{}

func parseTagField(tag string) string {

	tagParts := strings.Split(tag, ";")
	for _, part := range tagParts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func GenericQuery[T any](model *T, params QueryParams) ([]*T, error) {
	var results []*T

	modelType := reflect.TypeOf(model).Elem()

	query := middleware.DB.Model(model)

	for key, value := range params {
		field, ok := modelType.FieldByName(key)
		if !ok {
			log.Printf("Invalid query field: %s is not a field of the model", key)
			return nil, fmt.Errorf("invalid query field: %s", key)
		}
		columnName := parseTagField(string(field.Tag.Get("gorm")))
		if columnName == "" {

			columnName = key
		}
		query = query.Where(columnName+" = ?", value)
	}

	if err := query.Find(&results).Error; err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}

	return results, nil
}

func GenericQueryByObject[T any](condition *T) ([]*T, error) {
	var results []*T

	if err := middleware.DB.Where(condition).Find(&results).Error; err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	return results, nil
}
