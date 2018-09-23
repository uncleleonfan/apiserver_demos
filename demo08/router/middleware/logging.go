package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"time"

	"apiserver_demos/demo08/handler"
	"apiserver_demos/demo08/pkg/errno"

	"github.com/gin-gonic/gin"
	"github.com/lexkong/log"
	"github.com/willf/pad"
)

// bodyLogWriter实现了http.ResponseWriter的Write接口
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// 将网络网络响应结果写入body
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

//是的，这里还是很吃性能的，可以在debug时用，线上环境就不要用了
// Logging is a middleware function that logs the each request.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path

		//该中间件只记录业务请求，比如 /v1/user 和 /login 路径
		reg := regexp.MustCompile("(/v1/user|/login)")
		if !reg.MatchString(path) {
			return
		}

		// Skip for the health check requests.
		if path == "/sd/health" || path == "/sd/ram" || path == "/sd/cpu" || path == "/sd/disk" {
			return
		}

		// Read the Body content
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		}

		// Restore the io.ReadCloser to its original state
		//该中间件需要截获 HTTP 的请求信息，然后打印请求信息，因为 HTTP 的请求 Body，在读取过后会被置空，所以这里读取完后会重新赋值：
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// The basic informations.
		method := c.Request.Method
		ip := c.ClientIP()

		//这里可以打印请求相关到信息
		//log.Debugf("New request come in, path: %s, Method: %s, body `%s`", path, method, string(bodyBytes))
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		// Continue. 执行下一个中间件
		c.Next()

		// Calculates the latency.
		end := time.Now().UTC()
		latency := end.Sub(start)

		code, message := -1, ""

		// get code and message
		var response handler.Response
		// 反序列化
		if err := json.Unmarshal(blw.body.Bytes(), &response); err != nil {
			log.Errorf(err, "response body can not unmarshal to model.Response struct, body: `%s`", blw.body.Bytes())
			code = errno.InternalServerError.Code
			message = err.Error()
		} else {
			code = response.Code
			message = response.Message
		}
		//打印网络响应结果
		log.Infof("%-13s | %-12s | %s %s | {code: %d, message: %s}", latency, ip, pad.Right(method, 5, ""), path, code, message)
	}
}
