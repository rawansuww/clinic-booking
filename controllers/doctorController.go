package controllers

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rawansuww/clinic-booking/models"
)

type CreateDoctorInput struct {
	Name     string             `json:"name" binding:"required"`
	Schedule models.Appointment `json:"schedule"`
	Email    string             `json:"email" binding:"required"`
	Password string             `json:"password" binding:"required"`
}

type UpdateDoctorInput struct {
	Name     string             `json:"name"`
	Schedule models.Appointment `json:"schedule"`
	Email    string             `json:"email"`
	Password string             `json:"password"`
}

type Given struct {
	GivenDay time.Time `json:"day" binding:"required"`
}

type AutoGenerated struct {
	DoctorID     int           `json:"doctorID"`
	Doctor       models.Doctor `json:"doctor"`
	Appointments int           `json:"totalHours"`
}

type DoctorSchedule struct {
	Appointment models.Appointment
	Patient     models.Patient
}

type D struct {
	Doctor       int      `json:"doctorID"`
	Availability []string `gorm:"type:text" json:"availability"`
}

// GET /doctors
// Find all doctors
func FindDoctors(c *gin.Context) {
	var docs []models.Doctor
	var patient models.Patient

	email := c.GetString("email")

	if err := models.DB.Where("email = ?", email).First(&patient).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to be either a doctor or clinic admin to access this!"})
		return
	}

	models.DB.Model(&models.Doctor{}).Preload("Schedule").Find(&docs)

	c.JSON(http.StatusOK, gin.H{"doctors": docs})
}

// most appointments for a given day.. only accessible to admins
func FindDoctorsMost(c *gin.Context) {
	var input Given
	var doc models.Doctor
	var dID int
	var freq int
	var result []AutoGenerated

	role := c.GetString("role")

	if role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only clinic admins can see doctors with most appointments!"})
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dup1 := input.GivenDay.Format("2006-01-02") //getting zero error.. fix input later!
	fmt.Println(dup1)
	rows, err := models.DB.Raw("Select d_id, count(*) From appointments WHERE DATE(start_time) = ? Group By d_id order by count(*) desc", string(dup1)).Rows()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time to cry"})
		return
	}

	for rows.Next() {
		err := rows.Scan(&dID, &freq)
		if err != nil {
			log.Fatal(err)
		}
		if err := models.DB.Where("id = ?", dID).Find(&doc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found!"})
			return
		}

		models.DB.Preload("Schedule").Find(&doc)
		result = append(result, AutoGenerated{dID, doc, freq})
	}

	c.JSON(http.StatusOK, gin.H{"doctors with most appointments on this day": result})
}

//doctors with 6+ hrs on a given day.... to admins
func FindDoctorsLongest(c *gin.Context) {
	var input Given
	var doc models.Doctor
	var dID int
	var freq float32
	var result []AutoGenerated

	role := c.GetString("role")

	if role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only clinic admins can see doctors with most appointments!"})
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dup1 := input.GivenDay.Format("2006-01-02") //getting zero error.. fix input later!
	rows, err := models.DB.Raw("Select d_id, SUM((TIMEDIFF(end_time, start_time)/10000)) From appointments WHERE DATE(start_time) = ? Group By d_id HAVING SUM((TIMEDIFF(end_time, start_time)/10000))>6 order by SUM((TIMEDIFF(end_time, start_time)/10000)) desc", string(dup1)).Rows()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "time to cry"})
		return
	}

	for rows.Next() {
		err := rows.Scan(&dID, &freq)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("----")
		fmt.Println(dID, freq)

		if err := models.DB.Where("id = ?", dID).Find(&doc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found!"})
			return
		}

		models.DB.Preload("Schedule").Find(&doc)
		result = append(result, AutoGenerated{dID, doc, int(freq)})

	}

	if result == nil {
		c.JSON(http.StatusOK, gin.H{"error": "No doctors have 6+ hours for this day!"})
	} else {

		c.JSON(http.StatusOK, gin.H{"doctors with 6+ hours on this day": result})
	}
}

// GET /doctors/:id
// Find a doctor
func FindDoctor(c *gin.Context) {

	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)

	c.JSON(http.StatusOK, gin.H{"doctor": doc})
}

//list of appointmnes
func FindDoctorSchedule(c *gin.Context) {
	var schedule []DoctorSchedule
	var doc models.Doctor

	var patient models.Patient
	role := c.GetString("role")

	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)

	if role != "patient" {
		for i := 0; i < len(doc.Schedule); i++ {

			models.DB.Raw("SELECT name, email FROM patients WHERE id=?", doc.Schedule[i].PID).First(&patient)

			schd := DoctorSchedule{doc.Schedule[i], patient}
			schedule = append(schedule, schd)
		}
		c.JSON(http.StatusOK, gin.H{"schedule": schedule})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"schedule": doc.Schedule})
		return
	}

}

