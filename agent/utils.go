package agent

import (
	"fmt"

	"k8s.io/client-go/util/jsonpath"
	"k8s.io/kubectl/pkg/cmd/get"
)

func getFeilds(jsonPath string, resource interface{}) ([]string, error) {
	fields, err := get.RelaxedJSONPathExpression(jsonPath)
	if err != nil {
		return nil, err
	}
	j := jsonpath.New("jsonpath")
	if err := j.Parse(fields); err != nil {
		return nil, err
	}
	values, err := j.FindResults(resource)
	if err != nil {
		return nil, err
	}
	valueStrings := []string{}
	for arrIx := range values {
		for valIx := range values[arrIx] {
			valueStrings = append(valueStrings, fmt.Sprintf("%v", values[arrIx][valIx].Interface()))
		}
	}
	return valueStrings, nil
}
