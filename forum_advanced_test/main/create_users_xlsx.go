package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func main() {
	// Create a new Excel file
	f := excelize.NewFile()

	// Create a sheet named "Users"
	const sheetName = "Users"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// Write header row
	f.SetCellValue(sheetName, "A1", "Email")
	f.SetCellValue(sheetName, "B1", "Username")
	f.SetCellValue(sheetName, "C1", "Password")

	// Sample data
	users := []struct {
		Email    string
		Username string
		Password string
	}{
		{"test1@example.com", "testuser1", "pass12345"},
		{"test2@example.com", "testuser2", "passabcd"},
		{"admin@example.com", "adminuser", "adminSecret"},
	}

	// Fill data row by row
	for i, u := range users {
		row := i + 2 // data starts on row 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), u.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), u.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), u.Password)
	}

	// Save the file
	if err := f.SaveAs("users.xlsx"); err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Println("Successfully created users.xlsx with sample data.")
}
