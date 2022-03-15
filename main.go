package main

import (
	"github.com/rawansuww/clinic-booking/controllers"
	"github.com/rawansuww/clinic-booking/middleware"
	"github.com/rawansuww/clinic-booking/models"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	models.ConnectDatabase()

	// Routes
	public := r.Group("/public")
	registered := r.Group("/")
	registered.Use(middleware.Authz()) //JWT bearer token

	// all endpoints + basic CRUD (Create=general registration)
	//doctors
	registered.GET("/doctors", controllers.FindDoctors)
	registered.GET("/doctors/:id", controllers.FindDoctor)
	registered.GET("/doctors/:id/schedule", controllers.FindDoctorSchedule)
	registered.POST("/doctors/:id/slots", controllers.FindDoctorAvailability) //takes input date
	registered.PATCH("/doctors/:id", controllers.UpdateDoctor)
	registered.DELETE("/doctors/:id", controllers.DeleteDoctor)
	registered.POST("/doctors/most/count", controllers.FindDoctorsMost)
	registered.POST("/doctors/most/hours", controllers.FindDoctorsLongest)
	registered.GET("/doctors/slots/all", controllers.FindDoctorAvailAll)

	//patients
	registered.GET("/patients", controllers.FindPatients)
	registered.GET("/patients/:id", controllers.FindPatient)
	registered.GET("/patients/:id/history", controllers.FindPatientHistory)
	registered.PATCH("/patients/:id", controllers.UpdatePatient)
	registered.DELETE("/patients/:id", controllers.DeletePatient)

	//appointments
	registered.POST("/doctors/:id/schedule", controllers.CreateAppointment)
	registered.GET("/appointments/:id", controllers.FindAppointment)
	registered.DELETE("/appointments/:id", controllers.DeleteAppointment)

	//public
	public.POST("/login", controllers.Login)
	public.POST("/signup", controllers.Signup)

	// Run the server
	r.Run()
}
