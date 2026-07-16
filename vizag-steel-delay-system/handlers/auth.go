package handlers

import (
	"net/http"

	"vizag-steel-delay-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}

func ShowLogin(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		c.Redirect(http.StatusFound, "/entry")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func Login(c *gin.Context) {
	empNo := c.PostForm("emp_no")
	password := c.PostForm("password")

	var user models.User
	if err := db.Where("emp_no = ? AND active = ?", empNo, true).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid credentials or inactive account"})
		return
	}

	if user.Password != password {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid credentials"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.EmpNo)
	session.Set("role", user.Role)
	session.Set("emp_name", user.EmpName)
	session.Save()

	c.Header("HX-Redirect", "/entry")
	c.Redirect(http.StatusFound, "/entry")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userRole := session.Get("role")
		if userRole == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
