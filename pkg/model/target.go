package model

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"k8s.io/klog/v2"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/config"
)

const targetTableName = "targets"

type Target struct {
	ID            uint                  `gorm:"primarykey"`
	IsDel         soft_delete.DeletedAt `gorm:"softDelete:flag"`
	Address       string                `gorm:"index;type:varchar(255)"`
	Schema        string                `gorm:"type:char(5)"`
	MetricPath    string                `gorm:"index;type:varchar(255)"`
	ScrapeTime    int                   `gorm:"index;type:int"`
	ScrapeTimeout int                   `gorm:"index;type:int"`
	BearerToken   string                `gorm:"index;type:varchar(255)"`
	BaseAuth      string                `gorm:"index;type:varchar(255)"`
	Labels        []Label               `gorm:"many2many:target_labels;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Params        []Param               `gorm:"many2many:target_params;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Selectors     []Selector            `gorm:"many2many:target_selectors;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TargetRaw struct {
	Address       string `gorm:"index;type:varchar(255)" json:"address"`
	Schema        string `gorm:"type:char(5)" json:"schema"`
	MetricPath    string `gorm:"index;type:varchar(255)" json:"metric_path"`
	ScrapeTime    int    `gorm:"index;type:int" json:"scrape_time"`
	ScrapeTimeout int    `gorm:"index;type:int" json:"scrape_timeout"`
	BearerToken   string `gorm:"index;type:varchar(255)" json:"bearer_token"`
	BaseAuth      string `gorm:"index;type:varchar(255)" json:"base_auth"`
}

type TargetList struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type tempTargetList struct {
	ID            uint   `gorm:"primarykey"`
	Address       string `gorm:"index;type:varchar(255)"`
	Schema        string `gorm:"type:char(5)"`
	MetricPath    string `gorm:"index;type:varchar(255)"`
	SelectorKey   string
	SelectorValue string
}

type SwapResult struct {
	TargetID      uint
	Address       string
	MetricPath    string
	ScrapeTime    int
	ScrapeTimeout int
	LabelKey      string
	LabelValue    string
	BearerToken   string
	BaseAuth      string
	Schema        string
}

type swapMap struct {
	ID    int
	Key   string
	Value string
}

func (*TargetList) TableName() string {
	return targetTableName
}

func (*TargetRaw) TableName() string {
	return targetTableName
}

func (t *Target) BeforeDelete(tx *gorm.DB) (err error) {

	// 找到与此 Target 相关的所有 Params
	var params []Param
	if err := tx.Model(t).Association("Params").Find(&params); err != nil {
		klog.V(4).Infof("Error fetching params: %v", err)
		return err
	}

	for _, param := range params {
		paramsCount := tx.Model(&Target{}). // 假设 Target 是关联的主模型
							Joins("JOIN target_params ON target_params.param_id = ?").
							Where("target_params.target_id != ?", param.ID, t.ID).
							Find(nil)
		if paramsCount.RowsAffected == 0 {
			if err := tx.Delete(&param).Error; err != nil {
				klog.V(4).Infof("Error deleting param:", err)
				return err
			}
		}
	}

	if err := tx.Model(t).Association("Params").Clear(); err != nil {
		klog.V(4).Infof("Error cleaning params relation: %v", err)
		return err
	}

	// 找到与此 Target 相关的所有 Labels
	var labels []Label
	if err := tx.Model(t).Association("Labels").Find(&labels); err != nil {
		klog.V(4).Infof("Error fetching labels: %v", err)
		return err
	}

	for _, label := range labels {
		labelsCount := tx.Model(&Target{}). // 假设 Target 是关联的主模型
							Joins("JOIN target_labels ON target_labels.label_id = ?").
							Where("target_labels.target_id != ?", label.ID, t.ID).
							Find(nil)
		if labelsCount.RowsAffected == 0 {
			if err := tx.Delete(&label).Error; err != nil {
				klog.V(4).Infof("Error deleting label:", err)
				return err
			}
		}
	}

	if err := tx.Model(t).Association("Labels").Clear(); err != nil {
		klog.V(4).Infof("Error cleaning labels relation: %v", err)
		return err
	}

	var selectors []Selector
	if err := tx.Model(t).Association("Selectors").Find(&selectors); err != nil {
		klog.V(4).Infof("Error fetching selectors: %v", err)
		return err
	}
	for _, selector := range selectors {
		selectorCount := tx.Model(&Target{}). // 假设 Target 是关联的主模型
							Joins("JOIN target_selectors ON target_selectors.selector_id = ?").
							Where("target_selectors.target_id != ?", selector.ID, t.ID).
							Find(nil)
		if selectorCount.RowsAffected == 0 {
			if err := tx.Delete(&selector).Error; err != nil {
				klog.V(4).Infof("Error deleting selector:", err)
				return err
			}
		}
	}
	if err := tx.Model(t).Association("Selectors").Clear(); err != nil {
		klog.V(4).Infof("Error cleaning selectors relation: %v", err)
		return err
	}

	return nil
}

func CreateTargets(target *target.Target) (encounterError error) {
	tx := DB.Begin()
	defer func() {
		if encounterError != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 创建选择器
	instanceSelectors, encounterError := CreateSelectors(target.InstanceSelector)
	if encounterError != nil {
		return encounterError
	}

	for _, targetItem := range target.Targets {
		// 默认值处理
		if targetItem.MetricPath == "" {
			targetItem.MetricPath = "/metrics"
		}
		if targetItem.ScrapeTime == 0 {
			targetItem.ScrapeTime = 30
		}
		if targetItem.ScrapeTimeout == 0 {
			targetItem.ScrapeTimeout = 10
		}

		if targetItem.ScrapeTimeout > targetItem.ScrapeTime {
			targetItem.ScrapeTimeout = targetItem.ScrapeTime
		}
		// 处理 Address 字段，分离 Schema 和 Address
		var schema string
		if strings.HasPrefix(targetItem.Address, "http://") || strings.HasPrefix(targetItem.Address, "https://") {
			schema = strings.Split(targetItem.Address, "://")[0]
			targetItem.Address = strings.Split(targetItem.Address, "://")[1]
		} else {
			schema = "http" // 默认值
		}

		// 创建新的 Target 实例
		newTarget := &Target{
			Address:       targetItem.Address,
			Schema:        schema,
			MetricPath:    targetItem.MetricPath,
			ScrapeTime:    targetItem.ScrapeTime,
			ScrapeTimeout: targetItem.ScrapeTimeout,
		}

		if targetItem.Auth != nil {
			if targetItem.Auth.BearerToken != "" {
				newTarget.BearerToken = targetItem.Auth.BearerToken
			} else if targetItem.Auth.Base != "" {
				newTarget.BaseAuth = targetItem.Auth.Base
			}
		}

		// 动态构建 Selectors 查询条件
		// selector 为全局条件，所以放置最上部
		selectorConditions := DB
		if len(target.InstanceSelector) > 0 {
			for key, value := range target.InstanceSelector {
				selectorConditions = selectorConditions.Or("selectors.key = ? AND selectors.value = ?", key, value)
			}
		}
		var existTargets []tempTargetList
		preQuery := DB.Table(targetTableName).
			Select("targets.id as id, selectors.key as `selector_key`, selectors.value as `selector_value`, targets.address, targets.metric_path, targets.schema").
			Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
			Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
			Where("targets.`is_del` = 0").
			Where("targets.address = ? AND targets.metric_path = ? AND targets.scrape_time = ? AND targets.scrape_timeout = ?",
				newTarget.Address, newTarget.MetricPath, newTarget.ScrapeTime, newTarget.ScrapeTimeout).
			Where(selectorConditions)

		encounterError = preQuery.Find(&existTargets).Error
		if encounterError != nil {
			if errors.Is(encounterError, gorm.ErrRecordNotFound) && len(existTargets) == 0 {
				encounterError = DB.Model(&Target{}).Create(&newTarget).Error
			} else {
				return encounterError
			}
		} else {
			isCreateTarget := true
			for _, existTargetItem := range existTargets {
				// 先获取相关的 params
				// 用于查找已经存在的params
				var judgeTargetsParamsRelation []swapMap
				judgeParamQuery := DB.Table(targetTableName).
					Select("targets.id as id, params.key as `key`, params.value as `value`").
					Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
					Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
					Joins("JOIN target_params ON target_params.target_id = targets.id").
					Joins("JOIN params ON params.id = target_params.param_id")

				judgeParamQuery.Where("targets.id = ?", existTargetItem.ID)

				// 拼接selector
				if len(instanceSelectors) > 0 {
					judgeParamQuery = judgeParamQuery.Where(selectorConditions)
				}

				// 获取这个target的所有params
				if encounterError = judgeParamQuery.Scan(&judgeTargetsParamsRelation).Error; encounterError != nil {
					return
				}

				// 拼接param
				paramsMap := make(map[int]map[string]string)
				for _, param := range judgeTargetsParamsRelation {
					if paramsMap[param.ID] == nil {
						paramsMap[param.ID] = make(map[string]string)
					}
					paramsMap[param.ID][param.Key] = param.Value
				}

				existJudgeParamsString := mapToURLParams(paramsMap[int(existTargetItem.ID)])
				existTargetUniqueKey := hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprintf("%s://%s%s?%s", existTargetItem.Schema, existTargetItem.Address, existTargetItem.MetricPath, existJudgeParamsString))))

				newJudgeParamsString := mapToURLParams(targetItem.Params)
				newTargetUniqueKey := hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprintf("%s://%s%s?%s", schema, targetItem.Address, targetItem.MetricPath, newJudgeParamsString))))

				if existTargetUniqueKey == newTargetUniqueKey {
					isCreateTarget = false
				}
			}
			if isCreateTarget {
				if encounterError = DB.Model(&Target{}).Create(&newTarget).Error; encounterError != nil {
					return encounterError
				}
			}
		}

		if newTarget.ID != 0 {

			if len(targetItem.Labels) > 0 {
				// 动态构建查询条件
				queryLabels := DB.Table("targets").Select("*").
					Joins("LEFT JOIN target_labels ON target_labels.target_id = targets.id").
					Joins("LEFT JOIN labels ON labels.id = target_labels.label_id").
					Joins("LEFT JOIN target_selectors ON target_selectors.target_id = targets.id").
					Joins("LEFT JOIN selectors ON selectors.id = target_selectors.selector_id").
					Where("targets.address = ? AND targets.metric_path = ? AND targets.scrape_time = ? AND targets.scrape_timeout = ?",
						newTarget.Address, newTarget.MetricPath, newTarget.ScrapeTime, newTarget.ScrapeTimeout).
					Where("targets.id = ?", newTarget.ID)

				// 检查 Labels
				if len(targetItem.Labels) > 0 {
					labelConditions := DB
					for key, value := range targetItem.Labels {
						labelConditions = labelConditions.Or("labels.key = ? AND labels.value = ?", key, value)
					}
					queryLabels = queryLabels.Where(labelConditions)
				}
				// 动态关联selector条件
				if len(target.InstanceSelector) > 0 {
					queryLabels = queryLabels.Where(selectorConditions)
				}

				// 执行查询
				// 执行查询并检查结果
				var existingTargets []Target
				// 如果找到目标，说明已存在，跳过插入
				if encounterError = queryLabels.Find(&existingTargets).Error; encounterError == nil && len(existingTargets) != len(targetItem.Labels) {
					// 关联 Labels
					// 在这里调用 CreateLabels
					var createdLabels []Label
					if createdLabels, encounterError = CreateLabels(targetItem.Labels); encounterError == nil {
						if encounterError = DB.Model(&newTarget).Association("Labels").Append(createdLabels); encounterError != nil {
							return encounterError
						}
					}
				}
			}

			if len(targetItem.Params) > 0 {
				// 动态构建查询条件
				queryParams := DB.Table(targetTableName).Select("*").
					Joins("LEFT JOIN target_params ON target_params.target_id = targets.id").
					Joins("LEFT JOIN params ON params.id = target_params.param_id").
					Joins("LEFT JOIN target_selectors ON target_selectors.target_id = targets.id").
					Joins("LEFT JOIN selectors ON selectors.id = target_selectors.selector_id").
					Where("targets.address = ? AND targets.metric_path = ? AND targets.scrape_time = ? AND targets.scrape_timeout = ?",
						newTarget.Address, newTarget.MetricPath, newTarget.ScrapeTime, newTarget.ScrapeTimeout).
					Where("targets.id = ?", newTarget.ID)

				// 动态构建 Params 查询条件
				paramConditions := DB
				for key, value := range targetItem.Params {
					paramConditions = paramConditions.Or("params.key = ? AND params.value = ?", key, value)
				}
				queryParams = queryParams.Where(paramConditions)
				// 动态构建 Selectors 查询条件
				if len(target.InstanceSelector) > 0 {
					queryParams = queryParams.Where(selectorConditions)
				}
				// 执行查询
				// 执行查询并检查结果
				var existingTargets []Target
				// 如果找到目标，说明已存在，跳过插入
				if encounterError = queryParams.Find(&existingTargets).Error; encounterError == nil && len(existingTargets) != len(targetItem.Params) {
					// 关联 Params
					// 在这里调用 CreateLabels
					var createdParams []Param
					if createdParams, encounterError = CreateParams(targetItem.Params); encounterError == nil {
						if encounterError = DB.Model(&newTarget).Association("Params").Append(createdParams); encounterError != nil {
							return encounterError
						}
					}
				}
			}
			// 关联 Selectors
			if encounterError = DB.Model(&newTarget).Association("Selectors").Append(instanceSelectors); encounterError != nil {
				return encounterError
			}
		}

	}
	return
}

