package model

var param_table_name = "params"

type Param struct {
	ID      uint     `gorm:"primarykey"`
	Key     string   `json:"key" gorm:"index;type:varchar(255)"`
	Value   string   `json:"value" gorm:"index;type:varchar(255)"`
	Targets []Target `gorm:"many2many:target_params;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ParamList struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

func (*ParamList) TableName() string {
	return param_table_name
}

func CreateParams(params map[string]string) ([]Param, error) {
	var createdParams []Param
	tx := DB.Begin()
	for key, value := range params {
		var param Param
		if result := DB.Where(Param{Key: key, Value: value}).FirstOrCreate(&param); result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		createdParams = append(createdParams, param)
	}
	tx.Commit()
	return createdParams, nil
}
