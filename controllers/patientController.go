package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rawansuww/clinic-booking/models"
)

type CreatePatientInput struct {
	Name     string               `json:"name" binding:"required"`
	Email    string               `json:"email" binding:"required"`
	Password string               `json:"password" binding:"required"`
	History  []models.Appointment `json:"history"`
}

type UpdatePatientInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DoctorToPatientHist struct {
	Appointment models.Appointment
	Doctor      DoctorToPatient
}

type DoctorToPatient struct {
	Name  string
	Email string
}

// GET /patients
// Find all patients
func FindPatients(c *gin.Context) {
	role := c.GetString("role")
	if role == "admin" {
		var patients []models.Patient
		models.DB.Find(&patients)

		c.JSON(http.StatusOK, gin.H{"patients": patients})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to be a clinic admin to see the list of patients!"})
	}
}

// GET /patients/:id
// Find a patient
func FindPatient(c *gin.Context) {
	// Get model if exist
	var doc models.Patient
	role := c.GetString("role")
	if role == "admin" {
		if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"patient": doc})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to be a clinic admin to access patient information!"})
	}
}

//GET Patient History
func FindPatientHistory(c *gin.Context) {
	var patient models.Patient
	var pat models.Patient
	var doc models.Doctor
	var history []DoctorToPatientHist

	email := c.GetString("email")
	role := c.GetString("role")

	if role == "patient" {
		models.DB.Where("email=?", email).First(&pat)
		if fmt.Sprint(pat.ID) != c.Param("id") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot view another patient's history! Login to see your own history!"})
			return
		}
	} else {
		if err := models.DB.Where("id = ?", c.Param("id")).First(&patient).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
			return
		}

		models.DB.Preload("History").Find(&patient)

		for i := 0; i < len(patient.History); i++ {
			take := patient.History[i].DID

			models.DB.Raw("SELECT * FROM doctors WHERE id=?", take).First(&doc)
			doctor := DoctorToPatient{doc.Name, doc.Email}
			data := DoctorToPatientHist{patient.History[i], doctor}
			if patient.History[i].StartTime.Before(time.Now()) { //making sure history is iN PAST
				history = append(history, data)

			}

		}
		c.JSON(http.StatusOK, gin.H{"history": history})
	}
}

// PATCH /patients/:id
// Update a patient
func UpdatePatient(c *gin.Context) {
	var doc models.Patient
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Validate input
	var input UpdatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&doc).Updates(input)

	c.JSON(http.StatusOK, gin.H{"data": doc})
}

// DELETE /patients/:id
// Delete a patient
func DeletePatient(c *gin.Context) {

	role := c.GetString("role")

	if role == "admin" {
		var patient models.Patient
		if err := models.DB.Where("id = ?", c.Param("id")).First(&patient).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
			return
		}

		models.DB.Delete(&patient)

		c.JSON(http.StatusOK, gin.H{"deleted": true})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have to be a clinic admin to delete a patient record!"})
	}
}
