package query

type QueryWithID struct {
	ID uint `uri:"id" json:"id" yaml:"id" form:"id" binding:"required"`
}

type QueryWithName struct {
	Name string `uri:"name" json:"name" yaml:"name" form:"name" binding:"required"`
}

type QueryWithLabel struct {
	Key   string `uri:"key" json:"key" yaml:"key" form:"key" binding:"required"`
	Value string `uri:"value" json:"value" yaml:"value" form:"value" binding:"required"`
}

type QueryEditSelector struct {
	OldKey   string `json:"old_key" binding:"required"`   // 源键
	OldValue string `json:"old_value" binding:"required"` // 源值
	NewKey   string `json:"new_key" binding:"required"`   // 新键
	NewValue string `json:"new_value" binding:"required"` // 新值
}
