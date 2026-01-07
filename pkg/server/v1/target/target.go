package target

import (
	"github.com/gin-gonic/gin"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/model"
)

type TargetHanderV1 struct{}

func (t *TargetHanderV1) RegisterTargetAPI(g *gin.RouterGroup) {
	targetGroup := g.Group("/targets")
	targetGroup.GET("/cmd/:key/:value", t.listTargetByCmd)
	targetGroup.GET("/selector/:key/:value", t.listTargetWithSeletor)
	targetGroup.GET("/:id", t.getTargetOne)
	targetGroup.PUT("", t.createTargets)
	targetGroup.POST("/:id", t.changeTargetWithID)
	targetGroup.DELETE("", t.deleteTarget)
	targetGroup.DELETE("/name/:name", t.deleteTargetWithName)
	targetGroup.DELETE("/:id", t.deleteTargetWithID)
	targetGroup.DELETE("/label/:key/:value", t.deleteTargetWithLabel)
	targetGroup.DELETE("/clean", t.cleanDeletedTargets)
}

// listTargetWithSeletor godoc
// @Summary List target with instance labels
// @Description List target with instance labels
// @Tags Targets
// @Accept json
// @Produce json
// @Param key path string true "selector key name"
// @Param value path string true "selector value name"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/selector/{key}/{value} [get]
func (t *TargetHanderV1) listTargetWithSeletor(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &query.QueryWithLabel{}
	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	if targetMap, enconterError := model.ListTargetWithSelector(targetQuery); enconterError == nil {
		query.RawSuccessResponse(c, targetMap)
		return
	}
	query.RawSuccessResponse(c, nil)
}

// getOne godoc
// @Summary Get a target by ID
// @Description Retrieve a target using its ID
// @Tags Targets
// @Accept json
// @Produce json
// @Param id path string true "Target ID"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} model.TargetRaw // 假设 model.Target 是目标的结构体
// @Router /ph/v1/targets/{id} [get]
func (t *TargetHanderV1) getTargetOne(c *gin.Context) {
	var enconterError error
	targetQuery := &query.QueryWithID{}
	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	target, enconterError := model.GetTargetByID(targetQuery.ID) // 假设有这个函数
	if enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	query.RawSuccessResponse(c, target)
}

// listTargetByCmd godoc
// @Summary List target with instance labels
// @Description List target with instance labels
// @Tags Targets
// @Accept json
// @Produce json
// @Param key path string true "label key name"
// @Param value path string true "label value name"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/cmd/{key}/{value} [get]
func (t *TargetHanderV1) listTargetByCmd(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &query.QueryWithLabel{}
	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	if targetMap, enconterError := model.ListTargetWithCtl(targetQuery); enconterError == nil {
		query.RawSuccessResponse(c, targetMap)
		return
	}
	query.RawSuccessResponse(c, nil)
}

// createTargets godoc
// @Summary Create prometheus target.
// @Description Create prometheus target.
// @Tags Targets
// @Accept json
// @Produce json
// @Param query body target.Target false "body"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets [PUT]
func (t *TargetHanderV1) createTargets(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &target.Target{}
	if enconterError = c.Bind(&targetQuery); enconterError != nil {
		query.API500Response(c, enconterError)
		return
	}

	if enconterError = model.CreateTargets(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	query.SuccessResponse(c, query.OK, nil)
}

// createTargets godoc
// @Summary Create prometheus target.
// @Description Create prometheus target.
// @Tags Targets
// @Accept json
// @Produce json
// @Param query body target.Target false "body"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets [DELETE]
func (t *TargetHanderV1) deleteTarget(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &target.Target{}

	if enconterError = c.Bind(targetQuery); enconterError != nil {
		query.API500Response(c, enconterError)
		return
	}
	//if labels, enconterError := model.GetLabelsWithLabels(targetQuery.Labels); enconterError == nil {
	if enconterError := model.DeleteTargets(targetQuery); enconterError != nil {
		query.APIResponse(c, enconterError, nil)
		return
	}
	//}
	query.SuccessResponse(c, query.OK, nil)
}

// deleteInstanceWithName godoc
// @Summary Remove prometheus target with name.
// @Description Remove prometheus target with name.
// @Tags Targets
// @Accept x-www-form-urlencoded
// @Param name path string true "target name"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/name/{name} [DELETE]
func (t *TargetHanderV1) deleteTargetWithName(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &query.QueryWithName{}

	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}
	if enconterError = model.DeleteTargetWithName(targetQuery.Name); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}

// deleteTargetWithID godoc
// @Summary Remove prometheus target with target id.
// @Description Remove prometheus target target id.
// @Tags Targets
// @Accept x-www-form-urlencoded
// @Param name path int true "target id"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/{id} [DELETE]
func (t *TargetHanderV1) deleteTargetWithID(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &query.QueryWithID{}

	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}
	if enconterError = model.DeleteTargetWithID(targetQuery.ID); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}

// cleanDeletedTargets godoc
// @Summary Clean all targets marked as deleted.
// @Description Remove all targets where is_del = 1.
// @Tags Targets
// @Accept x-www-form-urlencoded
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/clean [DELETE]
func (t *TargetHanderV1) cleanDeletedTargets(c *gin.Context) {
	// 调用模型层的清理函数
	if err := model.CleanMarkAsDeleted(); err != nil {
		query.API400Response(c, err)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}

// changeTargetWithID godoc
// @Summary Update prometheus target with target id.
// @Description Update prometheus instance target id with provided parameters.
// @Tags Targets
// @Accept json
// @Param id path int true "target id"
// @Param target body target.TargetChg true "Target update information"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/{id} [post]
func (t *TargetHanderV1) changeTargetWithID(c *gin.Context) {
	var encounterError error
	targetQuery := &query.QueryWithID{}

	// 1. 获取 URI 参数
	if encounterError = c.ShouldBindUri(targetQuery); encounterError != nil {
		query.API400Response(c, encounterError)
		return
	}

	// 2. 获取 JSON 请求体
	updates := &target.TargetChg{}
	if encounterError = c.ShouldBindJSON(updates); encounterError != nil {
		query.API400Response(c, encounterError)
		return
	}

	if encounterError = model.ChangeTargetWithID(targetQuery.ID, updates); encounterError != nil {
		query.API400Response(c, encounterError)
		return
	}

	query.SuccessResponse(c, query.OK, nil)
}

// deleteInstanceWithName godoc
// @Summary Remove prometheus target with name.
// @Description Remove target instance with name.
// @Tags Targets
// @Accept x-www-form-urlencoded
// @Param name path string true "target name"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/targets/selector/{key}/{value} [DELETE]
func (t *TargetHanderV1) deleteTargetWithLabel(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error
	targetQuery := &query.QueryWithLabel{}
	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}

	if enconterError = c.ShouldBindUri(targetQuery); enconterError != nil {
		query.API400Response(c, enconterError)
		return
	}
	if enconterError := model.DeleteTargetWithLabel(targetQuery.Key, targetQuery.Value); enconterError != nil {
		query.RawSuccessResponse(c, enconterError)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}