func DeleteTargetWithID(id uint) (encounterError error) {
	existingTarget := &Target{}
	targetResult := DB.Model(&Target{}).Where("id = ? ", id).Find(existingTarget)
	if encounterError = targetResult.Error; encounterError == nil {
		if targetResult.RowsAffected > 0 {
			tx := DB.Begin()
			if encounterError = tx.Delete(existingTarget).Error; encounterError == nil {
				encounterError = tx.Commit().Error
			} else {
				encounterError = tx.Rollback().Error
			}
		} else {
			encounterError = fmt.Errorf("No target found with the provided id: %d", id)
		}
	}
	return encounterError
}

func CleanMarkAsDeleted() (encounterError error) {
	existingTargets := &[]Target{}
	// 查询所有标记为已删除的目标
	targetResult := DB.Model(&Target{}).Unscoped().Where("targets.`is_del` = 1").Find(existingTargets)
	if encounterError = targetResult.Error; encounterError == nil {
		if targetResult.RowsAffected > 0 {
			tx := DB.Begin()
			// 删除所有已标记为已删除的目标
			if encounterError = tx.Unscoped().Delete(existingTargets).Error; encounterError == nil {
				encounterError = tx.Commit().Error
			} else {
				_ = tx.Rollback().Error
			}
		} else {
			encounterError = fmt.Errorf("No targets marked as deleted found")
		}
	}
	return encounterError
}

