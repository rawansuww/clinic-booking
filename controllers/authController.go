package controllers

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	_ "gorm.io/gorm"

	//"github.com/rahmanfadhil/gin-bookstore/auth"
	"github.com/rahmanfadhil/gin-bookstore/middleware"
	"github.com/rahmanfadhil/gin-bookstore/models"
)

// Signup creates a user in db
func Signup(c *gin.Context) {
	var input SignUp

	err := c.ShouldBindJSON(&input)

	fmt.Println(string(input.Role))

	if err != nil {
		fmt.Println(err)

		c.JSON(400, gin.H{
			"msg": "invalid json",
		})
		c.Abort()

		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	if input.Role == "doctor" {
		doc := models.Doctor{Name: input.Name, Email: input.Email, Password: string(hashed), Role: input.Role}
		models.DB.Create(&doc)

		c.JSON(200, doc)
	}
	if input.Role == "patient" {
		patient := models.Patient{Name: input.Name, Email: input.Email, Password: string(hashed), Role: input.Role}
		models.DB.Create(&patient)

		c.JSON(200, patient)
	}
	if input.Role == "admin" {
		admin := models.Admin{Name: input.Name, Email: input.Email, Password: string(hashed), Role: input.Role}
		models.DB.Create(&admin)

		c.JSON(200, admin)
	}

	if err != nil {
		log.Println(err.Error())

		c.JSON(500, gin.H{
			"msg": "error hashing password",
		})
		c.Abort()

		return
	}

}

// controllers/public.go

// LoginPayload login body
type LoginPayload struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// LoginResponse token response
type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type SignUp struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

// Login logs users in
//THE LOGIN I NEED TO DECIDE HOW TO LOGIN AS DIFFERENT TYPES OF USERS....
func Login(c *gin.Context) {
	var payload LoginPayload
	//var user models.Doctor

	var user models.Doctor
	var doc models.Doctor
	var patient models.Patient
	var admin models.Admin

	//var user LoginPayload

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "One of the fields is missing. Invalid JSON",
		})
		c.Abort()
		return
	}
	if payload.Role == "doctor" {
		fmt.Println("enter")

		result := models.DB.Where("email = ?", payload.Email).Find(&(doc))
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(401, gin.H{
				"msg": "invalid user credentials",
			})
			c.Abort()
			return
		}
		copier.CopyWithOption(&user, &doc, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	}

	if payload.Role == "patient" {
		result := models.DB.Where("email = ?", payload.Email).Find(&patient)
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(401, gin.H{
				"msg": "invalid user credentials",
			})
			c.Abort()
			return
		}
		copier.CopyWithOption(&user, &patient, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	}
	if payload.Role == "admin" {
		result := models.DB.Where("email = ?", payload.Email).Find(&admin)
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(401, gin.H{
				"msg": "invalid user credentials",
			})
			c.Abort()
			return
		}
		copier.CopyWithOption(&user, &admin, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	}

	fmt.Println(user)
	//	result := models.DB.Where("email = ?", payload.Email).Find(&user)

	e := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if e != nil {

		c.JSON(401, gin.H{
			"msg": "Wrong password",
		})
		c.Abort()
		return
	}

	jwtWrapper := middleware.JwtWrapper{
		SecretKey:       "verysecretkey",
		Issuer:          "AuthService",
		ExpirationHours: 24,
	}

	signedToken, err := jwtWrapper.GenerateToken(user.Email, user.Role)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"msg": "error signing token",
		})
		c.Abort()
		return
	}

	tokenResponse := LoginResponse{
		Token:   signedToken,
		Message: "Login Successful",
	}

	c.JSON(200, tokenResponse)

}
