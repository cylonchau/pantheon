package app

import (
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/cylonchau/pantheon/pkg/config"
	"github.com/cylonchau/pantheon/pkg/server/router"
)

var http *gin.Engine
var stopCh = make(chan struct{})

func init() {
	gin.DefaultWriter = ioutil.Discard
	gin.DisableConsoleColor()
}

func NewHTTPSever() (err error) {
	http = gin.New()
	router.RegisteredRouter(http)
	klog.V(0).Infof("Listening and serving HTTP on %s:%s", config.CONFIG.Address, config.CONFIG.Port)

	if err = http.Run(fmt.Sprintf("%s:%s", config.CONFIG.Address, config.CONFIG.Port)); err != nil {
		return err
	}
	<-stopCh
	return
}