func DeleteTargetWithName(targetName string) (encounterError error) {
	existingTarget := &Target{}
	targetResult := DB.Model(&Target{}).Where("address = ? ", targetName).Find(existingTarget)
	if encounterError = targetResult.Error; encounterError == nil {
		if targetResult.RowsAffected > 0 {
			tx := DB.Begin()
			if encounterError = targetResult.Delete(existingTarget).Error; encounterError == nil {
				encounterError = tx.Commit().Error
			} else {
				encounterError = tx.Rollback().Error
			}
		} else {
			encounterError = errors.New("No target found with the provided name: " + targetName)
		}
	}
	return encounterError
}

func DeleteTargetWithLabel(key, value string) (encounterError error) {
	existingTarget := []*Target{}
	targetResult := DB.Preload("Labels", "key = ? AND value = ?", key, value).Find(&existingTarget)
	if encounterError = targetResult.Error; encounterError == nil {
		if targetResult.RowsAffected > 0 {
			tx := DB.Begin()
			if encounterError = targetResult.Delete(existingTarget).Error; encounterError == nil {
				encounterError = tx.Commit().Error
			} else {
				encounterError = tx.Rollback().Error
			}
		} else {
			encounterError = fmt.Errorf("No target found with the provided label <%s>:<%s>", key, value)
		}
	}
	return encounterError
}

