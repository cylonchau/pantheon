package model

var selector_table_name = "selectors"

type Selector struct {
	ID      uint     `gorm:"primarykey"`
	Key     string   `json:"key" gorm:"index;type:varchar(255)"`
	Value   string   `json:"value" gorm:"index;type:varchar(255)"`
	Targets []Target `gorm:"many2many:target_selectors;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SelectorList struct {
	Key   string `json:"key" gorm:"index;type:varchar(255)"`
	Value string `json:"value" gorm:"index;type:varchar(255)"`
}

func (*SelectorList) TableName() string {
	return selector_table_name
}

// CreateSelectors 创建 Selector
func CreateSelectors(selectors map[string]string) ([]Selector, error) {
	var createdSelectors []Selector
	for key, value := range selectors {
		var selector Selector
		result := DB.Where(Label{Key: key, Value: value}).FirstOrCreate(&selector)
		if result.Error != nil {
			return nil, result.Error
		}
		createdSelectors = append(createdSelectors, selector)
	}

	return createdSelectors, nil
}

// ListSelector 查询所有的 Selector
func ListSelector() (selectors []SelectorList, encounterError error) {
	selectors = make([]SelectorList, 0)
	encounterError = DB.Find(&selectors).Error
	return selectors, nil
}

// UpdateSelectorByKeyValue 更新指定 Key 和 Value 的 Selector
func UpdateSelectorByKeyValue(oldKey, oldValue, newKey, newValue string) (encounterError error) {
	var selector Selector

	// 查找现有的 Selector
	if encounterError = DB.Where("`key` = ? AND `value` = ?", oldKey, oldValue).First(&selector).Error; encounterError == nil {
		// 更新 Key 和 Value
		selector.Key = newKey
		selector.Value = newValue
		// 保存更新
		encounterError = DB.Save(&selector).Error
	}
	return encounterError
}
