package monitor

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cylonchau/pantheon/pkg/api/monitor"
	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/model"
)

type MonitorHandlerV1 struct{}

func (h *MonitorHandlerV1) RegisterMonitorAPI(g *gin.RouterGroup) {
	monitorGroup := g.Group("/monitors")
	monitorGroup.GET("", h.listMonitors)
	monitorGroup.PUT("", h.createMonitor)
	monitorGroup.POST("/:id", h.updateMonitor)
	monitorGroup.DELETE("/:id", h.deleteMonitor)
}

// @Summary      List monitor rules
// @Description  Get all active monitor rules
// @Tags         monitor
// @Accept       json
// @Produce      json
// @Success      200      {array}   model.MonitorRule
// @Failure      500      {object}  query.Response
// @Router       /ph/v1/monitors [get]
func (h *MonitorHandlerV1) listMonitors(c *gin.Context) {
	rules, err := model.GetMonitorRules()
	if err != nil {
		query.API500Response(c, err)
		return
	}
	query.RawSuccessResponse(c, rules)
}

// @Summary      Create monitor rule
// @Description  Add a new monitor rule
// @Tags         monitor
// @Accept       json
// @Produce      json
// @Param        rule     body      monitor.MonitorRuleReq  true  "Monitor Rule configuration"
// @Success      200      {object}  query.Response
// @Failure      400      {object}  query.Response
// @Failure      500      {object}  query.Response
// @Router       /ph/v1/monitors [put]
func (h *MonitorHandlerV1) createMonitor(c *gin.Context) {
	var req monitor.MonitorRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		query.API400Response(c, err)
		return
	}

	if req.MetricPath == "" {
		req.MetricPath = "/metrics"
	}

	rule := &model.MonitorRule{
		Name:           req.Name,
		Type:           req.Type,
		Namespace:      req.Namespace,
		SelectorString: req.SelectorString,
		PortName:       req.PortName,
		MetricPath:     req.MetricPath,
		LabelsString:   req.LabelsString,
		DropMetrics:    req.DropMetrics,
	}

	if err := model.CreateMonitorRule(rule); err != nil {
		query.API500Response(c, err)
		return
	}
	query.SuccessResponse(c, query.OK, rule)
}

// @Summary      Update monitor rule
// @Description  Modify an existing monitor rule
// @Tags         monitor
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Monitor Rule ID"
// @Param        rule     body      monitor.MonitorRuleReq  true  "Monitor Rule configuration"
// @Success      200      {object}  query.Response
// @Failure      400      {object}  query.Response
// @Failure      500      {object}  query.Response
// @Router       /ph/v1/monitors/{id} [post]
func (h *MonitorHandlerV1) updateMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		query.API400Response(c, err)
		return
	}

	var req monitor.MonitorRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		query.API400Response(c, err)
		return
	}

	if req.MetricPath == "" {
		req.MetricPath = "/metrics"
	}

	rule := &model.MonitorRule{
		Name:           req.Name,
		Type:           req.Type,
		Namespace:      req.Namespace,
		SelectorString: req.SelectorString,
		PortName:       req.PortName,
		MetricPath:     req.MetricPath,
		LabelsString:   req.LabelsString,
		DropMetrics:    req.DropMetrics,
	}

	if err := model.UpdateMonitorRule(uint(id), rule); err != nil {
		query.API500Response(c, err)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}

// @Summary      Delete monitor rule
// @Description  Remove a monitor rule by ID
// @Tags         monitor
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Monitor Rule ID"
// @Success      200      {object}  query.Response
// @Failure      400      {object}  query.Response
// @Failure      500      {object}  query.Response
// @Router       /ph/v1/monitors/{id} [delete]
func (h *MonitorHandlerV1) deleteMonitor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		query.API400Response(c, err)
		return
	}

	if err := model.DeleteMonitorRule(uint(id)); err != nil {
		query.API500Response(c, err)
		return
	}
	query.SuccessResponse(c, query.OK, nil)
}