func DeleteTargets(target *target.Target) error {
	// 检查是否提供了删除条件
	if len(target.Targets) == 0 {
		return errors.New("no targets specified for deletion")
	}

	// 开始构建查询
	query := DB.Table("targets").Select("DISTINCT targets.id").
		Joins("LEFT JOIN target_labels ON target_labels.target_id = targets.id").
		Joins("LEFT JOIN labels ON labels.id = target_labels.label_id").
		Joins("LEFT JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("LEFT JOIN selectors ON selectors.id = target_selectors.selector_id")
	//query := DB.Table("targets").Select("DISTINCT targets.id as id").
	//	Joins("LEFT JOIN target_labels ON target_labels.target_id = targets.id").
	//	Joins("LEFT JOIN labels ON labels.id = target_labels.label_id").
	//	Joins("LEFT JOIN target_selectors ON target_selectors.target_id = targets.id").
	//	Joins("LEFT JOIN selectors ON selectors.id = target_selectors.selector_id")

	// 处理 Instance Selectors (使用 OR 逻辑)
	if len(target.InstanceSelector) > 0 {
		selectorConditions := DB
		for key, value := range target.InstanceSelector {
			selectorConditions = selectorConditions.Or("selectors.key = ? AND selectors.value = ?", key, value)
		}
		query = query.Where(selectorConditions)
	}

	// 处理 Targets
	for _, targetItem := range target.Targets {
		// 处理 Labels (使用 OR 逻辑)
		if len(targetItem.Labels) > 0 {
			labelConditions := DB
			for key, value := range targetItem.Labels {
				labelConditions = labelConditions.Or("labels.key = ? AND labels.value = ?", key, value)
			}
			query = query.Where(labelConditions)
		}

		// 动态添加其他条件
		if targetItem.Address != "" {
			query = query.Where("targets.address = ?", targetItem.Address)
		}
		if targetItem.Auth != nil {
			if targetItem.Auth.Base != "" {
				query = query.Where("targets.base_auth = ?", targetItem.Auth.Base)
			}
			if targetItem.Auth.BearerToken != "" {
				query = query.Where("targets.bearer_token = ?", targetItem.Auth.BearerToken)
			}
		}

	}

	// 打印生成的 SQL 语句
	klog.V(4).Infof("Generated SQL before execution: %s", query.Statement.SQL.String())

	// 执行查询以获取符合条件的唯一目标ID
	var targetIDs []uint
	if err := query.Pluck("DISTINCT targets.id", &targetIDs).Error; err != nil {
		return err
	}

	// 如果没有找到符合条件的目标，返回错误
	if len(targetIDs) == 0 {
		return errors.New("no targets found matching the criteria")
	}

	//开始删除符合条件的目标
	for _, id := range targetIDs {
		tx := DB.Begin()

		// 删除目标
		if err := DB.Delete(&Target{}, id).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			return err
		}
		fmt.Printf("Target with ID %d deleted successfully\n", id)
	}

	return nil
}

