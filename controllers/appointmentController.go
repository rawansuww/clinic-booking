package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rawansuww/clinic-booking/models"
)

type BookSlot struct {
	PID       string        `json:"pID"` //you will get the pID from JWT TOKEN later
	DID       uint          `json:"dID"` //you will get the dID from URL param
	Booked    bool          `json:"booked"`
	StartTime time.Time     `json:"startTime" binding:"required"`
	EndTime   time.Time     `json:"endTime" binding:"required"`
	Duration  time.Duration `json:"duration"`
}

func CreateAppointment(c *gin.Context) {
	var input BookSlot
	var doc models.Doctor
	var app models.Appointment
	var count int64
	var totTime float32
	var overlap bool
	var patient models.Patient

	doctorID, er := strconv.Atoi(c.Param("id"))
	doc.ID = uint(doctorID)

	email := c.GetString("email")

	if err := models.DB.Where("email = ?", email).First(&patient).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to be logged in as a patient book an appointment!"})
		return
	}

	if er != nil {
		c.JSON(http.StatusBadRequest, nil)
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//--------------------------------- INPUT VALIDATION BLOCK
	attempt := input.StartTime
	attempt2 := input.EndTime
	now := time.Now()

	if attempt.Before(now) { //1. check if appointment not in past
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot book an appointment in the past!"})
		return
	}

	//2. check min and max duration of an appointment
	if input.EndTime.Sub(input.StartTime).Minutes() < 15 || input.EndTime.Sub(input.StartTime).Hours() > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment must be between 15 minutes and 2 hours!"})
		return
	}

	//3. to check for 12 patients [for a given date for a given doctor]

	t := attempt.Format("2006-01-02")

	models.DB.Raw("SELECT COUNT(*) FROM appointments WHERE DATE(start_time) = ? AND d_id=?", string(t), doc.ID).Scan(&count)

	if count >= 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The doctor already has 12 patients for this day! Please choose another day."})
		return
	}

	//3. to check for 8 hours total [for a given date for a given doctor]
	models.DB.Raw("SELECT SUM((TIMEDIFF(end_time, start_time)/10000)) FROM appointments WHERE DATE(start_time) = ? AND d_id=?", string(t), doc.ID).Scan(&totTime)

	if totTime >= 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This doctor already has 8 HOURS for this day! Please choose another day."})
		return
	}

	models.DB.Preload("Schedule").Find(&app)

	//5. CHECK that given booking is after 9 AM and before 5 PM of that date

	year1, month1, day1 := attempt.Date()

	day9AM := time.Date(year1, month1, day1, 9, 0, 0, 0, time.UTC)
	day5PM := time.Date(year1, month1, day1, 17, 0, 0, 0, time.UTC)

	if attempt.Before(day9AM) || attempt2.After(day5PM) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctors only work from 9 AM to 5 PM!"})
		return
	}

	//4.to check for duplicate times AND time overlaps!!!-----
	dup1 := attempt.Format("2006-01-02T15:04:05Z07:00")
	dup2 := attempt2.Format("2006-01-02T15:04:05Z07:00")
	models.DB.Raw("SELECT EXISTS (SELECT * FROM appointments WHERE d_id=? AND start_time BETWEEN ? AND ? OR end_time BETWEEN ? and ?)", doc.ID, string(dup1), string(dup2), string(dup1), string(dup2)).Scan(&overlap)
	if overlap {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your appointment time overlaps with another appointment!"})
		return
	}

	//--------------------------- end of validation BLOCK

	//finally, check if doc exist or if error in JSON sent
	if err := models.DB.First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found!"})
		return
	}

	// If all good, Create appointment
	apptmt := models.Appointment{PID: patient.ID, DID: doc.ID, StartTime: input.StartTime, EndTime: input.EndTime, Duration: time.Duration(input.EndTime.Sub(input.StartTime).Minutes())}
	models.DB.Create(&apptmt)

	c.JSON(http.StatusOK, gin.H{"appointment booked": apptmt})

	//models.DB.Model(&doc).Association("Schedule").Append(apptmt) //!!!! THIS IS CAUSING AN ERROR.... BIG ERROR
	models.DB.Model(&doc).Select("Schedule").Updates(apptmt)

}

// DELETE an appointment
func DeleteAppointment(c *gin.Context) {
	role := c.GetString("role")

	if role != "admin" && role != "doctor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You as a patient cannot cancel appointments! Sad!"})
		return
	}

	var doctor models.Doctor

	aID, er := strconv.Atoi(c.Param("id"))
	if er != nil {
		c.JSON(http.StatusBadRequest, nil)
	}

	var apptmt models.Appointment
	if err := models.DB.Where("id = ?", aID).First(&apptmt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if role == "doctor" {
		email := c.GetString("email")
		models.DB.Where("email = ?", email).First(&doctor)
		if doctor.ID != apptmt.DID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are not the doctor booked, so you cannot cancel this appointment!"})
			return
		}

	}

	models.DB.Delete(&apptmt)
	models.DB.Model(&doctor).Select("Schedule").Updates(apptmt)

	c.JSON(http.StatusOK, gin.H{"cancelled": true})
}

func FindAppointment(c *gin.Context) {
	role := c.GetString("role")

	var app models.Appointment
	var doctor models.Doctor
	var patient models.Patient

	if err := models.DB.Where("id = ?", c.Param("id")).First(&app).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if role == "patient" {
		models.DB.Where("email = ?", c.GetString("email")).First(&patient)
		if patient.ID == app.PID {
			models.DB.Raw("SELECT name, email FROM patients WHERE id=?", app.PID).Find(&patient)
			models.DB.Raw("SELECT name, email FROM doctors WHERE id=?", app.DID).Find(&doctor)
			c.JSON(http.StatusOK, gin.H{"appointment": app, "patient": patient, "doctor": doctor})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are not the patient who booked the appointment, not authorized to see appointment details!"})
		}
	} else if role == "doctor" {
		models.DB.Where("email = ?", c.GetString("email")).First(&doctor)
		if doctor.ID == app.DID {
			models.DB.Raw("SELECT name, email FROM patients WHERE id=?", app.PID).Find(&patient)
			models.DB.Raw("SELECT name, email FROM doctors WHERE id=?", app.DID).Find(&doctor)
			c.JSON(http.StatusOK, gin.H{"appointment": app, "patient": patient, "doctor": doctor})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are not the doctor who is booked for the appointment, not authorized to see appointment details!"})
		}
	} else { //admins can see any appointment ever!
		models.DB.Raw("SELECT name, email FROM patients WHERE id=?", app.PID).Find(&patient)
		models.DB.Raw("SELECT name, email FROM doctors WHERE id=?", app.DID).Find(&doctor)
		c.JSON(http.StatusOK, gin.H{"appointment": app, "patient": patient, "doctor": doctor})
	}

	models.DB.Preload("Schedule").Find(&app)

}
