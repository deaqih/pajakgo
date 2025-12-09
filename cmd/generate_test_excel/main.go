package main

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Transaction Data"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		fmt.Printf("Error creating sheet: %v\n", err)
		return
	}

	// Set headers
	headers := []string{
		"Document Type", "Document Number", "Posting Date", "Account",
		"Account Name", "Keterangan", "Debet", "Credit", "Net",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Test Data untuk propagation
	// IMPORTANT: Account codes HARUS sesuai dengan accounts di master data yang punya koreksi_obyek tertentu
	// Berdasarkan data accounts di database:
	// - 21080000 = VAT OUT VAT PAYABLE (koreksi_obyek: 'PK Cr')
	// - 21070600 = WITHHOLDING PPh ART 21 (koreksi_obyek: 'Wth 21 Cr')
	// - 21070700 = WITHHOLDING PPh ART 23 (2%) (koreksi_obyek: 'Wth 23 Cr')
	// - 21070900 = WITHHOLDING PPh ART 26 (koreksi_obyek: 'Wth 26 Cr')
	// - 21070100 = WITHHOLDING PPh ART 4 CONSTRUCTION SERV (koreksi_obyek: 'Wth 4.2 Cr')
	// - 21070500 = WITHHOLDING PPh ART 15 AIR SERVICE 1.80 (koreksi_obyek: 'Wth 15 Cr')
	testData := [][]interface{}{
		// Document INV-001 - Test PK Cr propagation
		{"Invoice", "INV-001", "2024-12-01", "21080000", "VAT OUT VAT PAYABLE", "Pajak Keluaran", 0, 1100000, 1100000},
		{"Invoice", "INV-001", "2024-12-01", "41010100", "FINISHED GOODS SALES TO DOMESTIC AFFILI", "Penjualan laptop", 0, 10000000, 10000000},
		{"Invoice", "INV-001", "2024-12-01", "11010002", "CASH ON HAND IDR", "Penerimaan tunai", 11100000, 0, 11100000},

		// Document INV-002 - Test Wth 21 Cr propagation
		{"Invoice", "INV-002", "2024-12-02", "21070600", "WITHHOLDING PPh ART 21", "Potongan PPh 21", 0, 500000, 500000},
		{"Invoice", "INV-002", "2024-12-02", "11010002", "CASH ON HAND IDR", "Pembayaran gaji", 10000000, 0, 10000000},
		{"Invoice", "INV-002", "2024-12-02", "43010000", "DIRECT SALARIES", "Gaji karyawan", 9500000, 0, 9500000},

		// Document INV-003 - Test Wth 23 Cr propagation
		{"Invoice", "INV-003", "2024-12-03", "21070700", "WITHHOLDING PPh ART 23 (2%)", "Potongan PPh 23", 0, 230000, 230000},
		{"Invoice", "INV-003", "2024-12-03", "11010002", "CASH ON HAND IDR", "Pembayaran jasa", 5000000, 0, 5000000},
		{"Invoice", "INV-003", "2024-12-03", "43260003", "PROFESIONAL FEE", "Jasa konsultasi", 4770000, 0, 4770000},

		// Document INV-004 - Test Wth 26 Cr propagation
		{"Invoice", "INV-004", "2024-12-04", "21070900", "WITHHOLDING PPh ART 26", "Potongan PPh 26", 0, 260000, 260000},
		{"Invoice", "INV-004", "2024-12-04", "11010002", "CASH ON HAND IDR", "Pembayaran jasa LN", 5000000, 0, 5000000},
		{"Invoice", "INV-004", "2024-12-04", "43260003", "PROFESIONAL FEE", "Jasa luar negeri", 4740000, 0, 4740000},

		// Document INV-005 - Test Wth 4.2 Cr propagation
		{"Invoice", "INV-005", "2024-12-05", "21070100", "WITHHOLDING PPh ART 4 CONSTRUCTION SERV", "Potongan PPh 4.2", 0, 420000, 420000},
		{"Invoice", "INV-005", "2024-12-05", "11010002", "CASH ON HAND IDR", "Pembayaran konstruksi", 10000000, 0, 10000000},
		{"Invoice", "INV-005", "2024-12-05", "16090001", "CONSTRUCTION IN PROGRESS - BUILDINGS", "Jasa konstruksi", 9580000, 0, 9580000},

		// Document INV-006 - Test Wth 15 Cr propagation
		{"Invoice", "INV-006", "2024-12-06", "21070500", "WITHHOLDING PPh ART 15 AIR SERVICE 1.80", "Potongan PPh 15", 0, 150000, 150000},
		{"Invoice", "INV-006", "2024-12-06", "11010002", "CASH ON HAND IDR", "Pembayaran pelayaran", 5000000, 0, 5000000},
		{"Invoice", "INV-006", "2024-12-06", "43230100", "TRANSPORTATION FREIGHT", "Jasa pelayaran", 4850000, 0, 4850000},

		// Document INV-007 - Test PM DB propagation (mencari account dengan PM DB flag)
		// Note: PM DB biasanya untuk VAT IN - perlu cek apakah ada di data
		{"Invoice", "INV-007", "2024-12-07", "11010002", "CASH ON HAND IDR", "Pembayaran pembelian", 11000000, 0, 11000000},
		{"Invoice", "INV-007", "2024-12-07", "21020002", "TRADE ACC. PAYABLE LOCAL - NON AFFILIAT", "Hutang supplier", 0, 10000000, 10000000},
		{"Invoice", "INV-007", "2024-12-07", "41010100", "FINISHED GOODS SALES TO DOMESTIC AFFILI", "Pembelian barang", 10000000, 0, 10000000},

		// Document INV-008 - Test Multiple flags (PK Cr + Wth 21 Cr)
		{"Invoice", "INV-008", "2024-12-08", "21080000", "VAT OUT VAT PAYABLE", "Pajak Keluaran", 0, 1100000, 1100000},
		{"Invoice", "INV-008", "2024-12-08", "21070600", "WITHHOLDING PPh ART 21", "Potongan PPh 21", 0, 500000, 500000},
		{"Invoice", "INV-008", "2024-12-08", "41010100", "FINISHED GOODS SALES TO DOMESTIC AFFILI", "Penjualan", 0, 15000000, 15000000},
		{"Invoice", "INV-008", "2024-12-08", "11010002", "CASH ON HAND IDR", "Penerimaan", 16600000, 0, 16600000},

		// Document INV-009 - Test different document numbers (koreksi dan obyek per row)
		{"Invoice", "INV-009", "2024-12-09", "41010100", "FINISHED GOODS SALES TO DOMESTIC AFFILI", "Penjualan A", 0, 5000000, 5000000},
		{"Invoice", "INV-010", "2024-12-09", "41010200", "FINISHED GOODS SALES TO DOMESTIC NON AF", "Penjualan B", 0, 3000000, 3000000},
		{"Invoice", "INV-011", "2024-12-09", "43010000", "DIRECT SALARIES", "Gaji karyawan", 5000000, 0, 5000000},
	}

	// Write test data
	for rowIdx, rowData := range testData {
		row := rowIdx + 2
		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 40)
	f.SetColWidth(sheetName, "F", "F", 30)
	f.SetColWidth(sheetName, "G", "G", 15)
	f.SetColWidth(sheetName, "H", "H", 15)
	f.SetColWidth(sheetName, "I", "I", 15)

	// Add instructions sheet
	instructionsSheet := "Instructions"
	instIndex, _ := f.NewSheet(instructionsSheet)
	f.SetCellValue(instructionsSheet, "A1", "TEST DATA INSTRUCTIONS")
	f.SetCellValue(instructionsSheet, "A3", "File ini dibuat untuk testing propagation logic")
	f.SetCellValue(instructionsSheet, "A5", "Test Cases:")
	f.SetCellValue(instructionsSheet, "A6", "1. INV-001: Test PK Cr propagation (account 21080000)")
	f.SetCellValue(instructionsSheet, "A7", "2. INV-002: Test Wth 21 Cr propagation (account 21070600)")
	f.SetCellValue(instructionsSheet, "A8", "3. INV-003: Test Wth 23 Cr propagation (account 21070700)")
	f.SetCellValue(instructionsSheet, "A9", "4. INV-004: Test Wth 26 Cr propagation (account 21070900)")
	f.SetCellValue(instructionsSheet, "A10", "5. INV-005: Test Wth 4.2 Cr propagation (account 21070100)")
	f.SetCellValue(instructionsSheet, "A11", "6. INV-006: Test Wth 15 Cr propagation (account 21070500)")
	f.SetCellValue(instructionsSheet, "A12", "7. INV-007: Test transaction without flag")
	f.SetCellValue(instructionsSheet, "A13", "8. INV-008: Test multiple flags (PK Cr + Wth 21 Cr)")
	f.SetCellValue(instructionsSheet, "A14", "9. INV-009-011: Test koreksi & obyek per row (different doc numbers)")
	f.SetCellValue(instructionsSheet, "A16", "ACCOUNT CODES YANG DIGUNAKAN:")
	f.SetCellValue(instructionsSheet, "A17", "- 21080000 -> VAT OUT VAT PAYABLE (koreksi_obyek: 'PK Cr')")
	f.SetCellValue(instructionsSheet, "A18", "- 21070600 -> WITHHOLDING PPh ART 21 (koreksi_obyek: 'Wth 21 Cr')")
	f.SetCellValue(instructionsSheet, "A19", "- 21070700 -> WITHHOLDING PPh ART 23 (koreksi_obyek: 'Wth 23 Cr')")
	f.SetCellValue(instructionsSheet, "A20", "- 21070900 -> WITHHOLDING PPh ART 26 (koreksi_obyek: 'Wth 26 Cr')")
	f.SetCellValue(instructionsSheet, "A21", "- 21070100 -> WITHHOLDING PPh ART 4 CONSTRUCTION (koreksi_obyek: 'Wth 4.2 Cr')")
	f.SetCellValue(instructionsSheet, "A22", "- 21070500 -> WITHHOLDING PPh ART 15 AIR SERVICE (koreksi_obyek: 'Wth 15 Cr')")
	f.SetCellValue(instructionsSheet, "A23", "")
	f.SetCellValue(instructionsSheet, "A24", "Account codes ini sudah ada di database Anda!")

	f.SetActiveSheet(instIndex)
	f.DeleteSheet("Sheet1")

	// Save first file
	outputPath1 := filepath.Join("storage", "uploads", "test_propagation_data.xlsx")
	if err := f.SaveAs(outputPath1); err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		return
	}

	fmt.Printf("✓ Test file 1 created: %s\n", outputPath1)
	fmt.Printf("  Total rows: %d\n", len(testData))

	// Create second file for multiple file upload test
	f2 := excelize.NewFile()
	defer f2.Close()

	sheetName2 := "Transaction Data"
	_, err = f2.NewSheet(sheetName2)
	if err != nil {
		fmt.Printf("Error creating sheet 2: %v\n", err)
		return
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f2.SetCellValue(sheetName2, cell, header)
	}

	// Set header style
	f2.SetCellStyle(sheetName2, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Additional test data for file 2 - menggunakan account codes yang benar dari database
	testData2 := [][]interface{}{
		{"Invoice", "INV-012", "2024-12-10", "21080000", "VAT OUT VAT PAYABLE", "Pajak Keluaran 2", 0, 2200000, 2200000},
		{"Invoice", "INV-012", "2024-12-10", "41010100", "FINISHED GOODS SALES TO DOMESTIC AFFILI", "Penjualan produk", 0, 20000000, 20000000},
		{"Payment", "PAY-002", "2024-12-10", "11010002", "CASH ON HAND IDR", "Pengeluaran tunai", 5000000, 0, 5000000},
		{"Payment", "PAY-002", "2024-12-10", "21020002", "TRADE ACC. PAYABLE LOCAL - NON AFFILIAT", "Bayar hutang", 0, 5000000, 5000000},
	}

	// Write test data 2
	for rowIdx, rowData := range testData2 {
		row := rowIdx + 2
		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f2.SetCellValue(sheetName2, cell, value)
		}
	}

	// Set column widths
	f2.SetColWidth(sheetName2, "A", "A", 15)
	f2.SetColWidth(sheetName2, "B", "B", 20)
	f2.SetColWidth(sheetName2, "C", "C", 15)
	f2.SetColWidth(sheetName2, "D", "D", 15)
	f2.SetColWidth(sheetName2, "E", "E", 40)
	f2.SetColWidth(sheetName2, "F", "F", 30)
	f2.SetColWidth(sheetName2, "G", "G", 15)
	f2.SetColWidth(sheetName2, "H", "H", 15)
	f2.SetColWidth(sheetName2, "I", "I", 15)

	f2.DeleteSheet("Sheet1")

	// Save second file
	outputPath2 := filepath.Join("storage", "uploads", "test_propagation_data_2.xlsx")
	if err := f2.SaveAs(outputPath2); err != nil {
		fmt.Printf("Error saving file 2: %v\n", err)
		return
	}

	fmt.Printf("✓ Test file 2 created: %s\n", outputPath2)
	fmt.Printf("  Total rows: %d\n", len(testData2))
	fmt.Printf("\nTotal test data: %d rows across 2 files\n", len(testData)+len(testData2))
}

func getColumnName(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+(index%26))) + result
		index = index/26 - 1
		if index < 0 {
			break
		}
	}
	return result
}

