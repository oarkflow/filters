package main

import (
	"fmt"
	"time"
)

// Function to calculate age based on date of birth (dob)
func calculateAge(dob time.Time) int {
	now := time.Now()
	age := now.Year() - dob.Year()

	// Adjust if birthday hasn't occurred yet this year
	if now.YearDay() < dob.YearDay() {
		age--
	}
	return age
}

// Function to check if age is between min and max using both expressions
func checkAgeWithConditions(dob time.Time, minAge, maxAge int) bool {
	// Calculate the age
	age := calculateAge(dob)

	// Expression 1: "age BETWEEN minAge AND maxAge"
	expr1 := age >= minAge && age <= maxAge

	// Expression 2: "minAge <= age AND maxAge >= age"
	expr2 := minAge <= age && maxAge >= age

	// Return true only if both expressions hold
	return expr1 && expr2
}

func main() {
	// Example data
	patientDOB := time.Date(1990, time.August, 15, 0, 0, 0, 0, time.UTC)
	minAge := 25
	maxAge := 40

	// Check if age is within range using both expressions
	isValidAge := checkAgeWithConditions(patientDOB, minAge, maxAge)

	if isValidAge {
		fmt.Println("The patient's age is within the specified range.")
	} else {
		fmt.Println("The patient's age is NOT within the specified range.")
	}
}
