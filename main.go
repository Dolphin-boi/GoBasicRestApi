package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserName string `gorm: "primarykey"`
	Email    string `gorm: "unique"`
	Fname    string
	Lname    string
}

type Data struct {
	Username string `json: "username"`
	Email    string `json: "email" binding require`
	Fname    string `json: "fname" binding require`
	Lname    string `json: "lname" binding require`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file")
	}

	dbUserName := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := dbUserName + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&User{})

	app := gin.Default()

	app.GET("/users", func(c *gin.Context) {
		var user User
		result := db.First(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"List": user})
	})

	app.GET("/users/:username", func(c *gin.Context) {
		username := c.Param("username")
		var user User
		result := db.Where("user_name = ?", username).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "User not exist"})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"username":   user.UserName,
			"email":      user.Email,
			"first name": user.Fname,
			"last name":  user.Lname})
	})

	app.POST("/users", func(c *gin.Context) {

		var data Data
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		result := db.Where("user_name = ?", data.Username).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				user = User{UserName: data.Username, Email: data.Email, Fname: data.Fname, Lname: data.Lname}
				result = db.Create(&user)
				if result.Error != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"New user": user.Email})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}
		c.JSON(http.StatusConflict, gin.H{"error": "User already exist"})
	})

	app.PUT("/users/:username", func(c *gin.Context) {

		username := c.Param("username")
		var data Data
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		result := db.Where("user_name = ?", username).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "User not exist"})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}
		user = User{UserName: username, Email: data.Email, Fname: data.Fname, Lname: data.Lname}
		result = db.Where("user_name = ?", user.UserName).Updates(User(user))
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"username":   user.UserName,
			"email":      user.Email,
			"first name": user.Fname,
			"last name":  user.Lname})
	})

	app.DELETE("/users/:username", func(c *gin.Context) {

		username := c.Param("username")

		var user User
		result := db.Where("user_name = ?", username).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "User not exist"})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			}
		}

		result = db.Where("user_name = ?", username).Delete(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "User is deleted"})
	})

	app.Run(":8000")
}
