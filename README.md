# Clinic Booking System REST API using Gin and Gorm

This is the code submission for Assessment #1 at CX using Golang, GinGonic and GORM.

How to use:

```
$ go run .
```

| ENDPOINT                    | Requirement   | Access     |
| -------------               | ------------- | -------- |
| GET("/doctors")             | Test1         | NewYork  |
| GET("/doctors/:id")                     | Test2         | Toronto  |
| GET("/doctors/:id/schedule")                 | Returns list of appointments along with associated patient         | All, but patients cannot see any patient info  
| POST("/doctors/:id/slots")                     | Test2         | Toronto  |
| PATCH("/doctors/:id")                    | Test2         | Toronto  |
| DELETE("/doctors/:id" )                 | Test2         | Toronto  |
| POST("/doctors/most/count")                    | Takes a given day (UTC), returns sorted list of doctors with most appointments         | Admins only  |
| POST("/doctors/most/hours")                     | Takes a given day (UTC), returns sorted list of doctors with most 6+ hours          | Admins only  |
| POST("/doctors/slots/all")                     | Takes a given day (UTC), returns slots of all doctors on that day      | Admins and patients only  
| GET("/patients")             | Test1         | NewYork  |