func ListTargetWithCtl(query *query.QueryWithLabel) (results []target.TargetList, encounterError error) {
	results = make([]target.TargetList, 0)
	// 先获取相关的 params
	var targetsParamsRelation []swapMap
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, params.key as `key`, params.value as `value`").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Joins("JOIN target_params ON target_params.target_id = targets.id").
		Joins("JOIN params ON params.id = target_params.param_id").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Where("targets.`is_del` = 0").
		Scan(&targetsParamsRelation).Error; encounterError != nil {
		return
	}

	paramsMap := make(map[int]map[string]string)
	for _, param := range targetsParamsRelation {
		if paramsMap[param.ID] == nil {
			paramsMap[param.ID] = make(map[string]string)
		}
		paramsMap[param.ID][param.Key] = param.Value
	}

	// 获取 labels 数据
	var targetsLabelsRelation []swapMap
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, labels.key as `key`, labels.value as `value`").
		Joins("JOIN target_labels ON target_labels.target_id = targets.id").
		Joins("JOIN labels ON labels.id = target_labels.label_id").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Where("targets.`is_del` = 0").
		Scan(&targetsLabelsRelation).Error; encounterError != nil {
		return
	}

	labelsMap := make(map[int]map[string]string)
	for _, param := range targetsLabelsRelation {
		if labelsMap[param.ID] == nil {
			labelsMap[param.ID] = make(map[string]string)
		}
		labelsMap[param.ID][param.Key] = param.Value
	}
	targets := []Target{}
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, targets.address, targets.schema, targets.metric_path, targets.scrape_time, targets.scrape_timeout, targets.bearer_token, targets.base_auth").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Where("targets.`is_del` = 0").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Order("id").
		Scan(&targets).Error; encounterError != nil {
		return
	}

	targetResults := map[string]target.TargetList{}
	for _, rawTarget := range targets {
		paramsString := mapToURLParams(paramsMap[int(rawTarget.ID)])
		uniqueKey := hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprintf("%://%s%s?%s", rawTarget.Schema, rawTarget.Address, rawTarget.MetricPath, paramsString))))

		targetResult := target.TargetList{
			ID:            rawTarget.ID,
			Address:       rawTarget.Schema + "://" + rawTarget.Address,
			MetricPath:    rawTarget.MetricPath,
			ScrapeTimeout: rawTarget.ScrapeTimeout,
			ScrapeTime:    rawTarget.ScrapeTime,
		}
		if rawTarget.BaseAuth != "" || rawTarget.BearerToken != "" {
			targetResult.Auth = &target.TargetAuth{}
			if rawTarget.BaseAuth != "" {
				targetResult.Auth.Base = rawTarget.BaseAuth
			}
			if rawTarget.BearerToken != "" {
				targetResult.Auth.BearerToken = rawTarget.BearerToken
			}
		}

		// 加入 labels
		if labels, exists := labelsMap[int(rawTarget.ID)]; exists {
			targetResult.Labels = labels
		}
		// 加入 params
		if params, exists := paramsMap[int(rawTarget.ID)]; exists {
			targetResult.Params = params
		}

		targetResults[uniqueKey] = targetResult

	}
	// 将聚合后的 targetMap 转换为 results 切片
	for _, result := range targetResults {
		results = append(results, result)
	}
	return results, encounterError

	return nil, encounterError
}

