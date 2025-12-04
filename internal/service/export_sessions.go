package service

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportSessionsList exports upload sessions list to Excel
func (s *ExcelService) ExportSessionsList(sessions []map[string]interface{}, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Create sheet for sessions
	sheetName := "Upload Sessions"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// Set headers
	headers := []string{
		"ID", "Session Code", "User ID", "Filename", "Total Rows",
		"Processed", "Failed", "Status", "Error Message", "Created At", "Updated At",
	}

	// Style headers
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Style data rows
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})

	// Write data
	for i, session := range sessions {
		row := i + 2

		// Write each cell
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), session["ID"])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), session["Session Code"])
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), session["User ID"])
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), session["Filename"])
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), session["Total Rows"])
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), session["Processed"])
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), session["Failed"])
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), session["Status"])
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), session["Error Message"])
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), session["Created At"])
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), session["Updated At"])

		// Apply style to entire row
		for j := 0; j < len(headers); j++ {
			cell := fmt.Sprintf("%s%d", string(rune('A'+j)), row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}

		// Add status-based styling
		if status, ok := session["Status"].(string); ok {
			statusCell := fmt.Sprintf("H%d", row)
			switch status {
			case "completed":
				statusStyle, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#D4EDDA"},
						Pattern: 1,
					},
				})
				f.SetCellStyle(sheetName, statusCell, statusCell, statusStyle)
			case "failed":
				statusStyle, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#F8D7DA"},
						Pattern: 1,
					},
				})
				f.SetCellStyle(sheetName, statusCell, statusCell, statusStyle)
			case "processing":
				statusStyle, _ := f.NewStyle(&excelize.Style{
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#FFF3CD"},
						Pattern: 1,
					},
				})
				f.SetCellStyle(sheetName, statusCell, statusCell, statusStyle)
			}
		}
	}

	// Auto-fit column widths
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15) // Default width
	}

	// Set specific column widths
	f.SetColWidth(sheetName, "B", "B", 20) // Session Code
	f.SetColWidth(sheetName, "D", "D", 25) // Filename
	f.SetColWidth(sheetName, "I", "I", 30) // Error Message
	f.SetColWidth(sheetName, "J", "J", 20) // Created At
	f.SetColWidth(sheetName, "K", "K", 20) // Updated At

	// Add summary at the bottom
	if len(sessions) > 0 {
		summaryRow := len(sessions) + 3
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "Summary:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), fmt.Sprintf("Total Sessions: %d", len(sessions)))

		// Count by status
		statusCounts := make(map[string]int)
		for _, session := range sessions {
			if status, ok := session["Status"].(string); ok {
				statusCounts[status]++
			}
		}

		row := summaryRow + 1
		for status, count := range statusCounts {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%s: %d", strings.Title(status), count))
			row++
		}

		// Style summary section
		summaryStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#F0F0F0"},
				Pattern: 1,
			},
		})
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), summaryStyle)
	}

	f.DeleteSheet("Sheet1")

	return f.SaveAs(outputPath)
}