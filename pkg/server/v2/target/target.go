package target

import (
	"github.com/gin-gonic/gin"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/model"
)

type TargetHanderV2 struct{}

func (t *TargetHanderV2) RegisterTargetAPI(g *gin.RouterGroup) {
	targetGroup := g.Group("/targets")
	targetGroup.GET("/selector/:key/:value", t.listTargetWithSeletor)

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
// @Router /ph/v2/targets/selector/{key}/{value} [get]
func (t *TargetHanderV2) listTargetWithSeletor(c *gin.Context) {
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
