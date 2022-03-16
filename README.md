# Clinic Booking System REST API using Gin and Gorm

This is the code submission for Assessment #1 at CX using Golang, GinGonic and GORM.

How to use:

```
$ go run .
```

| ENDPOINT                    | Requirement   | Access     |
| -------------               | ------------- | -------- |
| GET("/doctors")             | Returns list of all doctors in system         | All, but patients and doctors can only see name and email, while admins can see everything else  |
| GET("/doctors/:id")                     | Returns record of requested doctor         | All, but patients and doctors can only see name and email, while admins can see everything else  |
| GET("/doctors/:id/schedule")                 | Returns list of appointments along with associated patient         | All, but patients cannot see any patient info  
| POST("/doctors/:id/slots")                     | Takes a given day (UTC), returns string array of availabilities in a day for a requested doctor        | All, no restrictions  |
| PATCH("/doctors/:id")                    | Not required (part of CRUD)        | N/A  |
| DELETE("/doctors/:id" )                 | Not required (part of CRUD)         | Admins only  |
| POST("/doctors/most/count")                    | Takes a given day (UTC), returns sorted list of doctors with most appointments         | Admins only  |
| POST("/doctors/most/hours")                     | Takes a given day (UTC), returns sorted list of doctors with most 6+ hours          | Admins only  |
| POST("/doctors/slots/all")                     | Takes a given day (UTC), returns slots of all doctors on that day      | Admins and patients only  
| GET("/patients")             | Returns list of all patients in system         | Admins only  |
