package controllers

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rahmanfadhil/gin-bookstore/models"
	"golang.org/x/crypto/bcrypt"
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

// GET /books
// Find all books
func FindDoctors(c *gin.Context) { //take the gin context... revise this
	var docs []models.Doctor
	//	var doc models.Doctor
	var patient models.Patient
	//jsonData, err := c.GetRawData()

	email := c.GetString("email")

	if err := models.DB.Where("email = ?", email).First(&patient).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You need to be either a doctor or clinic admin to access this!"})
		return
	}

	models.DB.Model(&models.Doctor{}).Preload("Schedule").Find(&docs)

	c.JSON(http.StatusOK, gin.H{"doctors": docs})
}

type Given struct {
	GivenDay time.Time `json:"startTime" binding:"required"`
}

type AutoGenerated struct {
	DoctorID     int           `json:"doctorID"`
	Doctor       models.Doctor `json:"doctor"`
	Appointments int           `json:"totalHours"`
}

// most appointments for a given day..
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
		fmt.Println("----")
		fmt.Println(dID, freq)

		if err := models.DB.Where("id = ?", dID).Find(&doc).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found!"})
			return
		}

		models.DB.Preload("Schedule").Find(&doc)
		result = append(result, AutoGenerated{dID, doc, freq})

	}

	c.JSON(http.StatusOK, gin.H{"doctors with most appointments on this day": result})
}

//doctors with 6+ hrs on a given day....
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
	fmt.Println(dup1)

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

// GET /books/:id
// Find a book
func FindDoctor(c *gin.Context) {
	// Get model if exist
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)
	//models.DB.Preload("Availability").Find(&doc)

	c.JSON(http.StatusOK, gin.H{"doctor": doc})
}

type DoctorSchedule struct {
	Appointment models.Appointment
	Patient     models.Patient
}

//list of appointmnes
func FindDoctorSchedule(c *gin.Context) {
	// Get model if exist
	var schedule []DoctorSchedule
	var doc models.Doctor
	var patient models.Patient
	role := c.GetString("role")

	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)

	for i := 0; i < len(doc.Schedule); i++ {
		fmt.Println("helloooooo")

		if role == "patient" {

			models.DB.Raw("SELECT name, email FROM patients WHERE id=?", doc.Schedule[i].PID).Find(&patient)
		} else {
			models.DB.Raw("SELECT * FROM patients WHERE id=?", doc.Schedule[i].PID).First(&patient)
		}

		schd := DoctorSchedule{doc.Schedule[i], patient}
		schedule = append(schedule, schd)
	}

	//b, _ := json.Marshal(schedule)
	//
	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

//find availabitlies of ALLLL doctors
func FindDoctorAvailAll(c *gin.Context) {
	// Get model if exist
	var docs []models.Doctor
	var result []D
	models.DB.Preload("Schedule").Find(&docs)

	role := c.GetString("role")

	if role != "admin" || role != "patient" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only clinic admins and patients are authorized to see ALL doctors' availabilities!"})
		return
	}

	dayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 17, 0, 0, 0, time.UTC)

	//models.DB.Raw("SELECT * cFROM appointments WHERE DATE(start_time) = ? AND d_id=?", string(t), doc.ID).Scan(&count)

	for j := 0; j < len(docs); j++ {
		for i := 0; i < len(docs[j].Schedule); i++ {
			if i == 0 { //first appointment of the day!
				if !dayStart.After(docs[j].Schedule[i].StartTime) && !dayStart.Before(docs[j].Schedule[i].StartTime) {
					docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+(docs[j].Schedule[i]).StartTime.String()))
				}
			}

			if i+1 == len(docs[j].Schedule) { //last appointment of the day
				if !dayStart.Before(docs[j].Schedule[i].EndTime) && !dayStart.After(docs[j].Schedule[i].StartTime) {
					fmt.Println("neee")
					docs[j].Availability = append(docs[j].Availability, string((docs[j].Schedule[i]).EndTime.String()+"---"+dayEnd.String()))
				}

			}

			if len(docs[j].Schedule) > 1 { //if i have two appointments or more, so that i dont go out of bounds!
				if i == len(docs[j].Schedule)-1 {
					//handle final element here...
					break
				}

				t1 := (docs[j].Schedule[i]).EndTime     //end time of one appointment
				t2 := (docs[j].Schedule[i+1]).StartTime //start time of next appointment

				//if dayStart.Unix() - t1.Unix() < int64(time.Hour.Seconds() * 24)

				day1, month1, year1 := dayStart.Date()
				day2, month2, year2 := t1.Date()

				/**

				if dayStart.After(t1) || dayStart.Before(t1) {
					docs[j].Availability = nil
					fmt.Println("yea?")
					docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+dayEnd.String()))
				} **/

				if day1 != day2 && month1 != month2 && year1 != year2 {
					//docs[j].Availability = nil

					fmt.Println("yea?")
					docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+dayEnd.String()))
				} else {
					fmt.Println("nooo")
					docs[j].Availability = append(docs[j].Availability, string(t1.String()+"---"+t2.String()))
				}

				//docs[j].Availability = append(docs[j].Availability, string(t1.String()+"---"+t2.String()))
			}

		}

		if len(docs[j].Schedule) == 0 { //if i have NO appointments, the doc is available from 9 to 5
			//empty availability array in case of cancelled appointments...
			docs[j].Availability = nil
			docs[j].Availability = append(docs[j].Availability, string(dayStart.String()+"---"+dayEnd.String()))
		}

		result = append(result, D{int(docs[j].ID), docs[j].Availability})
	}

	//	models.DB.Model(&docs).Select("availability").Updates(doc.Availability)

	c.JSON(http.StatusOK, gin.H{"availability": result})
}

