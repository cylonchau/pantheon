package selector

import metav1 "github.com/cylonchau/pantheon/pkg/api/meta/v1"

type Selector struct {
	*metav1.TypeMeta `form:"kind,omitempty" json:"kind,omitempty" yaml:"kind,omitempty"`
	Selectors        []SelectorItem `form:"selectors" json:"selectors"`
}

type SelectorItem struct {
	Key   string `form:"key" json:"key" yaml:"key" binding:"required"`
	Value string `form:"value" json:"value" yaml:"value" binding:"required"`
}
