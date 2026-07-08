package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/cylonchau/pantheon/pkg/api/query"
	"github.com/cylonchau/pantheon/pkg/config"
)

type ProxyHanderV1 struct{}

func (p *ProxyHanderV1) RegisterProxyAPI(g *gin.RouterGroup) {
	proxyGroup := g.Group("/proxy")
	proxyGroup.GET("", p.proxy)
}

// proxy godoc
// @Summary Reverse Proxy
// @Description This endpoint proxies requests to another server.
// @Tags Proxy
// @Accept json
// @Produce json
// @Param schema query string true "Protocol (http/https)"
// @Param host query string true "Host to proxy to"
// @Param port query string true "Port to proxy to"
// @Param base query string false "Basic auth credentials"
// @Param bearer query string false "Bearer token for authentication"
// @Param param1 query string false "Additional parameter 1"
// @Param param2 query string false "Additional parameter 2"
// @securityDefinitions.apikey BearerAuth
// @Success 200 {object} interface{}
// @Router /ph/v1/proxy [get]
func (p *ProxyHanderV1) proxy(c *gin.Context) {

	start := time.Now()
	duration := time.Since(start)

	// 1. 获取参数和参数校验
	schema := c.Query("schema")
	host := c.Query("host")
	port := c.Query("port")
	base := c.Query("base")
	bearer := c.Query("bearer")
	path := c.Query("path")
	dropPattern := c.Query("drop")

	// 验证必需的参数
	if schema == "" || host == "" || port == "" {
		query.API400Response(c, fmt.Errorf("schema, host, port are required"))
		return
	}

	// 验证 schema
	if schema != "http" && schema != "https" {
		query.API400Response(c, fmt.Errorf("invalid schema; only 'http' and 'https' are allowed"))
		return
	}

	// 验证 host
	if host == "" {
		query.API400Response(c, fmt.Errorf("invalid host"))
		return
	}
	// 验证 port
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		query.API400Response(c, fmt.Errorf("invalid port; must be a number between 1 and 65535"))
		return
	}

	// 构建目标 URL，省略默认端口
	target := fmt.Sprintf("%s://%s", schema, host)

	if (schema == "http" && portNum != 80) || (schema == "https" && portNum != 443) {
		target = fmt.Sprintf("%s:%s", target, port)
	}

	if path != "" {
		decodedPath, _ := url.QueryUnescape(path)
		if decodedPath[0] != '/' {
			decodedPath = "/" + decodedPath
		}
		target = fmt.Sprintf("%s%s", target, decodedPath)
	}

	// 解析目标 URL
	targetURL, err := url.Parse(target)
	if err != nil {
		query.API400Response(c, fmt.Errorf("invalid target URL"))
		return
	}
	fmt.Println(targetURL)
	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	if dropPattern != "" {
		proxy.ModifyResponse = func(resp *http.Response) error {
			if resp.StatusCode != http.StatusOK {
				return nil
			}

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()

			filteredBytes := filterMetrics(bodyBytes, dropPattern)

			resp.Body = ioutil.NopCloser(bytes.NewReader(filteredBytes))
			resp.ContentLength = int64(len(filteredBytes))
			resp.Header.Set("Content-Length", strconv.Itoa(len(filteredBytes)))
			return nil
		}
	}
	// 设置超时

	proxy.Transport = http.Client{
		Timeout: time.Duration(config.CONFIG.ProxyTimeout) * time.Second, // 设置更长的超时时间
	}.Transport
	// 设置认证头和其他参数
	proxy.Director = func(req *http.Request) {
		req.URL = targetURL

		// 添加认证头
		if base != "" {
			req.Header.Set("Authorization", "Basic "+base)
		} else if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}

		// 添加 User-Agent
		if userAgent := c.Request.Header.Get("User-Agent"); userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}

		// 添加其他参数，排除已知参数
		query := c.Request.URL.Query()
		for key := range query {
			if key != "schema" && key != "host" && key != "port" && key != "base" && key != "bearer" && key != "path" {
				for _, value := range query[key] {
					req.URL.Query().Add(key, value)
				}
			}
		}
		// 将参数更新到请求 URL
		req.URL.RawQuery = req.URL.Query().Encode()
	}
	klog.V(4).Infof("Proxying request to: %s", targetURL.String())
	proxy.ServeHTTP(c.Writer, c.Request)
	if c.Writer.Status() != http.StatusOK {
		klog.Errorf("Failed to proxy request, status code: %d", c.Writer.Status())
	}
	// 记录日志，格式化为 Nginx 访问日志格式
	klog.V(4).Infof("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %s",
		c.ClientIP(),
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		c.Request.Method,
		c.Request.RequestURI,
		c.Request.Proto,
		c.Writer.Status(),
		c.Writer.Size(),
		c.Request.Referer(),
		c.Request.UserAgent(),
		duration,
	)
}

func filterMetrics(rawMetrics []byte, dropPattern string) []byte {
	if dropPattern == "" {
		return rawMetrics
	}
	re, err := regexp.Compile(dropPattern)
	if err != nil {
		klog.Errorf("Invalid drop regex pattern: %s, error: %v", dropPattern, err)
		return rawMetrics
	}

	scanner := bufio.NewScanner(bytes.NewReader(rawMetrics))
	var buf bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		// 1. Check if comment/metadata line
		if strings.HasPrefix(line, "#") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# HELP") || strings.HasPrefix(trimmed, "# TYPE") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 3 {
					metricName := parts[2]
					if re.MatchString(metricName) {
						continue // drop metadata for this metric
					}
				}
			}
			buf.WriteString(line + "\n")
			continue
		}

		// 2. Extract metric name
		idx := strings.IndexAny(line, "{ \t")
		var metricName string
		if idx == -1 {
			metricName = line
		} else {
			metricName = line[:idx]
		}

		// 3. Drop if it matches regex
		if re.MatchString(metricName) {
			continue
		}
		buf.WriteString(line + "\n")
	}

	return buf.Bytes()
}
