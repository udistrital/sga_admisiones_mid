package services

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func GenerarReporteCodigos(idPeriodo int64, idProyecto int64) {

	data := [][]interface{}{
		{1, "John", 30},
		{2, "Alex", 20},
		{3, "Bob", 40},
	}

	file, err := excelize.OpenFile("static/templates/PruebaExcel.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	var lastCell = ""
	for i, row := range data {
		dataRow := i + 7
		for j, col := range row {
			file.SetCellValue("Hoja1", fmt.Sprintf("%s%d", string(rune(65+j)), dataRow), col)
			lastCell = fmt.Sprintf("%s%d", string(rune(65+j)), dataRow)
		}
	}

	file.SetCellValue("Hoja1", "B4", "FACULTAD: FACULTAD ARTES ASAB")
	file.SetCellValue("Hoja1", "D4", "PROYECTO CURRICULAR: MAESTRIA ESTUDIOS ARTISTICOS")
	file.SetCellValue("Hoja1", "F4", "PERIODO: 202-3")

	style, err := file.NewStyle(
		&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "00000000", Style: 1},
				{Type: "right", Color: "00000000", Style: 1},
				{Type: "top", Color: "00000000", Style: 1},
				{Type: "bottom", Color: "00000000", Style: 1},
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#d9e1f2"},
				Pattern: 1,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("LastCell: " + lastCell)
	file.SetCellStyle("Hoja1", "A7", lastCell, style)

	if err := file.SaveAs("static/templates/Modificado.xlsx"); err != nil {
		log.Fatal(err)
	}

}
