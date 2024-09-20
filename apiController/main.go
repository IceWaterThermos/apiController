package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"sync"
)

var (
	dynamicRoutes = make(map[string]interface{})
	routesMutex   = sync.RWMutex{} // 동시성을 처리하기 위한 뮤텍스
)

func main() {
	router := gin.Default()

	// 정적 파일 제공
	router.Static("/static", "./static")

	// 홈 페이지
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", nil)
	})

	// 회원가입 페이지
	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	// 로그인 페이지
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// 로그인 성공 후 메뉴 페이지
	router.GET("/menu", func(c *gin.Context) {
		c.HTML(http.StatusOK, "menu.html", nil)
	})

	// createapi 페이지
	router.GET("/createapi", func(c *gin.Context) {
		c.HTML(http.StatusOK, "createapi.html", nil)
	})

	// apicontroller 페이지
	router.GET("/apicontroller", func(c *gin.Context) {
		c.HTML(http.StatusOK, "apicontroller.html", nil)
	})

	// 로그인 처리
	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		if username == "admin" && password == "password" {
			c.Redirect(http.StatusFound, "/menu")
		} else {
			c.String(http.StatusUnauthorized, "Invalid credentials")
		}
	})

	// 동적으로 핸들러 생성
	router.POST("/createapi", func(c *gin.Context) {
		var requestData struct {
			Url  string                 `json:"url"`
			Json map[string]interface{} `json:"json"`
		}

		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		// 동적으로 라우트를 추가하기 위해 뮤텍스 사용
		routesMutex.Lock()
		dynamicRoutes[requestData.Url] = requestData.Json
		routesMutex.Unlock()

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// API Controller 프록시 처리
	router.Any("/proxy", func(c *gin.Context) {
		targetUrl := c.Query("url")

		routesMutex.RLock()
		data, exists := dynamicRoutes[targetUrl]
		routesMutex.RUnlock()

		if exists {
			// 동적 API 처리
			switch c.Request.Method {
			case http.MethodGet:
				c.JSON(http.StatusOK, data)
			case http.MethodPost, http.MethodPut:
				var newData map[string]interface{}
				if err := c.BindJSON(&newData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}
				// JSON 업데이트
				routesMutex.Lock()
				dynamicRoutes[targetUrl] = newData
				routesMutex.Unlock()
				c.JSON(http.StatusOK, gin.H{"message": "API updated", "data": newData})
			case http.MethodDelete:
				// API 삭제
				routesMutex.Lock()
				delete(dynamicRoutes, targetUrl)
				routesMutex.Unlock()
				c.JSON(http.StatusOK, gin.H{"message": "API deleted"})
			default:
				c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			}
			return
		}

		// 동적 API가 없을 경우 외부 API로 프록시 요청
		method := c.Request.Method
		req, err := http.NewRequest(method, targetUrl, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		req.Header = c.Request.Header
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Request failed"})
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	})

	// HTML 템플릿 로드
	router.LoadHTMLGlob("templates/*")

	router.Run(":8080")
}

// GET 요청 처리 함수
func handleGetRequest(c *gin.Context, url string) {
	c.JSON(http.StatusOK, gin.H{
		"message": "GET request successful!",
		"url":     url,
	})
}

// POST 요청 처리 함수
func handlePostRequest(c *gin.Context, url string) {
	var jsonData map[string]interface{}
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "POST request successful!",
		"url":     url,
		"data":    jsonData,
	})
}

// PUT 요청 처리 함수
func handlePutRequest(c *gin.Context, url string) {
	var jsonData map[string]interface{}
	if err := c.BindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "PUT request successful!",
		"url":     url,
		"data":    jsonData,
	})
}

// DELETE 요청 처리 함수
func handleDeleteRequest(c *gin.Context, url string) {
	c.JSON(http.StatusOK, gin.H{
		"message": "DELETE request successful!",
		"url":     url,
	})
}