func ListTargetWithSelector(query *query.QueryWithLabel) (results []TargetList, encounterError error) {
	results = make([]TargetList, 0)
	// 先获取相关的 params
	var targetsParamsRelation []swapMap
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, params.key as `key`, params.value as `value`").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Joins("JOIN target_params ON target_params.target_id = targets.id").
		Joins("JOIN params ON params.id = target_params.param_id").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Where("targets.`is_del` = 0").
		Scan(&targetsParamsRelation).Error; encounterError != nil {
		return
	}

	paramsMap := make(map[int]map[string]string)
	for _, param := range targetsParamsRelation {
		if paramsMap[param.ID] == nil {
			paramsMap[param.ID] = make(map[string]string)
		}
		paramsMap[param.ID][fmt.Sprintf("__param_%s", param.Key)] = param.Value
	}

	// 获取 labels 数据
	var targetsLabelsRelation []swapMap
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, labels.key as `key`, labels.value as `value`").
		Joins("JOIN target_labels ON target_labels.target_id = targets.id").
		Joins("JOIN labels ON labels.id = target_labels.label_id").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Where("targets.`is_del` = 0").
		Scan(&targetsLabelsRelation).Error; encounterError != nil {
		return
	}

	labelsMap := make(map[int]map[string]string)
	for _, label := range targetsLabelsRelation {
		if labelsMap[label.ID] == nil {
			labelsMap[label.ID] = make(map[string]string)
		}
		labelsMap[label.ID][label.Key] = label.Value
	}

	targets := []Target{}
	if encounterError = DB.Table(targetTableName).
		Select("targets.id as id, targets.address, targets.schema, targets.metric_path, targets.scrape_time, targets.scrape_timeout, targets.bearer_token, targets.base_auth").
		Joins("JOIN target_selectors ON target_selectors.target_id = targets.id").
		Joins("JOIN selectors ON selectors.id = target_selectors.selector_id").
		Where("selectors.key = ? AND selectors.value = ?", query.Key, query.Value).
		Where("targets.`is_del` = 0").
		Scan(&targets).Error; encounterError != nil {
		return
	}
	targetResults := make(map[string]TargetList)
	for _, target := range targets {
		paramsString := mapToURLParams(paramsMap[int(target.ID)])
		uniqueKey := hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprintf("%://%s%s?%s", target.Schema, target.Address, target.MetricPath, paramsString))))

		var targetResult TargetList
		if target.BearerToken != "" || target.BaseAuth != "" {
			proxyParsedURL := parseConfigURL(config.CONFIG.ProxyAddress)

			targetResult = TargetList{
				Targets: []string{proxyParsedURL.Host},
				Labels: map[string]string{
					"instance":            target.Address,
					"__scrape_interval__": fmt.Sprintf("%ds", target.ScrapeTime),
					"__scrape_timeout__":  fmt.Sprintf("%ds", target.ScrapeTimeout),
					"__metrics_path__":    proxyParsedURL.Path,
					"__scheme__":          proxyParsedURL.Scheme,
				},
			}
		} else {
			// 创建新 TargetList
			targetResult = TargetList{
				Targets: []string{target.Address},
				Labels: map[string]string{
					"instance":            target.Address,
					"__scrape_interval__": fmt.Sprintf("%ds", target.ScrapeTime),
					"__scrape_timeout__":  fmt.Sprintf("%ds", target.ScrapeTimeout),
					"__metrics_path__":    target.MetricPath,
					"__scheme__":          target.Schema,
				},
			}
		}

		// 加入 labels
		if labels, exists := labelsMap[int(target.ID)]; exists {
			for k, v := range labels {
				targetResult.Labels[k] = v
			}
		}

		// 加入 params
		if params, exists := paramsMap[int(target.ID)]; exists {
			for k, v := range params {
				targetResult.Labels[k] = v
			}
		}

		if value, exists := targetResult.Labels["__param_target"]; exists {
			targetResult.Labels["instance"] = value
		}

		// 处理 BearerToken 和 BaseAuth
		if target.BearerToken != "" || target.BaseAuth != "" {
			if target.BearerToken != "" {
				targetResult.Labels["__param_bearer"] = target.BearerToken
			}
			if target.BaseAuth != "" {
				baseAuthEncoded := base64.StdEncoding.EncodeToString([]byte(target.BaseAuth))
				targetResult.Labels["__param_base"] = baseAuthEncoded
			}
			if target.BearerToken != "" && target.BaseAuth != "" {
				delete(targetResult.Labels, "__param_base")
			}

			// 处理 schema 和 host 和 path
			proxyHostPattern := `^(?P<host>[\w.-]+|\d{1,3}(\.\d{1,3}){3}):(?P<port>\d{1,5})$`
			proxyHostRe := regexp.MustCompile(proxyHostPattern)
			if proxyHostRe.MatchString(target.Address) {
				proxyHostParts := strings.Split(target.Address, ":")
				targetResult.Labels["__param_host"] = proxyHostParts[0]
				targetResult.Labels["__param_port"] = proxyHostParts[1]
			} else {
				targetResult.Labels["__param_host"] = target.Address
				switch target.Schema {
				case "http":
					targetResult.Labels["__param_port"] = "80" // 默认值
				case "https":
					targetResult.Labels["__param_port"] = "443" // 默认值
				default:
					targetResult.Labels["__param_port"] = "80" // 默认值
				}
			}
			targetResult.Labels["__param_path"] = target.MetricPath
			targetResult.Labels["__param_schema"] = target.Schema
		}

		// 添加到 targetMap
		targetResults[uniqueKey] = targetResult
	}

	// 将聚合后的 targetMap 转换为 results 切片
	for _, result := range targetResults {
		results = append(results, result)
	}

	return results, nil
}

