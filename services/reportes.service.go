package services

import (
	"fmt"
	"log"
	"sort"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/xuri/excelize/v2"
)

func GenerarReporteCodigos(idPeriodo int64, idProyecto int64) {
	//Mapa para guardar los admitidos
	var admitidos []map[string]interface{}
	var errGetAll = false

	//Obtener Datos del periodo
	var periodo map[string]interface{}
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+fmt.Sprintf("periodo/%v", idPeriodo), &periodo)
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "[map[]]" {
		errGetAll = true
	}

	//Obtener Datos del proyecto & facultad
	var facultad map[string]interface{}

	var proyecto map[string]interface{}
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_institucion/%v", idProyecto), &proyecto)
	if errProyecto != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		errGetAll = true
	} else {

		//Obtener Datos de la facultad
		errFacultad := request.GetJson("http://"+beego.AppConfig.String("OikosService")+fmt.Sprintf("dependencia/%v", proyecto["FacultadId"]), &facultad)
		if errFacultad != nil || fmt.Sprintf("%v", periodo) == "map[]" {
			errGetAll = true
		}
	}

	//Inscripciones de admitidos
	var inscripciones []map[string]interface{}
	fmt.Println("http://" + beego.AppConfig.String("InscripcionService") + fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:ADMITIDO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", idProyecto, idPeriodo))
	errInscripciones := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:ADMITIDO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", idProyecto, idPeriodo), &inscripciones)
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "map[]" {
		errGetAll = true
	}

	if !errGetAll {

		//Base para la comparación de codigo
		if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "3" {
			periodo["Data"].(map[string]interface{})["Ciclo"] = "2"
		}
		codigoBase := fmt.Sprintf("%v%v%v", periodo["Data"].(map[string]interface{})["Year"], periodo["Data"].(map[string]interface{})["Ciclo"], proyecto["Codigo"])

		for _, inscripcion := range inscripciones {

			//Obtener Datos basicos Tercero
			var tercero []map[string]interface{}
			errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscripcion["PersonaId"]), &tercero)
			if errTercero != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {

				errGetAll = true
				break
			}

			//Obtener Documento Tercero
			var terceroDocumento []map[string]interface{}
			errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:%v", inscripcion["PersonaId"]), &terceroDocumento)
			if errTerceroDocumento != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {

				errGetAll = true
				break
			}

			//Obtener Codigo Tercero
			var terceroCodigo []map[string]interface{}
			errTerceroCodigo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CODE,Activo:true,TerceroId:%v,Numero__contains:%v", inscripcion["PersonaId"], codigoBase), &terceroCodigo)
			if errTerceroCodigo != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {

				errGetAll = true
				break
			}

			admitidos = append(admitidos, map[string]interface{}{
				"Nombre":          fmt.Sprintf("%v %v", tercero[0]["PrimerNombre"], tercero[0]["SegundoNombre"]),
				"PrimerApellido":  tercero[0]["PrimerApellido"],
				"SegundoApellido": tercero[0]["SegundoApellido"],
				"Estado":          "Admitido",
				"Documento":       terceroDocumento[0]["Numero"],
				"Codigo":          terceroCodigo[0]["Numero"],
			})
		}
	}

	if !errGetAll {

		//Añadir información de la cabecera de el excel
		infoCabecera := map[string]interface{}{
			"Facultad": facultad["Nombre"],
			"Proyecto": proyecto["Nombre"],
			"Periodo":  periodo["Data"].(map[string]interface{})["Nombre"],
		}

		//Función que genera el reporte en xlsx
		GenerarExcelReporteCodigos(admitidos, infoCabecera)
	} else {
		fmt.Println("Fallo")
	}

}

func GenerarExcelReporteCodigos(admitidosMap []map[string]interface{}, infoCabecera map[string]interface{}) {

	var admitidos [][]interface{}

	//Organizar por códigos
	sort.Slice(admitidosMap, func(i, j int) bool {
		return admitidosMap[i]["Codigo"].(string) < admitidosMap[j]["Codigo"].(string)
	})

	for i, admitido := range admitidosMap {
		fila := []interface{}{
			i + 1,
			admitido["PrimerApellido"],
			admitido["SegundoApellido"],
			admitido["Nombre"],
			admitido["Estado"],
			admitido["Documento"],
			admitido["Codigo"],
		}
		admitidos = append(admitidos, fila)
	}

	file, err := excelize.OpenFile("static/templates/PruebaExcel.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	var lastCell = ""
	for i, row := range admitidos {
		dataRow := i + 7
		for j, col := range row {
			file.SetCellValue("Hoja1", fmt.Sprintf("%s%d", string(rune(65+j)), dataRow), col)
			lastCell = fmt.Sprintf("%s%d", string(rune(65+j)), dataRow)
		}
	}

	file.SetCellValue("Hoja1", "B4", fmt.Sprintf("Facultad: %v", infoCabecera["Facultad"]))
	file.SetCellValue("Hoja1", "D4", fmt.Sprintf("Proyecto: %v", infoCabecera["Proyecto"]))
	file.SetCellValue("Hoja1", "F4", fmt.Sprintf("Periodo: %v", infoCabecera["Periodo"]))

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
