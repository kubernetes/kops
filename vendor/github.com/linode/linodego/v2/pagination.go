package linodego

/**
 * Pagination and Filtering types and helpers
 */

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// PageOptions are the pagination parameters for List endpoints
type PageOptions struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	Results int `json:"results"`
}

// ListOptions are the pagination and filtering (TODO) parameters for endpoints
// nolint
type ListOptions struct {
	*PageOptions
	PageSize int    `json:"page_size"`
	Filter   string `json:"filter"`

	// QueryParams allows for specifying custom query parameters on list endpoint
	// calls. QueryParams should be an instance of a struct containing fields with
	// the `query` tag.
	QueryParams any
}

// NewListOptions simplified construction of ListOptions using only
// the two writable properties, Page and Filter
func NewListOptions(page int, filter string) *ListOptions {
	return &ListOptions{PageOptions: &PageOptions{Page: page}, Filter: filter}
}

// Hash returns the sha256 hash of the provided ListOptions.
// This is necessary for caching purposes.
func (l ListOptions) Hash() (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return "", fmt.Errorf("failed to cache ListOptions: %w", err)
	}

	h := sha256.New()

	h.Write(data)

	return hex.EncodeToString(h.Sum(nil)), nil
}

func createListOptionsToRequestMutator(opts *ListOptions) func(*http.Request) error {
	if opts == nil {
		return nil
	}

	// Return a mutator to apply query parameters and headers
	return func(req *http.Request) error {
		query := req.URL.Query()

		// Apply QueryParams from ListOptions if present
		if opts.QueryParams != nil {
			params, err := flattenQueryStruct(opts.QueryParams)
			if err != nil {
				return fmt.Errorf("failed to apply list options: %w", err)
			}

			for key, value := range params {
				query.Set(key, value)
			}
		}

		// Apply pagination options
		if opts.PageOptions != nil && opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}

		if opts.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(opts.PageSize))
		}

		// Apply filters as headers
		if len(opts.Filter) > 0 {
			req.Header.Set("X-Filter", opts.Filter)
		}

		// Assign the updated query back to the request URL
		req.URL.RawQuery = query.Encode()

		return nil
	}
}

type PagedResponse interface {
	endpoint(...any) string
	castResult(*http.Request, string) (int, int, error)
}

// flattenQueryStruct flattens a structure into a Resty-compatible query param map.
// Fields are mapped using the `query` struct tag.
func flattenQueryStruct(val any) (map[string]string, error) {
	result := make(map[string]string)

	reflectVal := reflect.ValueOf(val)

	// Deref pointer if necessary
	if reflectVal.Kind() == reflect.Pointer {
		if reflectVal.IsNil() {
			return nil, fmt.Errorf("QueryParams is a nil pointer")
		}

		reflectVal = reflect.Indirect(reflectVal)
	}

	if reflectVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"expected struct type for the QueryParams but got: %s",
			reflectVal.Kind().String(),
		)
	}

	valType := reflectVal.Type()

	for i := range valType.NumField() {
		currentField := valType.Field(i)

		queryTag, ok := currentField.Tag.Lookup("query")
		// Skip untagged fields
		if !ok {
			continue
		}

		valField := reflectVal.FieldByName(currentField.Name)
		if !valField.IsValid() {
			return nil, fmt.Errorf("invalid query param tag: %s", currentField.Name)
		}

		// Skip if it's a zero value
		if valField.IsZero() {
			continue
		}

		// Deref the pointer is necessary
		if valField.Kind() == reflect.Pointer {
			valField = reflect.Indirect(valField)
		}

		fieldString, err := queryFieldToString(valField)
		if err != nil {
			return nil, err
		}

		result[queryTag] = fieldString
	}

	return result, nil
}

func queryFieldToString(value reflect.Value) (string, error) {
	switch value.Kind() {
	case reflect.String:
		return value.String(), nil
	case reflect.Int64, reflect.Int32, reflect.Int:
		return strconv.FormatInt(value.Int(), 10), nil
	case reflect.Bool:
		return strconv.FormatBool(value.Bool()), nil
	default:
		return "", fmt.Errorf("unsupported query param type: %s", value.Type().Name())
	}
}
