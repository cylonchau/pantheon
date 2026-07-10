package model

import (
	"strings"
)

var monitorRuleTableName = "monitor_rules"

type MonitorRule struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	Name           string `gorm:"index;type:varchar(255)" json:"name"`
	Type           string `gorm:"type:varchar(20)" json:"type"`             // "pod" or "service"
	Namespace      string `gorm:"type:varchar(255)" json:"namespace"`       // K8s namespace
	SelectorString string `gorm:"type:text" json:"selector"`               // Comma-separated app=portal,env=prod
	PortName       string `gorm:"type:varchar(50)" json:"port_name"`        // Target port name or port number
	MetricPath     string `gorm:"type:varchar(255)" json:"metric_path"`     // e.g. /metrics
	LabelsString   string `gorm:"type:text" json:"labels"`                 // Labels to inject: k1=v1,k2=v2
	DropMetrics    string `gorm:"type:text" json:"drop_metrics"`           // Regex to drop metrics
}

func (*MonitorRule) TableName() string {
	return monitorRuleTableName
}

func CreateMonitorRule(rule *MonitorRule) error {
	rule.Name = strings.TrimSpace(rule.Name)
	return DB.Create(rule).Error
}

func GetMonitorRules() (rules []MonitorRule, err error) {
	err = DB.Find(&rules).Error
	return
}

func GetMonitorRuleByID(id uint) (rule MonitorRule, err error) {
	err = DB.Where("id = ?", id).First(&rule).Error
	return
}

func UpdateMonitorRule(id uint, updated *MonitorRule) error {
	var rule MonitorRule
	if err := DB.Where("id = ?", id).First(&rule).Error; err != nil {
		return err
	}

	// Update fields
	rule.Name = strings.TrimSpace(updated.Name)
	rule.Type = updated.Type
	rule.Namespace = updated.Namespace
	rule.SelectorString = updated.SelectorString
	rule.PortName = updated.PortName
	rule.MetricPath = updated.MetricPath
	rule.LabelsString = updated.LabelsString
	rule.DropMetrics = updated.DropMetrics

	return DB.Save(&rule).Error
}

func DeleteMonitorRule(id uint) error {
	return DB.Where("id = ?", id).Delete(&MonitorRule{}).Error
}