type D struct {
	Doctor       int      `json:"doctorID"`
	Availability []string `gorm:"type:text" json:"availability"`
}

//
func FindDoctorAvailability(c *gin.Context) {
	// Get model if exist
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Preload("Schedule").Find(&doc)

	dayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 0, 0, 0, time.UTC)
	dayEnd := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 17, 0, 0, 0, time.UTC)

	for i := 0; i < len(doc.Schedule); i++ {
		if i == 0 { //first appointment of the day!

			if !dayStart.After(doc.Schedule[i].StartTime) {
				doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+(doc.Schedule[i]).StartTime.String()))
			}
		}

		if i+1 == len(doc.Schedule) { //last appointment of the day
			if !dayStart.Before(doc.Schedule[i].EndTime) {
				fmt.Println("neee")
				doc.Availability = append(doc.Availability, string((doc.Schedule[i]).EndTime.String()+"---"+dayEnd.String()))
			}

		}

		if len(doc.Schedule) > 1 { //if i have two appointments or more, so that i dont go out of bounds!
			if i == len(doc.Schedule)-1 {
				//handle final element here...
				break
			}
			t1 := (doc.Schedule[i]).EndTime     //end time of one appointment
			t2 := (doc.Schedule[i+1]).StartTime //start time of next appointment

			if dayStart.After(t1) || dayStart.Before(t1) || dayEnd.Before(t2) || dayEnd.After(t2) {
				doc.Availability = nil
				fmt.Println("yea?")
				doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+dayEnd.String()))
			} else {
				fmt.Println("nooo")
				doc.Availability = append(doc.Availability, string(t1.String()+"---"+t2.String()))
			}

		}

	}

	if len(doc.Schedule) == 0 { //if i have NO appointments, the doc is available from 9 to 5
		//empty availability array in case of cancelled appointments...
		doc.Availability = nil
		doc.Availability = append(doc.Availability, string(dayStart.String()+"---"+dayEnd.String()))
	}

	models.DB.Model(&doc).Select("availability").Updates(doc.Availability)

	c.JSON(http.StatusOK, gin.H{"availability": doc.Availability})
}

// POST /books
// Create new book
func CreateDoctor(c *gin.Context) {
	// Validate input

	var input CreateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create book
	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	doc := models.Doctor{Name: input.Name, Email: input.Email, Password: string(hashed), Role: "doctor"}
	models.DB.Create(&doc)

	//////////

	c.JSON(http.StatusOK, gin.H{"doctor": doc})

}

// PATCH /books/:id
// Update a book
func UpdateDoctor(c *gin.Context) {
	// Get model if exist
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	// Validate input
	var input UpdateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&doc).Updates(input)

	c.JSON(http.StatusOK, gin.H{"data": doc})
}

// DELETE /books/:id
// Delete a book
func DeleteDoctor(c *gin.Context) {
	// Get model if exist
	var doc models.Doctor
	if err := models.DB.Where("id = ?", c.Param("id")).First(&doc).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	models.DB.Delete(&doc)

	c.JSON(http.StatusOK, gin.H{"data": true})
}