//find availabitlies of ALLLL doctors
func FindDoctorAvailAll(c *gin.Context) {
	var docs []models.Doctor
	var result []D
	var input Given
	var givenSched []models.Appointment

	role := c.GetString("role")

	if role != "admin" && role != "patient" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only clinic admins and patients are authorized to see ALL doctors' availabilities!"})
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Preload("Schedule").Find(&docs)
	year, month, day := input.GivenDay.Date()

	dayStart := time.Date(year, month, day, 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(year, month, day, 17, 0, 0, 0, time.UTC)

	for j := 0; j < len(docs); j++ {
		models.DB.Raw("SELECT * FROM appointments WHERE appointments.d_id=? AND DATE(appointments.start_time)=? AND appointments.deleted_at IS NULL order by start_time asc", docs[j].ID, string(input.GivenDay.Format("2006-01-02"))).Find(&givenSched)

		if len(givenSched) == 0 { //if i have NO appointments, the doc is available from 9 to 5
			fmt.Println("givensched is zero")
			docs[j].Availability = nil
			docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+dayEnd.String()))
		} else if len(givenSched) == 1 {
			fmt.Println("givensched is one")
			if !dayStart.After(givenSched[0].StartTime) {
				fmt.Println("givensched entered at [0]")
				docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+(givenSched[0]).StartTime.String()))
				docs[j].Availability = append(docs[j].Availability, (givenSched[0]).EndTime.String()+"---"+string(dayEnd.String()))
			}
		} else if len(givenSched) > 1 {
			fmt.Println("givensched is greater than 1")
			for i := 0; i < len(givenSched); i++ {
				t := (givenSched[i]).StartTime
				t1 := (givenSched[i]).EndTime //end time of one appointment

				if i == len(givenSched)-1 { //handle final element here...
					docs[j].Availability = append(docs[j].Availability, string(t1.String()+"---"+string(dayEnd.String())))
					break
				}

				t2 := (givenSched[i+1]).StartTime //start time of next appointment
				if i == 0 {
					docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+t.String()))
					docs[j].Availability = append(docs[j].Availability, string(t1.String()+"---"+t2.String()))
				} else {
					docs[j].Availability = append(docs[j].Availability, string(t1.String()+"---"+t2.String()))
				}

			}

		}

		models.DB.Model(&docs[j]).Select("availability").Updates(docs[j].Availability)

		result = append(result, D{int(docs[j].ID), docs[j].Availability})
	}

	c.JSON(http.StatusOK, gin.H{"availability": result})
}

//POST to give it a date!
func FindDoctorAvailability(c *gin.Context) { //ISSUE!!!! input date is showing 0001-01-01
	var doc models.Doctor
	var input Given
	var givenSched []models.Appointment
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)
	year, month, day := input.GivenDay.Date()

	dayStart := time.Date(year, month, day, 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(year, month, day, 17, 0, 0, 0, time.UTC)

	models.DB.Raw("SELECT * FROM appointments WHERE appointments.d_id=? AND DATE(appointments.start_time)=? AND appointments.deleted_at IS NULL order by start_time asc", c.Param("id"), string(input.GivenDay.Format("2006-01-02"))).Find(&givenSched)

	if len(givenSched) == 0 { //if i have NO appointments, the doc is available from 9 to 5
		fmt.Println("givensched is zero")
		doc.Availability = nil
		doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+dayEnd.String()))
	} else if len(givenSched) == 1 {
		fmt.Println("givensched is one")
		if !dayStart.After(givenSched[0].StartTime) {
			fmt.Println("givensched entered at [0]")
			doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+(givenSched[0]).StartTime.String()))
			doc.Availability = append(doc.Availability, (givenSched[0]).EndTime.String()+"---"+string(dayEnd.String()))
		}
	} else if len(givenSched) > 1 {
		fmt.Println("givensched is greater than 1")
		for i := 0; i < len(givenSched); i++ {
			t := (givenSched[i]).StartTime
			t1 := (givenSched[i]).EndTime //end time of one appointment

			if i == len(givenSched)-1 { //handle final element here...
				doc.Availability = append(doc.Availability, string(t1.String()+"---"+string(dayEnd.String())))
				break
			}

			t2 := (givenSched[i+1]).StartTime //start time of next appointment
			if i == 0 {
				doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+t.String()))
				doc.Availability = append(doc.Availability, string(t1.String()+"---"+t2.String()))
			} else {
				doc.Availability = append(doc.Availability, string(t1.String()+"---"+t2.String()))
			}

		}

	}

	models.DB.Model(&doc).Select("availability").Updates(doc.Availability)
	c.JSON(http.StatusOK, gin.H{"availability": doc.Availability})
}

// PATCH /doctors/:id
// Update a doctor
func UpdateDoctor(c *gin.Context) {
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	var input UpdateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&doc).Updates(input)

	c.JSON(http.StatusOK, gin.H{"data": doc})
}

// DELETE /doctors/:id
// Delete a doctor
func DeleteDoctor(c *gin.Context) {
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Delete(&doc)

	c.JSON(http.StatusOK, gin.H{"data": true})
}
