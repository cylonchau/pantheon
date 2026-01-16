package target

import (
	metav1 "github.com/cylonchau/pantheon/pkg/api/meta/v1"
)

type Target struct {
	*metav1.TypeMeta `form:"kind,omitempty" json:"kind,omitempty" yaml:"kind,omitempty"`
	Addresses        []string          `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Targets          []TargetItem      `form:"targets" json:"targets" yaml:"targets"`
	InstanceSelector map[string]string `json:"selectors" yaml:"selectors" binding:"required"`
}

type TargetItem struct {
	Address       string            `form:"address" json:"address" yaml:"address" binding:"required"`
	MetricPath    string            `form:"metric_path,default=/metrics" json:"metric_path,default=/metrics" yaml:"metric_path,omitempty"`
	ScrapeTime    int               `form:"scrape_time,default=30" json:"scrape_time,default=30" yaml:"scrape_time,omitempty" `
	ScrapeTimeout int               `form:"scrape_timeout,default=10" json:"scrape_timeout,omitempty" yaml:"scrape_timeout,omitempty"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" form:"labels,omitempty"`
	Params        map[string]string `json:"params,omitempty" yaml:"params,omitempty" form:"params,omitempty"`
	Auth          *TargetAuth       `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type TargetAuth struct {
	Base        string `form:"base" json:"base,omitempty" yaml:"base,omitempty"`
	BearerToken string `form:"bearer_token" json:"bearer_token,omitempty" yaml:"bearer_token,omitempty"`
}

type TargetList struct {
	ID               uint              `form:"id" json:"id" yaml:"id"`
	Address          string            `form:"address" json:"address,omitempty" yaml:"address" binding:"required"`
	MetricPath       string            `form:"metric_path,default=/metrics" json:"metric_path,default=/metrics,omitempty" yaml:"metric_path"`
	ScrapeTime       int               `form:"scrap_time,default=30" json:"scrape_time,default=30,omitempty" yaml:"scrap_time" `
	ScrapeTimeout    int               `form:"scrape_timeout,default=10" json:"scrape_timeout,default=10,omitempty" yaml:"scrap_timeout"`
	Labels           map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" form:"labels,omitempty"`
	Params           map[string]string `json:"params,omitempty" yaml:"params,omitempty" form:"params,omitempty"`
	InstanceSelector map[string]string `json:"instanceSelector,omitempty"`
	LabelsString     string            `json:"labels_string,omitempty"`
	ParamsString     string            `json:"params_string,omitempty"`
	SelectorsString  string            `json:"selectors_string,omitempty"`
	Auth             *TargetAuth       `json:"auth,omitempty"`
}

type TargetChg struct {
	Address       string      `form:"address" json:"address,omitempty" yaml:"address"`
	MetricPath    string      `form:"metric_path,default=/metrics" json:"metric_path,default=/metrics,omitempty" yaml:"metric_path"`
	ScrapeTime    int         `form:"scrap_time,default=30" json:"scrape_time,default=30,omitempty" yaml:"scrap_time" `
	ScrapeTimeout int         `form:"scrape_timeout,default=10" json:"scrape_timeout,default=10,omitempty" yaml:"scrap_timeout"`
	Auth          *TargetAuth `json:"auth,omitempty"`
}
