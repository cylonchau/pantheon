package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	v1Proxy "github.com/cylonchau/pantheon/pkg/server/v1/proxy"
	v1Selector "github.com/cylonchau/pantheon/pkg/server/v1/selector"
	v1Target "github.com/cylonchau/pantheon/pkg/server/v1/target"
	v2Target "github.com/cylonchau/pantheon/pkg/server/v2/target"
)

func RegisteredRouter(e *gin.Engine) {
	phAPIGroup := e.Group("/ph")
	phv1Group := phAPIGroup.Group("/v1")
	phv2Group := phAPIGroup.Group("/v2")

	targetHanderV1 := &v1Target.TargetHanderV1{}
	targetHanderV1.RegisterTargetAPI(phv1Group)

	selectorHanderV1 := &v1Selector.SelectorHanderV1{}
	selectorHanderV1.RegisterSelectorAPI(phv1Group)

	proxyHanderV1 := &v1Proxy.ProxyHanderV1{}
	proxyHanderV1.RegisterProxyAPI(phv1Group)

	targetHanderV2 := &v2Target.TargetHanderV2{}
	targetHanderV2.RegisterTargetAPI(phv2Group)

	e.Handle("GET", "/doc/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/doc/doc.json")))
	e.GET("/doc", func(c *gin.Context) {
		c.Redirect(302, "/doc/index.html")
	})

}
