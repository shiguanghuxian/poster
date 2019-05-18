package transport

import (
	"fmt"
	"net/http"
	"time"

	gin "github.com/gin-gonic/gin"
	"github.com/shiguanghuxian/poster/program/config"
	"github.com/shiguanghuxian/poster/program/service"
)

// HTTPTransport 提供http服务生成海报
type HTTPTransport struct {
	cfg *config.HTTPConfig
}

// NewHTTPTransport 创建http接口对象
func NewHTTPTransport(cfg *config.HTTPConfig) *HTTPTransport {
	if cfg.Port < 0 {
		cfg.Port = 80
	}
	return &HTTPTransport{
		cfg: cfg,
	}
}

// ListenAndServe 监听启动服务
func (s *HTTPTransport) ListenAndServe(debug bool) (err error) {
	// 当前运行模式
	if debug == true {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// gin默认引擎
	router := gin.Default()
	// 跨域问题
	router.Use(s.middleware())
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(s.cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.WriteTimeout) * time.Second,
	}
	// 生成海报api
	router.POST("/create", s.createPoster)

	// 启动监听
	err = server.ListenAndServe()
	return
}

func (s *HTTPTransport) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

// 生成一个海报
func (s *HTTPTransport) createPoster(c *gin.Context) {
	// 捕获异常
	req := new(service.PosterParam)
	err := c.Bind(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 生成图片
	srv, err := service.NewService(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	img, err := srv.DrawPoster()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Header("Content-Type", "image/jpeg")
	c.Writer.Write(img)
	// // 返回图片，json格式
	// c.JSON(http.StatusOK, gin.H{
	// 	"image": img,
	// })
}
