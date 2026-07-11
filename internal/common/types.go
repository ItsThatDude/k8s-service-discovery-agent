package common

import (
	"fmt"
	"strings"
)

type NameSource string

const (
	NameSourceName       NameSource = "name"
	NameSourceAnnotation NameSource = "annotation"
	NameSourceLabel      NameSource = "label"
)

func (p *NameSource) String() string {
	return string(*p)
}
func (p *NameSource) Set(value string) error {
	switch NameSource(strings.ToLower(value)) {
	case
		NameSourceName, NameSourceAnnotation, NameSourceLabel:
		*p = NameSource(value)
		return nil
	default:
		return fmt.Errorf("Must be one of: name, annotation or label")
	}
}
