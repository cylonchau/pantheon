package monitor

type MonitorRuleReq struct {
	Name           string `json:"name" binding:"required"`
	Type           string `json:"type" binding:"required,oneof=pod service"`
	Namespace      string `json:"namespace" binding:"required"`
	SelectorString string `json:"selector" binding:"required"`
	PortName       string `json:"port_name" binding:"required"`
	MetricPath     string `json:"metric_path"` // e.g. /metrics (optional, defaults to /metrics)
	LabelsString   string `json:"labels"`      // Comma-separated: env=prod,owner=sre (optional)
	DropMetrics    string `json:"drop_metrics"` // Optional regex to drop metrics
}
