package main

import (
	"github.com/rahmanfadhil/gin-bookstore/controllers"
	"github.com/rahmanfadhil/gin-bookstore/middleware"
	"github.com/rahmanfadhil/gin-bookstore/models"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Connect to database
	models.ConnectDatabase()
	//REMEMBER TO GROUP THE API ROUTES TO PUBLIC AND PRIVATE!!!!!!!!!
	// Routes

	public := r.Group("/public")
	registered := r.Group("/")
	registered.Use(middleware.Authz())

	registered.GET("/doctors", controllers.FindDoctors)
	registered.GET("/doctors/:id", controllers.FindDoctor)
	registered.GET("/doctors/:id/schedule", controllers.FindDoctorSchedule)
	registered.GET("/doctors/:id/slots", controllers.FindDoctorAvailability)
	registered.POST("/doctors", controllers.CreateDoctor)
	registered.PATCH("/doctors/:id", controllers.UpdateDoctor)
	registered.DELETE("/doctors/:id", controllers.DeleteDoctor)
	registered.POST("/doctors/most/count", controllers.FindDoctorsMost) //didnt adjust other func
	registered.POST("/doctors/most/hours", controllers.FindDoctorsLongest)
	registered.GET("/doctors/slots/all", controllers.FindDoctorAvailAll)

	//for patients
	registered.GET("/patients", controllers.FindPatients)
	registered.GET("/patients/:id", controllers.FindPatient)
	registered.GET("/patients/:id/history", controllers.FindPatientHistory)
	//	r.POST("/patients", controllers.CreatePatient)
	registered.PATCH("/patients/:id", controllers.UpdatePatient)
	registered.DELETE("/patients/:id", controllers.DeletePatient)

	//for appointments
	registered.POST("/doctors/:id/schedule", controllers.CreateAppointment)
	registered.GET("/appointments/:id", controllers.FindAppointment)
	registered.DELETE("/appointments/:id", controllers.DeleteAppointment)

	//group the api endpoints into public and private!  using TOKE
	public.POST("/login", controllers.Login)
	public.POST("/signup", controllers.Signup)

	// Run the server
	r.Run()
}
