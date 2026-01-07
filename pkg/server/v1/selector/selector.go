package selector

import (
	"github.com/gin-gonic/gin"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/model"
)

type SelectorHanderV1 struct{}

func (t *SelectorHanderV1) RegisterSelectorAPI(g *gin.RouterGroup) {
	seletorGroup := g.Group("/selectors")
	seletorGroup.GET("", t.listSelectors)
	seletorGroup.POST("", t.updateSelector)

}

// listSelectors godoc
// @Summary List selectors
// @Description List selectors
// @Tags Selectors
// @Accept json
// @Produce json
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/selectors [get]
func (t *SelectorHanderV1) listSelectors(c *gin.Context) {
	// 1. 获取参数和参数校验
	var enconterError error

	if selectorMap, enconterError := model.ListSelector(); enconterError == nil {
		query.RawSuccessResponse(c, selectorMap)
		return
	}
	query.API400Response(c, enconterError)
}

// updateSelector godoc
// @Summary Update selector
// @Description Update selector by old key and value
// @Tags Selectors
// @Accept json
// @Produce json
// @securityDefinitions.apikey BearerAuth
// @Param request body query.QueryEditSelector true "Update Selector Request"
// @Success 200 {object} query.Response
// @Failure 400 {object} query.Response
// @Router /ph/v1/selectors [post]
func (t *SelectorHanderV1) updateSelector(c *gin.Context) {
	var request query.QueryEditSelector

	// 1. 绑定请求参数
	if err := c.ShouldBindJSON(&request); err != nil {
		query.API400Response(c, err)
		return
	}

	// 2. 调用模型层进行更新
	if err := model.UpdateSelectorByKeyValue(request.OldKey, request.OldValue, request.NewKey, request.NewValue); err != nil {
		query.API400Response(c, err)
		return
	}

	// 3. 成功返回更新后的选择器
	query.SuccessResponse(c, nil, nil)
}