func GetTargetByID(targetID uint) (target TargetRaw, encounterError error) {
	encounterError = DB.Table(targetTableName).Where("id = ?", targetID).Where("`is_del` = 0").First(&target).Error
	return
}

func ChangeTargetWithID(id uint, updates *target.TargetChg) (encounterError error) {
	existingTarget := &Target{}
	targetResult := DB.Model(&Target{}).Where("id = ?", id).First(existingTarget)
	if encounterError = targetResult.Error; encounterError == nil {
		if targetResult.RowsAffected > 0 {
			// 创建一个事务
			tx := DB.Begin()

			// 创建一个更新映射
			updateData := Target{}

			if updates.Address != "" {
				updateData.Address = updates.Address
				var schema string
				if strings.HasPrefix(updates.Address, "http://") || strings.HasPrefix(updates.Address, "https://") {
					schema = strings.Split(updates.Address, "://")[0]
					updates.Address = strings.Split(updates.Address, "://")[1]
				} else {
					schema = "http" // 默认值
				}
				updateData.Schema = schema
			}
			if updates.MetricPath != "" {
				updateData.MetricPath = updates.MetricPath
			}
			if updates.ScrapeTime > 0 {
				updateData.ScrapeTime = updates.ScrapeTime
			}
			if updates.ScrapeTimeout > 0 {
				updateData.ScrapeTimeout = updates.ScrapeTimeout
			}

			if updates.Auth != nil {
				if updates.Auth.Base != "" {
					updateData.BaseAuth = updates.Auth.Base
				}
				if updates.Auth.BearerToken != "" {
					updateData.BearerToken = updates.Auth.BearerToken
				}
			}

			// 执行更新
			if encounterError = tx.Model(existingTarget).Updates(updateData).Error; encounterError == nil {
				encounterError = tx.Commit().Error
			} else {
				tx.Rollback()
			}
		} else {
			encounterError = fmt.Errorf("No target found with the provided id: %d", id)
		}
	}
	return encounterError
}
