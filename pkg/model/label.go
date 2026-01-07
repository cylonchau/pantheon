package model

import "gorm.io/gorm"

var label_table_name = "labels"

type Label struct {
	ID      uint     `gorm:"primarykey"`
	Key     string   `json:"key" gorm:"index;type:varchar(255)"`
	Value   string   `json:"value" gorm:"index;type:varchar(255)"`
	Targets []Target `gorm:"many2many:target_labels;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type LabelList struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (*LabelList) TableName() string {
	return label_table_name
}

func CreateLabels(labels map[string]string) ([]Label, error) {
	var createdLabels []Label
	tx := DB.Begin()
	for key, value := range labels {
		var label Label
		if result := DB.Where(Label{Key: key, Value: value}).FirstOrCreate(&label); result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		createdLabels = append(createdLabels, label)
	}
	tx.Commit()
	return createdLabels, nil
}

func GetLabelsWithLabels(labels map[string]string) ([]Label, error) {
	var createdLabels []Label
	for key, value := range labels {
		var label Label
		result := DB.Where(Label{Key: key, Value: value}).Find(&label)
		if result.Error != nil {
			if result.Error != gorm.ErrRecordNotFound {
				continue
			}
			return nil, result.Error
		}
		createdLabels = append(createdLabels, label)
	}

	return createdLabels, nil
}
