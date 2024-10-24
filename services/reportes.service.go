package services

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/phpdave11/gofpdf"
	"github.com/udistrital/sga_admisiones_mid/helpers"
	"github.com/udistrital/sga_admisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/xlsx2pdf"
	"github.com/xuri/excelize/v2"
)

func safeFirstChar(s string) string {
	if len(s) > 0 {
		return string(s[0])
	}
	return ""
}

func ListarDataInscripcionEvaluacion(dataOrganizada []map[string]interface{}, requisitosOrganizada []map[string]interface{}) map[string]interface{} {

	indx := 7
	index := 1
	colIndex := 0 // Inicializa el índice de columna
	CantidadColumnas := 0

	file, err := excelize.OpenFile("static/templates/ListadoInscripcionEvaluacion.xlsx")
	if err != nil {
		log.Fatal(err)
		return map[string]interface{}{
			"Error": err.Error(),
		}
	}

	style := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
	}

	styleColumn := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#A9A9A9"}, // Gris claro, fondo 2, oscuro 25%
			Pattern: 1,
		},
	}

	styleID, err := file.NewStyle(style)
	if err != nil {
		return map[string]interface{}{
			"Error": err.Error(),
		}
	}

	styleCalumnID, err := file.NewStyle(styleColumn)
	if err != nil {
		return map[string]interface{}{
			"Error": err.Error(),
		}
	}

	file.SetCellValue("Hoja1", "B"+strconv.Itoa(1), " UNIVERSIDAD DISTRITAL FRANCISCO JOSÉ DE CALDAS REPORTE CODIFICACIÓN LISTADO  INSCRIPCION EVALUACIÓN "+requisitosOrganizada[0]["ProyectoAcademico"].(string)+" "+requisitosOrganizada[0]["PeriodoId"].(string))

	for _, requisito := range requisitosOrganizada {
		if porcentajeEspecifico, ok := requisito["PorcentajeEspecifico"].(map[string]interface{}); ok {
			for _, area := range porcentajeEspecifico {
				if areaMap, ok := area.([]interface{}); ok {
					for _, requisitoArea := range areaMap {

						colLetter := string(rune('F' + colIndex)) // Convierte el índice de columna a letra
						if requisitoAreaMap, ok := requisitoArea.(map[string]interface{}); ok {
							cell := colLetter + strconv.Itoa(indx)
							file.SetCellValue("Hoja1", cell, fmt.Sprintf("%v", requisitoAreaMap["Nombre"]))
							CantidadColumnas = colIndex
							colIndex++ // Incrementa el índice de columna

						}

					}
				}
			}
		}
	}

	for _, data := range dataOrganizada {
		indx++

		file.SetCellValue("Hoja1", "A"+strconv.Itoa(indx), index)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(indx), data["TipoDocumento"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(indx), data["Documento"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(indx), data["Nombre"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(indx), data["TipoInscripcionId"])
		for _, dataEvaluacion := range data["DetalleEvaluacion"].([]map[string]interface{}) {
			for _, area := range dataEvaluacion["DetalleCalificacion"].(map[string]interface{}) {
				if areaMap, ok := area.([]interface{}); ok {
					for _, requisitoArea := range areaMap {
						if requisitoAreaMap, ok := requisitoArea.(map[string]interface{}); ok {
							for columna := 0; columna <= CantidadColumnas; columna++ {
								colLetter := string(rune('F' + columna)) // Convierte el índice de columna a letra
								value, err := file.GetCellValue("Hoja1", colLetter+strconv.Itoa(7))
								if err != nil {
									log.Fatal(err)
								}
								for key := range requisitoAreaMap {
									if value == key {
										cell := colLetter + strconv.Itoa(indx)
										file.SetCellValue("Hoja1", cell, requisitoAreaMap[key])

									}
								}

							}

							//Estilos aplicandose
							file.MergeCell("Hoja1", "B"+strconv.Itoa(1), string('F'+CantidadColumnas)+strconv.Itoa(5))
							file.MergeCell("Hoja1", "A"+strconv.Itoa(6), string('F'+CantidadColumnas)+strconv.Itoa(6))
							file.SetCellStyle("Hoja1", "A"+strconv.Itoa(6), string('F'+CantidadColumnas)+strconv.Itoa(6), styleCalumnID) // Combina las celdas
							file.SetCellStyle("Hoja1", "F"+strconv.Itoa(7), string('F'+CantidadColumnas)+strconv.Itoa(7), styleCalumnID) // Aplica el estilo
							file.SetCellStyle("Hoja1", "A"+strconv.Itoa(indx), "E"+strconv.Itoa(indx), styleID)                          // Aplica el estilo
							file.SetCellStyle("Hoja1", "F"+strconv.Itoa(indx), string('F'+CantidadColumnas)+strconv.Itoa(indx), styleID) // Aplica el estilo

						}

					}
					index++
					if err := file.SaveAs("static/templates/ListadoInscripcionEvaluacionDiligenciado.xlsx"); err != nil {
						return map[string]interface{}{
							"Error": err.Error(),
						}
					}
				}
			}
		}
	}

	// Guarda el archivo en un buffer
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		log.Fatal(err)
	}

	// Codifica el contenido del buffer en Base64
	base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	fmt.Println(base64Str)

	return map[string]interface{}{
		"Excel": base64Str,
	}
}

func ListadoInscripcionEvaluacion(idPeriodo int64, idProyecto int64) (APIResponseDTO requestresponse.APIResponse) {

	var ProyectoAcademico map[string]interface{}

	var Periodo map[string]interface{}

	var periodo string

	var requisitos []interface{}

	var requisitosOrganizada []map[string]interface{}

	var inscripciones []interface{}

	var dataOrganizada []map[string]interface{}

	var respuesta []map[string]interface{}

	var documento map[string]interface{}

	var inscripcion []interface{}

	var detalleEvaluacionOrganizada []map[string]interface{}

	//Consulta Periodo
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+"periodo?query=Id:"+strconv.FormatInt(idPeriodo, 10), &Periodo)
	if errPeriodo != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar Periodo: "+errPeriodo.Error())
	}

	if data, ok := Periodo["Data"].([]interface{}); ok && len(data) > 0 {
		if firstData, ok := data[0].(map[string]interface{}); ok {
			periodo = firstData["Nombre"].(string)
		}
	}

	//Consulta Proyecto
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion/"+strconv.FormatInt(idProyecto, 10), &ProyectoAcademico)
	if errProyecto != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar Proyecto Academico: "+errProyecto.Error())
	}

	errRequisitos := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"/requisito_programa_academico?query=PeriodoId:"+strconv.FormatInt(idPeriodo, 10)+",ProgramaAcademicoId:"+strconv.FormatInt(idProyecto, 10)+"&limit=0", &requisitos)
	if errRequisitos != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar Requisitos: "+errRequisitos.Error())
	}

	//Organizar Requisitos
	for _, requisito := range requisitos {
		requisitoMap, ok := requisito.(map[string]interface{})
		if !ok {
			fmt.Println("Error: requisito is not a map")
		}

		var PorcentajeEspecifico map[string]interface{}

		porcentajeGeneralStr, ok := requisitoMap["PorcentajeEspecifico"].(string)
		if !ok {
			fmt.Println("Error: PorcentajeGeneral is not a string")
		}

		if err := json.Unmarshal([]byte(porcentajeGeneralStr), &PorcentajeEspecifico); err != nil {
			fmt.Println("Error deserializando PorcentajeEspecifico:", err)
			continue
		}

		requisitosOrganizada = append(requisitosOrganizada, map[string]interface{}{
			"Nombre":               requisitoMap["RequisitoId"].(map[string]interface{})["Nombre"],
			"PorcentajeGeneral":    requisitoMap["PorcentajeGeneral"],
			"PorcentajeEspecifico": PorcentajeEspecifico,
			"ProyectoAcademico":    ProyectoAcademico["Nombre"],
			"PeriodoId":            periodo,
		})

	}

	//Consulta Inscripciones correspondientes al periodo y proyecto
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=Activo:true,PeriodoId:"+strconv.FormatInt(idPeriodo, 10)+",ProgramaAcademicoId:"+strconv.FormatInt(idProyecto, 10)+"&sortby=Id&order=asc&limit=0", &inscripcion)
	if errInscripcion != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar Inscripciones: "+errInscripcion.Error())
	}

	//Se mapea
	for _, ins := range inscripcion {
		if insMap, ok := ins.(map[string]interface{}); ok && len(insMap) > 0 {
			inscripciones = append(inscripciones, ins)
		}
	}

	//Se consulta informacion de cada inscripcion
	for _, inscripcion := range inscripciones {
		var consultaPorPersona interface{}

		var idInscripcion = strconv.FormatFloat(inscripcion.(map[string]interface{})["Id"].(float64), 'f', -1, 64)

		var detalleEvaluacion interface{}

		//Consulta persona
		errConsultarPersona := request.GetJson("http://"+beego.AppConfig.String("TerceroMid")+"personas/"+strconv.FormatFloat(inscripcion.(map[string]interface{})["PersonaId"].(float64), 'f', -1, 64), &consultaPorPersona)
		if errConsultarPersona != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar persona: "+errConsultarPersona.Error())
		}

		//Consulta detalle evaluacion
		errEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=InscripcionId:"+idInscripcion, &detalleEvaluacion)
		if errEvaluacion != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar evaluacion: "+errEvaluacion.Error())
		}

		for _, detalle := range detalleEvaluacion.([]interface{}) {

			var detalleEvalaucionJson map[string]interface{}
			detalleEvaluacionOrganizada = nil
			detalle := detalle.(map[string]interface{})

			DetalleCalificacionstr := detalle["DetalleCalificacion"].(string)

			if err := json.Unmarshal([]byte(DetalleCalificacionstr), &detalleEvalaucionJson); err != nil {
				fmt.Println("Error deserializando PorcentajeEspecifico:", err)
				continue
			}

			detalleEvaluacionOrganizada = append(detalleEvaluacionOrganizada, map[string]interface{}{
				"Nombre":              detalle["RequisitoProgramaAcademicoId"].(map[string]interface{})["RequisitoId"].(map[string]interface{})["Nombre"],
				"DetalleCalificacion": detalleEvalaucionJson,
			})
		}

		dataOrganizada = append(dataOrganizada, map[string]interface{}{
			"Id":                inscripcion.(map[string]interface{})["Id"],
			"TipoDocumento":     consultaPorPersona.(map[string]interface{})["Data"].(map[string]interface{})["TipoIdentificacion"].(map[string]interface{})["Nombre"],
			"Documento":         consultaPorPersona.(map[string]interface{})["Data"].(map[string]interface{})["NumeroIdentificacion"],
			"Nombre":            consultaPorPersona.(map[string]interface{})["Data"].(map[string]interface{})["NombreCompleto"],
			"TipoInscripcionId": inscripcion.(map[string]interface{})["TipoInscripcionId"].(map[string]interface{})["Nombre"],
			"DetalleEvaluacion": detalleEvaluacionOrganizada,
		})
		respuesta = append(respuesta, dataOrganizada...)
	}

	documento = ListarDataInscripcionEvaluacion(dataOrganizada, requisitosOrganizada)

	return requestresponse.APIResponseDTO(true, 200, documento)

}

func ListadoAspirantesAdmitidos(id_Periodo string, id_Estado_Fomracion string, id_Curricular string) (APIResponseDTO requestresponse.APIResponse) {

	var inscripciones []interface{}
	var aspirantes []interface{}
	var personas []map[string]interface{}
	var ICFES []interface{}
	var inscripcionesPregrado []interface{}

	// Convertir id_Estado_Formacion a float64
	id_Estado_FormacionFloat, err := strconv.ParseFloat(id_Estado_Fomracion, 64)
	if err != nil {
		log.Fatalf("Error al convertir id_Estado_Formacion a float64: %v", err)
	}

	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=ProgramaAcademicoId:"+id_Curricular+"&PeriodoId:"+id_Periodo, &inscripciones)
	if errInscripcion != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar Inscripciones: "+errInscripcion.Error())
	}

	for _, inscripcion := range inscripciones {
		if inscripcion.(map[string]interface{})["EstadoInscripcionId"].(map[string]interface{})["Id"].(float64) == id_Estado_FormacionFloat {
			aspirantes = append(aspirantes, inscripcion)
		}

	}

	for _, aspirante := range aspirantes {
		var tercero []interface{}
		var dataDocumento []interface{}
		idInscripcion := aspirante.(map[string]interface{})["Id"].(float64)
		personaId := aspirante.(map[string]interface{})["PersonaId"].(float64)
		idInscripcionString := strconv.Itoa(int(idInscripcion))
		personaIdStirng := strconv.Itoa(int(personaId))

		errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero?query=Id:"+personaIdStirng, &tercero)
		if errTercero != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar terceros: "+errTercero.Error())
		}

		errDataDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=TerceroId.Id:"+personaIdStirng, &dataDocumento)
		if errDataDocumento != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar terceros: "+errDataDocumento.Error())
		}

		errInscripcionPregrado := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion_pregrado?query=InscripcionId.Id:"+idInscripcionString, &inscripcionesPregrado)
		if errInscripcionPregrado != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar Inscripciones Pregrados: "+errInscripcionPregrado.Error())
		}

		errDataSNP := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=InscripcionId:"+idInscripcionString, &ICFES)
		if errDataSNP != nil {
			return requestresponse.APIResponseDTO(false, 500, "Error en consultar DetalleEvaluaciones: "+errDataSNP.Error())
		}

		if aspiranteMap, ok := aspirante.(map[string]interface{}); ok {
			if terceroMap, ok := tercero[0].(map[string]interface{}); ok {
				if documentoMap, ok := dataDocumento[0].(map[string]interface{}); ok {
					if inscripcionesPregradoMap, ok := inscripcionesPregrado[0].(map[string]interface{}); ok {
						for consultaICFESFor := range ICFES {
							if consultaICFESMap, ok := ICFES[consultaICFESFor].(map[string]interface{}); ok {
								if consultaICFESMap["RequisitoProgramaAcademicoId"].(map[string]interface{})["RequisitoId"].(map[string]interface{})["Nombre"] == "ICFES" {
									detalleCalificacionString := consultaICFESMap["DetalleCalificacion"].(string)
									var detalleCalificacionObj map[string]interface{}

									// Convertir el string JSON a []byte y luego deserializarlo
									err := json.Unmarshal([]byte(detalleCalificacionString), &detalleCalificacionObj)
									if err != nil {
										log.Fatalf("Error al deserializar DetalleCalificacion: %v", err)
									}

									globalValue := detalleCalificacionObj["GLOBAL"]

									persona := map[string]interface{}{
										"Credencial":  " ",
										"Nombre":      fmt.Sprintf("%s %s", terceroMap["PrimerNombre"], terceroMap["SegundoNombre"]),
										"Apellido":    fmt.Sprintf("%s %s", terceroMap["PrimerApellido"], terceroMap["SegundoApellido"]),
										"Documento":   documentoMap["Numero"],
										"SNP":         inscripcionesPregradoMap["CodigoIcfes"],
										"ICFES":       globalValue,
										"Ponderado":   aspiranteMap["NotaFinal"],
										"Inscripcion": aspiranteMap["TipoInscripcionId"].(map[string]interface{})["Nombre"],
										"Estado":      aspiranteMap["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
									}
									personas = append(personas, persona)
									fmt.Println("inscripcionesPregrado")
									fmt.Println(inscripcionesPregradoMap)
								}
							}
						}

					}

				}
			}
		}
	}

	//Esto se pasara a otra funcion

	file, err := excelize.OpenFile("static/templates/ListadoAdmitidos.xlsx")
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	style := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
	}

	styleID, err := file.NewStyle(style)
	if err != nil {
		fmt.Println(err)
	}

	indx := 0

	if id_Estado_Fomracion == "2" {
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(6), "Listado de Admitidos")

	}

	if id_Estado_Fomracion == "4" {
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(6), "Listado No Admitidos")

	}

	for i, row := range personas {
		dataRow := i + 8
		numeroRegistros := 1 + i
		indx = dataRow
		file.SetCellValue("Hoja1", "A"+strconv.Itoa(dataRow), numeroRegistros)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(dataRow), row["Credencial"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(dataRow), row["Nombre"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(dataRow), row["Apellido"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(dataRow), row["Documento"])
		file.SetCellValue("Hoja1", "F"+strconv.Itoa(dataRow), row["SNP"])
		file.SetCellValue("Hoja1", "G"+strconv.Itoa(dataRow), row["ICFES"])
		file.SetCellValue("Hoja1", "H"+strconv.Itoa(dataRow), row["Ponderado"])
		file.SetCellValue("Hoja1", "I"+strconv.Itoa(dataRow), row["Inscripcion"])
		file.SetCellValue("Hoja1", "J"+strconv.Itoa(dataRow), row["Estado"])

		if g, ok := row["general"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "T"+strconv.Itoa(dataRow), g["pbm"])
		}
		file.SetCellStyle("Hoja1", "A"+strconv.Itoa(dataRow), "J"+strconv.Itoa(dataRow), styleID)

	}

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A1:J%d", indx-1))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	if err := file.SaveAs("static/templates/ListadoAdmitidosDiligenciado.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Conversión a pdf

	//Creación plantilla base
	pdf := gofpdf.New("L", "mm", "Letter", "")
	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		WFx:      2.02,
		HFx:      2.925,
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		Footer:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          215.9,
			Ht:          1778,
		},
	}

	//Adición de header para colocar el logo de la universidad
	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/Escudo_UD.png", 26.25, 25, 25, 0, false, "", 0, "")
		}
	}
	excelPdf.ConvertSheets()
	if err != nil {
		logs.Error(err)
	}

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Pdf": encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)

}

func ListadoAspirantesOficializados(id_Periodo string, id_Nivel_Fomracion string, id_Estado_Formacion string) (APIResponseDTO requestresponse.APIResponse) {

	var proyectos map[string]interface{}
	var inscripciones []interface{}
	var tercero []interface{}
	var dataDocumento []interface{}
	var dataInfoComplementaria []interface{}
	var personas []map[string]interface{}

	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoCurricularmid")+"proyecto-academico?query=NivelFormacionId:"+id_Nivel_Fomracion, &proyectos)
	if errProyecto != nil {
		return requestresponse.APIResponseDTO(false, 500, "Error en consultar proyectos: "+errProyecto.Error())
	}

	if proyectosData, ok := proyectos["Data"].([]interface{}); ok {
		for _, proyecto := range proyectosData {
			facultad := fmt.Sprintf("%v", proyecto.(map[string]interface{})["NombreFacultad"])
			if proyectoAcademico, found := proyecto.(map[string]interface{})["ProyectoAcademico"].(map[string]interface{}); found {
				if id, idFound := proyectoAcademico["Id"]; idFound {

					fmt.Println("http://" + beego.AppConfig.String("InscripcionService") + "inscripcion?query=ProgramaAcademicoId:" + strconv.Itoa(int(id.(float64))) + ",PeriodoId:" + id_Periodo + ",EstadoInscripcionId.Id:" + id_Estado_Formacion)

					errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=ProgramaAcademicoId:"+strconv.Itoa(int(id.(float64)))+",PeriodoId:"+id_Periodo+",EstadoInscripcionId.Id:"+id_Estado_Formacion, &inscripciones)
					if errInscripcion != nil {
						break
						//return requestresponse.APIResponseDTO(false, 500, "Error en consultar Inscripciones: "+errInscripcion.Error())
					}

					for _, inscripcion := range inscripciones {
						idInscripcion := fmt.Sprintf("%v", inscripcion.(map[string]interface{})["Id"])
						idPersona := fmt.Sprintf("%v", inscripcion.(map[string]interface{})["PersonaId"])

						errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero?query=Id:"+idPersona, &tercero)
						if errTercero != nil {
							break
							//return requestresponse.APIResponseDTO(false, 500, "Error en consultar terceros: "+errTercero.Error())
						}

						if terceroMap, ok := tercero[0].(map[string]interface{}); ok {
							idTercero := fmt.Sprintf("%v", terceroMap["Id"])

							errTerceroDocument := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=TerceroId.Id:"+idTercero, &dataDocumento)
							if errTerceroDocument != nil {
								return requestresponse.APIResponseDTO(false, 500, "Error en consultar Documentos de Aspirantes: "+errTerceroDocument.Error())
							}

							if documentoMap, ok := dataDocumento[0].(map[string]interface{}); ok {

								errTerceroInfoComplementaria := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"info_complementaria_tercero?query=TerceroId.Id:"+idTercero, &dataInfoComplementaria)
								if errTerceroInfoComplementaria != nil {
									return requestresponse.APIResponseDTO(false, 500, "Error en consultar Documentos de Aspirantes: "+errTerceroInfoComplementaria.Error())
								}

								var telefono string
								var correo string

								for _, infoComplementaria := range dataInfoComplementaria {

									if infoComplementariaMap, ok := infoComplementaria.(map[string]interface{}); ok {
										if infoComplementariaID, ok := infoComplementariaMap["InfoComplementariaId"].(map[string]interface{}); ok {
											if infoComplementariaID["Nombre"] == "CORREO" {

												// Obtener el valor de "Dato" y deserializar el JSON
												dato := fmt.Sprintf("%v", infoComplementariaMap["Dato"])
												var datoMap map[string]interface{}
												if err := json.Unmarshal([]byte(dato), &datoMap); err != nil {
													fmt.Println("Error deserializando JSON:", err)
												} else {
													// Obtener el valor de "value"
													if value, ok := datoMap["value"]; ok {
														correo = fmt.Sprintf("%v", value)
													}
												}
											}

											if infoComplementariaID["Nombre"] == "TELEFONO" {

												// Obtener el valor de "Dato" y deserializar el JSON
												dato := fmt.Sprintf("%v", infoComplementariaMap["Dato"])
												var datoMap map[string]interface{}
												if err := json.Unmarshal([]byte(dato), &datoMap); err != nil {
													fmt.Println("Error deserializando JSON:", err)
												} else {
													// Obtener el valor de "principal"
													if principal, ok := datoMap["principal"]; ok {
														telefono = fmt.Sprintf("%v", principal)
													}
												}
											}
										}
									}
								}

								persona := map[string]interface{}{
									"Facultad":       facultad,
									"Codigo":         idInscripcion,
									"Documento":      documentoMap["Numero"],
									"Nombre":         fmt.Sprintf("%s %s", terceroMap["PrimerNombre"], terceroMap["SegundoNombre"]),
									"Apellido":       fmt.Sprintf("%s %s", terceroMap["PrimerApellido"], terceroMap["SegundoApellido"]),
									"Correopersonal": correo,
									"Telefono":       telefono,
									"Correosugerido": fmt.Sprintf("%s %s %s %s",
										safeFirstChar(terceroMap["PrimerNombre"].(string)),
										safeFirstChar(terceroMap["SegundoNombre"].(string)),
										terceroMap["PrimerApellido"],
										safeFirstChar(terceroMap["SegundoApellido"].(string))+"@udistrital.edu.co"),
									"Correoasignado": terceroMap["UsuarioWSO2"],
								}

								personas = append(personas, persona)
							}
						}
					}
				}
			}
		}
	}

	//Esto se pasara a otra funcion

	file, err := excelize.OpenFile("static/templates/ListadoOficializadosOficializados.xlsx")
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	style := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
	}

	styleID, err := file.NewStyle(style)
	if err != nil {
		fmt.Println(err)
	}

	indx := 0

	if id_Estado_Formacion == "11" {
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(6), "Listado de oficializados")

	}

	if id_Estado_Formacion == "12" {
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(6), "Listado No oficializados")

	}

	for i, row := range personas {
		dataRow := i + 8
		numeroRegistros := 1 + i
		indx = dataRow
		file.SetCellValue("Hoja1", "A"+strconv.Itoa(dataRow), numeroRegistros)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(dataRow), row["Facultad"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(dataRow), row["Codigo"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(dataRow), row["Documento"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(dataRow), row["Nombre"])
		file.SetCellValue("Hoja1", "F"+strconv.Itoa(dataRow), row["Apellido"])
		file.SetCellValue("Hoja1", "G"+strconv.Itoa(dataRow), row["Correopersonal"])
		file.SetCellValue("Hoja1", "H"+strconv.Itoa(dataRow), row["Telefono"])
		file.SetCellValue("Hoja1", "I"+strconv.Itoa(dataRow), row["Correosugerido"])
		file.SetCellValue("Hoja1", "J"+strconv.Itoa(dataRow), row["Correoasignado"])

		if g, ok := row["general"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "T"+strconv.Itoa(dataRow), g["pbm"])
		}
		file.SetCellStyle("Hoja1", "A"+strconv.Itoa(dataRow), "J"+strconv.Itoa(dataRow), styleID)

	}

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A1:J%d", indx-1))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	if err := file.SaveAs("static/templates/ListadoOficializadosOficializadosDiligenciado.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Conversión a pdf

	//Creación plantilla base
	pdf := gofpdf.New("L", "mm", "Letter", "")
	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		WFx:      2.02,
		HFx:      2.925,
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		Footer:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          215.9,
			Ht:          1778,
		},
	}

	//Adición de header para colocar el logo de la universidad
	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/Escudo_UD.png", 26.25, 25, 25, 0, false, "", 0, "")
		}
	}
	excelPdf.ConvertSheets()
	if err != nil {
		logs.Error(err)
	}

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Pdf": encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)
}

func InformeLiquidacionPosgrado(data []byte) (APIResponseDTO requestresponse.APIResponse) {

	var admitidos []map[string]interface{}

	if err := json.Unmarshal(data, &admitidos); err == nil {
		fmt.Println(admitidos)
	} else {
		return helpers.ErrEmiter(err)

	}

	indx := 0

	file, err := excelize.OpenFile("static/templates/TemplateInformePosgrado.xlsx")
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	style := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
	}

	styleID, err := file.NewStyle(style)
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	for i, row := range admitidos {
		//Esta data no es correspondiente aL INFORME DE POSGRADO, TOCA HACER CORRECION
		fmt.Println("Row", row)
		fmt.Println("Row", row["PrimerApellido"])
		dataRow := i + 9
		numeroRegistros := 1 + i
		indx = dataRow
		file.SetCellValue("Hoja1", "A"+strconv.Itoa(dataRow), numeroRegistros)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(dataRow), row["codigo"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(dataRow), row["documento"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(dataRow), row["nombres"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(dataRow), row["apellidos"])
		if a, ok := row["A"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "F"+strconv.Itoa(dataRow), a["A1"])
			file.SetCellValue("Hoja1", "G"+strconv.Itoa(dataRow), a["puntajeA1"])
			file.SetCellValue("Hoja1", "H"+strconv.Itoa(dataRow), a["A2"])
			file.SetCellValue("Hoja1", "I"+strconv.Itoa(dataRow), a["puntajeA2"])
			file.SetCellValue("Hoja1", "J"+strconv.Itoa(dataRow), a["A3"])
			file.SetCellValue("Hoja1", "K"+strconv.Itoa(dataRow), a["puntajeA3"])
		}
		if b, ok := row["B"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "L"+strconv.Itoa(dataRow), b["B1"])
			file.SetCellValue("Hoja1", "M"+strconv.Itoa(dataRow), b["puntajeB1"])
			file.SetCellValue("Hoja1", "M"+strconv.Itoa(dataRow), b["B2"])
			file.SetCellValue("Hoja1", "O"+strconv.Itoa(dataRow), b["puntajeB2"])
			file.SetCellValue("Hoja1", "P"+strconv.Itoa(dataRow), b["B3"])
			file.SetCellValue("Hoja1", "Q"+strconv.Itoa(dataRow), b["puntajeB3"])
			file.SetCellValue("Hoja1", "R"+strconv.Itoa(dataRow), b["B4"])
			file.SetCellValue("Hoja1", "S"+strconv.Itoa(dataRow), b["puntajeB4"])
		}

		if g, ok := row["general"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "T"+strconv.Itoa(dataRow), g["pbm"])
		}
		file.SetCellStyle("Hoja1", "A"+strconv.Itoa(dataRow), "AB"+strconv.Itoa(dataRow), styleID)

	}

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A1:X%d", indx-1))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	if err := file.SaveAs("static/templates/TemplateInformePosgradoModificado.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Conversión a pdf

	//Creación plantilla base
	pdf := gofpdf.New("L", "mm", "", "")
	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		WFx:      2.02,
		HFx:      2.925,
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		Footer:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          215.9,
			Ht:          1778,
		},
	}

	//Adición de header para colocar el logo de la universidad
	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/Escudo_UD.png", 26.25, 25, 25, 0, false, "", 0, "")
		}
	}
	excelPdf.ConvertSheets()
	if err != nil {
		logs.Error(err)
	}

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Pdf": encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)
}

func InformeLiquidacionPregrado(data []byte) (APIResponseDTO requestresponse.APIResponse) {

	var admitidos []map[string]interface{}

	if err := json.Unmarshal(data, &admitidos); err == nil {
		fmt.Println(admitidos)
	} else {
		return helpers.ErrEmiter(err)

	}

	indx := 0

	file, err := excelize.OpenFile("static/templates/TemplateInformePregrado.xlsx")
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	style := &excelize.Style{
		Border: []excelize.Border{
			{
				Type:  "left",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "top",
				Color: "#000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "#000000",
				Style: 1,
			},
		},
	}

	styleID, err := file.NewStyle(style)
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	for i, row := range admitidos {
		fmt.Println("Row", row)
		fmt.Println("Row", row["PrimerApellido"])
		dataRow := i + 12
		numeroRegistros := 1 + i
		indx = dataRow
		file.SetCellValue("Hoja1", "A"+strconv.Itoa(dataRow), numeroRegistros)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(dataRow), row["codigo"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(dataRow), row["documento"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(dataRow), row["nombres"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(dataRow), row["apellidos"])
		if a, ok := row["A"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "F"+strconv.Itoa(dataRow), a["A1"])
			file.SetCellValue("Hoja1", "G"+strconv.Itoa(dataRow), a["puntajeA1"])
			file.SetCellValue("Hoja1", "H"+strconv.Itoa(dataRow), a["A2"])
			file.SetCellValue("Hoja1", "I"+strconv.Itoa(dataRow), a["puntajeA2"])
			file.SetCellValue("Hoja1", "J"+strconv.Itoa(dataRow), a["A3"])
			file.SetCellValue("Hoja1", "K"+strconv.Itoa(dataRow), a["puntajeA3"])
		}
		if b, ok := row["B"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "L"+strconv.Itoa(dataRow), b["B1"])
			file.SetCellValue("Hoja1", "M"+strconv.Itoa(dataRow), b["puntajeB1"])
			file.SetCellValue("Hoja1", "M"+strconv.Itoa(dataRow), b["B2"])
			file.SetCellValue("Hoja1", "O"+strconv.Itoa(dataRow), b["puntajeB2"])
			file.SetCellValue("Hoja1", "P"+strconv.Itoa(dataRow), b["B3"])
			file.SetCellValue("Hoja1", "Q"+strconv.Itoa(dataRow), b["puntajeB3"])
			file.SetCellValue("Hoja1", "R"+strconv.Itoa(dataRow), b["B4"])
			file.SetCellValue("Hoja1", "S"+strconv.Itoa(dataRow), b["puntajeB4"])
		}

		if g, ok := row["general"].(map[string]interface{}); ok {
			file.SetCellValue("Hoja1", "T"+strconv.Itoa(dataRow), g["pbm"])
		}
		file.SetCellStyle("Hoja1", "A"+strconv.Itoa(dataRow), "x"+strconv.Itoa(dataRow), styleID)

	}

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A1:X%d", indx-1))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	if err := file.SaveAs("static/templates/TemplateInformePregradoModificado.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Conversión a pdf

	//Creación plantilla base
	pdf := gofpdf.New("L", "mm", "Letter", "")
	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          200,
			Ht:          300,
		},
	}

	//Adición de header para colocar el logo de la universidad
	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/Escudo_UD.png", 26.25, 25, 25, 0, false, "", 0, "")
		}
	}
	excelPdf.ConvertSheets()
	if err != nil {
		logs.Error(err)
	}

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Pdf": encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)
}

func errEmiter(errData error, infoData ...string) requestresponse.APIResponse {
	if errData != nil {
		return requestresponse.APIResponseDTO(false, 400, nil, errData.Error())
	}

	if len(infoData) > 0 && (infoData[0] == "[map[]]" || infoData[0] == "map[]") {
		return requestresponse.APIResponseDTO(false, 404, nil, "No se encontraron datos")
	}

	return requestresponse.APIResponseDTO(false, 400, "nil")
}

func GenerarReporteCodigos(idPeriodo int64, idProyecto int64) requestresponse.APIResponse {
	//Mapa para guardar los admitidos
	var admitidos []map[string]interface{}

	//Obtener Datos del periodo
	periodo, errPeriodo := obtenerInfoPeriodo(fmt.Sprintf("%v", idPeriodo))
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		return errEmiter(errPeriodo, fmt.Sprintf("%v", periodo))
	}

	//Obtener Datos del proyecto & facultad
	proyecto, facultad, err := obtenerInfoProyectoyFacultad(fmt.Sprintf("%v", idProyecto))
	if err != nil || fmt.Sprintf("%v", proyecto) == "map[]" || fmt.Sprintf("%v", facultad) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", proyecto), fmt.Sprintf("%v", facultad))
	}

	//Inscripciones de admitidos
	var inscripciones []map[string]interface{}
	errInscripciones := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:ADMITIDO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", idProyecto, idPeriodo), &inscripciones)
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "[map[]]" {
		return errEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	}

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
			return helpers.ErrEmiter(errTercero, fmt.Sprintf("%v", tercero))
		}

		//Obtener Documento Tercero
		var terceroDocumento []map[string]interface{}
		errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:%v", inscripcion["PersonaId"]), &terceroDocumento)
		if errTerceroDocumento != nil || fmt.Sprintf("%v", terceroDocumento) == "[map[]]" {
			return errEmiter(errTerceroDocumento, fmt.Sprintf("%v", terceroDocumento))
		}

		//Obtener Codigo Tercero
		var terceroCodigo []map[string]interface{}
		errTerceroCodigo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CODE,Activo:true,TerceroId:%v,Numero__contains:%v", inscripcion["PersonaId"], codigoBase), &terceroCodigo)
		if errTerceroCodigo != nil || fmt.Sprintf("%v", terceroCodigo) == "[map[]]" {
			return errEmiter(errTerceroCodigo, fmt.Sprintf("%v", terceroCodigo))
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

	//Añadir información de la cabecera de el excel
	infoCabecera := map[string]interface{}{
		"Facultad": facultad["Nombre"],
		"Proyecto": proyecto["Nombre"],
		"Periodo":  periodo["Data"].(map[string]interface{})["Nombre"],
	}

	//Función que genera el reporte en xlsx
	return generarExcelReporteCodigos(admitidos, infoCabecera)

}

func generarExcelReporteCodigos(admitidosMap []map[string]interface{}, infoCabecera map[string]interface{}) requestresponse.APIResponse {

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

	//Abrir Plantilla Excel
	file, err := excelize.OpenFile("static/templates/TemplateReporteCodificacion.xlsx")
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	var lastCell = ""

	for i, row := range admitidos {
		dataRow := i + 8
		for j, col := range row {
			file.SetCellValue("Hoja1", fmt.Sprintf("%s%d", string(rune(65+j)), dataRow), col)
			lastCell = fmt.Sprintf("%s%d", string(rune(65+j)), dataRow)
		}
	}

	file.SetCellValue("Hoja1", "B5", fmt.Sprintf("Facultad: %v", infoCabecera["Facultad"]))
	file.SetCellValue("Hoja1", "D5", fmt.Sprintf("Proyecto: %v", infoCabecera["Proyecto"]))
	file.SetCellValue("Hoja1", "F5", fmt.Sprintf("Periodo: %v", infoCabecera["Periodo"]))

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A2:%v", lastCell))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	//creación de el estilo para el excel
	style2, err := file.NewStyle(
		&excelize.Style{
			Alignment: &excelize.Alignment{Horizontal: "center"},
			Border: []excelize.Border{
				{Type: "left", Color: "00000000", Style: 1},
				{Type: "right", Color: "00000000", Style: 1},
				{Type: "top", Color: "00000000", Style: 1},
				{Type: "bottom", Color: "00000000", Style: 1},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
		return helpers.ErrEmiter(err)
	}

	file.SetCellStyle("Hoja1", "A7", lastCell, style2)

	//Conversión a pdf

	//Creación plantilla base
	pdf := gofpdf.New("", "", "", "")

	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          200,
			Ht:          300,
		},
	}

	//Adición de header para colocar el logo de la universidad
	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/Escudo_UD.png", 26.25, 25, 25, 0, false, "", 0, "")
		}
	}

	excelPdf.ConvertSheets()

	//Adición de colores al excel luego de generar el pdf

	//creación de el estilo para el excel
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
		return helpers.ErrEmiter(err)
	}

	file.SetCellStyle("Hoja1", "A7", lastCell, style)

	//Guardado en local excel
	/*if err := file.SaveAs("static/templates/Modificado.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}*/

	//Guardado en local PDF ----> Si se guarda en local el PDF se borra de el buffer y no se genera el base 64
	/*err = pdf.OutputFileAndClose("static/templates/Reporte.pdf")
	if err != nil {
		return errEmiter(err)
	}*/

	//Conversión a base 64

	//Excel
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return helpers.ErrEmiter(err)
	}

	encodedFileExcel := base64.StdEncoding.EncodeToString(buffer.Bytes())

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Excel": encodedFileExcel,
		"Pdf":   encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)

}

/*
	Funciones de obtención de información recurrente
*/

func obtenerInfoPeriodo(idPeriodo string) (map[string]interface{}, error) {
	//Obtener Datos del periodo
	var periodo map[string]interface{}
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+fmt.Sprintf("periodo/%v", idPeriodo), &periodo)
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		return periodo, errPeriodo
	}

	return periodo, nil
}

func obtenerInfoProyectoyFacultad(idProyecto string) (map[string]interface{}, map[string]interface{}, error) {
	//Obtener Datos del proyecto & facultad
	var facultad map[string]interface{}

	var proyecto map[string]interface{}
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_institucion/%v", idProyecto), &proyecto)
	if errProyecto != nil || fmt.Sprintf("%v", proyecto) == "map[]" {
		return proyecto, facultad, errProyecto
	} else {
		//Obtener Datos de la facultad
		errFacultad := request.GetJson("http://"+beego.AppConfig.String("OikosService")+fmt.Sprintf("dependencia/%v", proyecto["FacultadId"]), &facultad)
		if errFacultad != nil || fmt.Sprintf("%v", facultad) == "map[]" {
			return proyecto, facultad, errFacultad
		}
	}

	return proyecto, facultad, nil
}

func obtenerInfoTercero(idTercero string) ([]map[string]interface{}, error) {
	//Obtener Datos basicos Tercero
	var tercero []map[string]interface{}
	errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero?query=Id:"+idTercero, &tercero)
	if errTercero != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
		return tercero, errTercero
	} else {
		return tercero, nil
	}
}

func obtenerDocumentoTercero(idTercero string) ([]map[string]interface{}, error) {
	//Obtener Documento Tercero
	var terceroDocumento []map[string]interface{}
	errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=Activo:true,TerceroId:"+idTercero, &terceroDocumento)
	if errTerceroDocumento != nil || fmt.Sprintf("%v", terceroDocumento) == "[map[]]" {
		return terceroDocumento, errTerceroDocumento
	} else {
		return terceroDocumento, nil
	}
}

func obtenerCorreoTercero(idTercero string) (correo string) {
	//Obtener Correo Tercero
	var terceroCorreo []map[string]interface{}
	errTerceroCorreo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId__Nombre:CORREO,Activo:true", idTercero), &terceroCorreo)
	if errTerceroCorreo != nil || fmt.Sprintf("%v", terceroCorreo) == "[map[]]" {
		correo = "NA"
	} else {
		var correoPrincipal map[string]interface{}
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", terceroCorreo[0]["Dato"])), &correoPrincipal); err == nil {
			correo = fmt.Sprintf("%v", correoPrincipal["Data"])
		} else {
			fmt.Print(err.Error())
		}
	}
	return correo
}

func obtenerTelefonoTercero(idTercero string) (telefono string) {
	//Obtener Telefono Tercero
	var terceroTelefono []map[string]interface{}
	errTerceroTelefono := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId__Nombre:TELEFONO,Activo:true", idTercero), &terceroTelefono)
	if errTerceroTelefono != nil || fmt.Sprintf("%v", terceroTelefono) == "[map[]]" {
		telefono = "NA"
	} else {
		var telefonoPrincipal map[string]interface{}
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", terceroTelefono[0]["Dato"])), &telefonoPrincipal); err == nil {
			telefono = strconv.FormatFloat(telefonoPrincipal["principal"].(float64), 'f', -1, 64)
		} else {
			fmt.Print(err.Error())
		}
	}

	return telefono
}

func obtenerEnfasis(idTercero string) (nombreEnfasis string) {
	//Obtener Enfasis
	var enfasis []map[string]interface{}
	errEnfasis := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_enfasis?query=Id:%v", idTercero), &enfasis)
	if errEnfasis != nil || fmt.Sprintf("%v", enfasis) == "[map[]]" {
		nombreEnfasis = "NA"
	} else {
		nombreEnfasis = fmt.Sprintf("%v", enfasis[0]["EnfasisId"].(map[string]interface{})["Nombre"])
	}

	return nombreEnfasis
}

/*
	Reportes dinamicos
	1 -> Inscritos por  programa
	2 -> Admitidos por  programa
	3 -> Aspirantes por programa
	4 -> Apirantes de todos los programas
	5 -> Transferencias internas
	6 -> Transferencias externas
	7 -> Reintegros
*/

func ReporteDinamico(data []byte) requestresponse.APIResponse {
	var reporte models.ReporteEstructura
	var respuesta requestresponse.APIResponse
	if err := json.Unmarshal(data, &reporte); err == nil {
		if reporte.TipoReporte != 0 {
			if reporte.TipoReporte < 4 {
				respuesta = reporteInscritosPorPrograma(reporte)
			} else if reporte.TipoReporte == 4 {
				respuesta = reporteAspirantesPeriodoYnivel(reporte)
			} else {
				respuesta = reporteTransferenciasReintegros(reporte)
			}
		}

	} else {
		respuesta = requestresponse.APIResponseDTO(false, 400, nil)
	}

	return respuesta
}

func columnasParaEliminar(columnasSolicitadas []string) []string {

	//Maximo de columnas con el tamaño de header definido
	columnasMaximas := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	var columnasParaEliminar []string
	var contains bool

	for _, maxima := range columnasMaximas {
		contains = false
		for _, solicitada := range columnasSolicitadas {
			if maxima == solicitada {
				contains = true
			}
		}
		if !contains {
			columnasParaEliminar = append(columnasParaEliminar, maxima)
		}
	}

	return columnasParaEliminar
}

// Funcion para reporte de Inscritos por programa
func reporteInscritosPorPrograma(infoReporte models.ReporteEstructura) requestresponse.APIResponse {

	var inscritos [][]interface{}

	//Definir Columnas a eliminar
	infoReporte.Columnas = columnasParaEliminar(infoReporte.Columnas)

	//Obtener proyecto y facultad
	proyecto, facultad, err := obtenerInfoProyectoyFacultad(fmt.Sprintf("%v", infoReporte.Proyecto))
	if err != nil || fmt.Sprintf("%v", proyecto) == "map[]" || fmt.Sprintf("%v", facultad) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", proyecto), fmt.Sprintf("%v", facultad))
	}

	periodo, err := obtenerInfoPeriodo(fmt.Sprintf("%v", infoReporte.Periodo))
	if err != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", periodo))
	}

	//Primer o segundo semestre segun el ciclo
	if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "1" {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "PRIMER"
	} else {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "SEGUNDO"
	}

	dataHeader := map[string]interface{}{
		"Año":                periodo["Data"].(map[string]interface{})["Year"],
		"Semestre":           periodo["Data"].(map[string]interface{})["Ciclo"],
		"ProyectoCurricular": strings.ToUpper(fmt.Sprintf("%v", proyecto["Nombre"])),
		"Indices": []interface{}{
			"#",
			"Documento",
			"Nombre Completo",
			"Telefono",
			"Correo",
		},
	}

	var inscripciones []map[string]interface{}
	var errInscripciones error

	if infoReporte.TipoReporte == 1 {

		//Añadir headers no compartidos
		dataHeader["Indices"] = append(dataHeader["Indices"].([]interface{}),
			"Credencial",
			"Enfasis",
			"Descuento",
			"Estado inscripción")

		//Hacer consulta especifica para estado inscrito
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:INSCRITO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,TipoInscripcionId__Id:%v&limit=0", infoReporte.Proyecto, infoReporte.Periodo, infoReporte.TipoInscripcion), &inscripciones)

	} else if infoReporte.TipoReporte == 2 {

		//Añadir headers no compartidos
		dataHeader["Indices"] = append(dataHeader["Indices"].([]interface{}),
			"Credencial",
			"Enfasis",
			"Estado inscripción",
			"Puntaje")

		//Hacer consulta especifica para estado ADMITIDO U OPCIONADO
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre__in:ADMITIDO|OPCIONADO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,TipoInscripcionId__Id:%v&limit=0", infoReporte.Proyecto, infoReporte.Periodo, infoReporte.TipoInscripcion), &inscripciones)
	} else {
		//Añadir headers no compartidos
		dataHeader["Indices"] = append(dataHeader["Indices"].([]interface{}),
			"Tipo inscripción",
			"Estado inscripción")

		//Hacer consulta especifica para aspirantes
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,TipoInscripcionId__Id:%v&limit=0", infoReporte.Proyecto, infoReporte.Periodo, infoReporte.TipoInscripcion), &inscripciones)
	}

	//Si existen inscripciones entonces
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "[map[]]" {
		return errEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	} else {

		for _, inscripcion := range inscripciones {
			//Datos basicos tercero
			tercero, err := obtenerInfoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
				return errEmiter(err)
			}

			//Obtener Documento Tercero
			terceroDocumento, err := obtenerDocumentoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", terceroDocumento) == "[map[]]" {
				return errEmiter(err)
			}
			//Obtener Telefono Tercero

			terceroTelefono := obtenerTelefonoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))

			//Obtener Correo Tercero
			terceroCorreo := obtenerCorreoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))

			//Obtener enfasis
			enfasis := obtenerEnfasis(fmt.Sprintf("%v", inscripcion["EnfasisId"]))

			//Obtener descuentos solicitados
			var nombreDescuento string
			var descuento []map[string]interface{}
			errDescuento := request.GetJson("http://"+beego.AppConfig.String("DescuentosService")+fmt.Sprintf("solicitud_descuento?query=TerceroId:%v,PeriodoId:%v,DescuentosDependenciaId__DependenciaId:%v", inscripcion["PersonaId"], infoReporte.Periodo, infoReporte.Proyecto), &descuento)
			if errDescuento != nil || fmt.Sprintf("%v", descuento) == "[map[]]" {
				nombreDescuento = "NA"
			} else {
				nombreDescuento = fmt.Sprintf("%v",
					descuento[0]["DescuentosDependenciaId"].(map[string]interface{})["TipoDescuentoId"].(map[string]interface{})["Nombre"])
			}

			inscrito := []interface{}{
				terceroDocumento[0]["Numero"],
				tercero[0]["NombreCompleto"],
				terceroTelefono,
				terceroCorreo,
			}

			if infoReporte.TipoReporte == 1 {
				inscrito = append(inscrito, inscripcion["Id"], enfasis, nombreDescuento, inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"])
			} else if infoReporte.TipoReporte == 2 {
				inscrito = append(inscrito, inscripcion["Id"], enfasis, inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"], inscripcion["NotaFinal"])
			} else {
				inscrito = append(inscrito, inscripcion["TipoInscripcionId"].(map[string]interface{})["Nombre"], inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"])
			}
			inscritos = append(inscritos, inscrito)

		}

		return generarXlsxyPdfIncripciones(infoReporte, inscritos, dataHeader)

	}

}

func crearInfoComplementaria(aspirantes []map[string]interface{}) string {
	var inscripcionSolicitada = 0
	var admitido = 0
	var opcionado = 0
	var noAdmitido = 0
	var inscrito = 0
	var inscritoObservacion = 0

	totalInscritos := len(aspirantes)
	for _, aspirante := range aspirantes {
		switch aspirante["EstadoInscripcionId"].(map[string]interface{})["Id"].(float64) {
		case 1:
			inscripcionSolicitada++
		case 2:
			admitido++
		case 3:
			opcionado++
		case 4:
			noAdmitido++
		case 5:
			inscrito++
		case 6:
			inscritoObservacion++
		}
	}

	return fmt.Sprintf("Inscripción solicitada: %v       Admitido: %v      Opcionado: %v      No admitido: %v      Inscrito: %v      Inscrito con observación: %v      Total aspirantes: %v", inscripcionSolicitada, admitido, opcionado, noAdmitido, inscrito, inscritoObservacion, totalInscritos)
}

func reporteAspirantesPeriodoYnivel(infoReporte models.ReporteEstructura) requestresponse.APIResponse {

	var aspirantes [][][]interface{}

	var proyectos []map[string]interface{}
	var facultades []map[string]interface{}
	var dataHeader []map[string]interface{}

	//Definir Columnas a eliminar
	infoReporte.Columnas = columnasParaEliminar(infoReporte.Columnas)

	periodo, err := obtenerInfoPeriodo(fmt.Sprintf("%v", infoReporte.Periodo))
	if err != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", periodo))
	}

	//Primer o segundo semestre segun el ciclo
	if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "1" {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "PRIMER"
	} else {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "SEGUNDO"
	}

	//Obtener información de todos los proyectos
	respuesta, err := GetAspirantesDeProyectosActivos(fmt.Sprintf("%v", infoReporte.Proyecto), fmt.Sprintf("%v", infoReporte.Periodo), "3")

	for _, proyecto := range respuesta.([]map[string]interface{}) {

		//Obtener proyecto y facultad
		proyectoAspirantes, facultadAspirantes, err := obtenerInfoProyectoyFacultad(fmt.Sprintf("%v", proyecto["ProyectoId"]))
		if err != nil || fmt.Sprintf("%v", proyectoAspirantes) == "map[]" || fmt.Sprintf("%v", facultadAspirantes) == "map[]" {
			return errEmiter(err, fmt.Sprintf("%v", proyectoAspirantes), fmt.Sprintf("%v", facultadAspirantes))
		}
		proyectos = append(proyectos, proyectoAspirantes)
		facultades = append(facultades, facultadAspirantes)

		aspiranteArray := [][]interface{}{}
		for _, aspirante := range proyecto["Aspirantes"].([]map[string]interface{}) {

			//Definir data aspirante
			aspiranteArray = append(aspiranteArray, []interface{}{
				aspirante["NumeroDocumento"],
				aspirante["NombreAspirante"],
				aspirante["Telefono"],
				aspirante["Email"],
				aspirante["NotaFinal"],
				aspirante["TipoInscripcion"],
				aspirante["Enfasis"],
				aspirante["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				aspirante["EstadoRecibo"]})
		}
		aspirantes = append(aspirantes, aspiranteArray)
	}

	for i, proyecto := range proyectos {
		dataHeader = append(dataHeader, map[string]interface{}{
			"ProyectoCurricular":        strings.ToUpper(fmt.Sprintf("%v", proyecto["Nombre"])),
			"InformacionComplementaria": crearInfoComplementaria(respuesta.([]map[string]interface{})[i]["Aspirantes"].([]map[string]interface{})),
			"Indices": []interface{}{
				"#",
				"Documento",
				"Nombre Completo",
				"Telefono",
				"Correo",
				"Puntaje",
				"Tipo de inscripción",
				"Enfasis",
				"Estado inscripción",
				"Estado recibo",
			},
		})
	}

	//Abrir Plantilla Excel
	file, err := excelize.OpenFile("static/templates/ReporteInscritosOriginal.xlsx")
	if err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Agregar data al reporte

	//Fila de inicio de la plantilla
	var lastRow = 5

	file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE ASPIRANTES  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", periodo["Data"].(map[string]interface{})["Ciclo"], periodo["Data"].(map[string]interface{})["Year"]))

	var dataRow = 0
	var lastCell = ""
	for i, aspirante := range aspirantes {

		//Colocar indices al reporte
		//Proyecto
		file.MergeCell("Hoja1", fmt.Sprintf("A%v", lastRow+1), fmt.Sprintf("J%v", lastRow+1))
		file.SetCellValue("Hoja1", fmt.Sprintf("A%v", lastRow+1), dataHeader[i]["ProyectoCurricular"])
		file.SetRowHeight("Hoja1", lastRow+1, 35)
		//Información complementaria
		file.MergeCell("Hoja1", fmt.Sprintf("A%v", lastRow+2), fmt.Sprintf("J%v", lastRow+2))
		file.SetCellValue("Hoja1", fmt.Sprintf("A%v", lastRow+2), dataHeader[i]["InformacionComplementaria"])
		//Indices de columna
		for k, header := range dataHeader[i]["Indices"].([]interface{}) {
			file.SetCellValue("Hoja1", fmt.Sprintf("%v%v", string(rune(65+k)), lastRow+3), header)
		}

		lastRow = lastRow + 4
		for j, row := range aspirante {
			dataRow = (j + lastRow)
			file.SetCellValue("Hoja1", fmt.Sprintf("A%v", dataRow), j+1)
			for h, col := range row {
				file.SetCellValue("Hoja1", fmt.Sprintf("%s%d", string(rune(65+h+1)), dataRow), col)
				lastCell = fmt.Sprintf("%s%d", string(rune(65+h+1)), dataRow)
			}
			file.SetRowHeight("Hoja1", dataRow, 30)
		}
		lastRow = dataRow
	}

	//creación de el estilo para el excel
	style, err := file.NewStyle(
		&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
				WrapText:   true,
			},
			Border: []excelize.Border{
				{Type: "left", Color: "00000000", Style: 1},
				{Type: "right", Color: "00000000", Style: 1},
				{Type: "top", Color: "00000000", Style: 1},
				{Type: "bottom", Color: "00000000", Style: 1},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	file.SetCellStyle("Hoja1", "A8", lastCell, style)

	//Redimensión de el excel para que el convertidor tome todas las celdas
	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A2:%v", lastCell))
	if errDimesion != nil {
		return errEmiter(errDimesion)
	}

	//Funcion reverse columans
	for i, j := 0, len(infoReporte.Columnas)-1; i < j; i, j = i+1, j-1 {
		infoReporte.Columnas[i], infoReporte.Columnas[j] = infoReporte.Columnas[j], infoReporte.Columnas[i]
	}

	//Eliminador de columnas
	for _, columna := range infoReporte.Columnas {
		file.RemoveCol("Hoja1", columna)
	}

	//Definir ancho dinamico de las columnas
	//167.5 es el ancho total del reporte
	var anchoPorColumna = float64(167.5) / float64(10-len(infoReporte.Columnas))
	file.SetColWidth("Hoja1", "A", string(rune(65+(10-len(infoReporte.Columnas)))), anchoPorColumna)

	//Insertar header Xlsx
	if err := file.AddPicture("Hoja1", "A2", "static/images/HeaderEstaticoRecortado.jpg",
		&excelize.GraphicOptions{
			ScaleX:  0.20, //Escalado en x de la imagen
			ScaleY:  0.15, //Escalado en y de la imagen
			OffsetX: 2,    //Espacio entre la celda y la imagen para x
			OffsetY: 2,    //Espacio entre la celda y la imagen para y
		},
	); err != nil {
		errEmiter(err)
	}

	//Creación plantilla base
	pdf := gofpdf.New("", "", "", "")

	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          600,
			Ht:          370,
		},
	}

	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/HeaderEstaticoRecortado.jpg", 25, 25, 320, 25, false, "", 0, "")
		}
	}

	excelPdf.ConvertSheets()

	/*if err := file.SaveAs("static/templates/ModificadoInscritos.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	err = pdf.OutputFileAndClose("static/templates/ReporteInscrito.pdf") //----> Si se guarda en local el PDF se borra de el buffer y no se genera el base 64
	if err != nil {
		return errEmiter(err)
	}*/

	//Excel
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return errEmiter(err)
	}

	encodedFileExcel := base64.StdEncoding.EncodeToString(buffer.Bytes())

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuestaFront := map[string]interface{}{
		"Excel": encodedFileExcel,
		"Pdf":   encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuestaFront)

}

func reporteTransferenciasReintegros(infoReporte models.ReporteEstructura) requestresponse.APIResponse {
	var inscritos [][]interface{}

	//Definir Columnas a eliminar
	infoReporte.Columnas = columnasParaEliminar(infoReporte.Columnas)

	//Obtener proyecto y facultad
	proyecto, facultad, err := obtenerInfoProyectoyFacultad(fmt.Sprintf("%v", infoReporte.Proyecto))
	if err != nil || fmt.Sprintf("%v", proyecto) == "map[]" || fmt.Sprintf("%v", facultad) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", proyecto), fmt.Sprintf("%v", facultad))
	}

	periodo, err := obtenerInfoPeriodo(fmt.Sprintf("%v", infoReporte.Periodo))
	if err != nil || fmt.Sprintf("%v", periodo) == "map[]" {
		return errEmiter(err, fmt.Sprintf("%v", periodo))
	}

	//Primer o segundo semestre segun el ciclo
	if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "1" {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "PRIMER"
	} else {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "SEGUNDO"
	}

	dataHeader := map[string]interface{}{
		"Año":                periodo["Data"].(map[string]interface{})["Year"],
		"Semestre":           periodo["Data"].(map[string]interface{})["Ciclo"],
		"ProyectoCurricular": strings.ToUpper(fmt.Sprintf("%v", proyecto["Nombre"])),
		"Indices": []interface{}{
			"#",
			"Documento",
			"Nombre Completo",
			"Telefono",
			"Correo",
			"Puntaje",
			"Tipo de inscripción",
			"Enfasis",
			"Estado inscripción",
		},
	}

	var inscripciones []map[string]interface{}
	var errInscripciones error

	//Definir Ids de consulta en el CRUD de solicitudes
	if infoReporte.EstadoInscripcion == "solicitada" {
		infoReporte.EstadoInscripcion = "1"
	} else if infoReporte.EstadoInscripcion == "admitido" {
		infoReporte.EstadoInscripcion = "2"
	} else if infoReporte.EstadoInscripcion == "generada" {
		infoReporte.EstadoInscripcion = ""
	} else if infoReporte.EstadoInscripcion == "gestion" {
		infoReporte.EstadoInscripcion = ""
	} else if infoReporte.EstadoInscripcion == "aprobada" {
		infoReporte.EstadoInscripcion = "24"
	} else if infoReporte.EstadoInscripcion == "rechazada" {
		infoReporte.EstadoInscripcion = ""
	}

	if infoReporte.EstadoInscripcion == "solicitada" || infoReporte.EstadoInscripcion == "admitido" {
		//Hacer consulta especifica segun el tipo de inscripción y estado de inscripcion
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,TipoInscripcionId__Id:%v,EstadoInscripcionId:%v&limit=0", infoReporte.Proyecto, infoReporte.Periodo, infoReporte.TipoInscripcion, infoReporte.EstadoInscripcion), &inscripciones)
	} else {
		//Hacer consulta especifica segun el tipo de inscripción
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,TipoInscripcionId__Id:%v&limit=0", infoReporte.Proyecto, infoReporte.Periodo, infoReporte.TipoInscripcion), &inscripciones)
	}

	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "[map[]]" {
		return requestresponse.APIResponseDTO(false, 400, nil, "Falla inscripciones")
	} else {

		for _, inscripcion := range inscripciones {

			inscripcion := inscripcion
			estado := infoReporte.EstadoInscripcion

			//Datos basicos tercero
			tercero, err := obtenerInfoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
				return errEmiter(err)
			}

			//Obtener Documento Tercero
			terceroDocumento, err := obtenerDocumentoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", terceroDocumento) == "[map[]]" {
				return errEmiter(err)
			}
			//Obtener Telefono Tercero

			terceroTelefono := obtenerTelefonoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))

			//Obtener Correo Tercero
			terceroCorreo := obtenerCorreoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))

			//Obtener enfasis
			enfasis := obtenerEnfasis(fmt.Sprintf("%v", inscripcion["EnfasisId"]))

			//Definir data del inscrito
			inscrito := []interface{}{
				terceroDocumento[0]["Numero"],
				tercero[0]["NombreCompleto"],
				terceroTelefono,
				terceroCorreo,
				inscripcion["NotaFinal"],
				inscripcion["TipoInscripcionId"].(map[string]interface{})["Nombre"],
				enfasis}

			if estado != "solicitada" && estado != "admitido" {
				var solicitudes []map[string]interface{}
				errSolicitudes := request.GetJson("http://"+beego.AppConfig.String("SolicitudesService")+fmt.Sprintf("solicitud?query=Activo:true,EstadoTipoSolicitudId__TipoSolicitud__Id:25,EstadoTipoSolicitudId__EstadoId__Id:%v,Referencia__contains:%v&limit=0", estado, inscripcion["Id"]), &solicitudes)
				if errSolicitudes != nil || fmt.Sprintf("%v", solicitudes) == "[map[]]" {
				} else {
					for _, solicitud := range solicitudes {
						// Deserializa el JSON en un mapa
						var referencia map[string]interface{}
						if err := json.Unmarshal([]byte(solicitud["Referencia"].(string)), &referencia); err != nil {
							return errEmiter(err)
						}

						// Verifica si el InscripcionId está presente y es igual al buscado
						if referencia["InscripcionId"] == inscripcion["Id"] {
							inscrito = append(inscrito, solicitudes[0]["EstadoTipoSolicitudId"].(map[string]interface{})["EstadoId"].(map[string]interface{})["Nombre"])
							inscritos = append(inscritos, inscrito)
						}
					}
				}
			} else {
				inscrito = append(inscrito, inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"])
				inscritos = append(inscritos, inscrito)
			}

		}
	}
	return generarXlsxyPdfIncripciones(infoReporte, inscritos, dataHeader)

}

func generarXlsxyPdfIncripciones(infoReporte models.ReporteEstructura, inscritos [][]interface{}, dataHeader map[string]interface{}) requestresponse.APIResponse {

	//Abrir Plantilla Excel
	file, err := excelize.OpenFile("static/templates/ReporteInscritosOriginal.xlsx")
	if err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	//Organizar por Nombre
	sort.Slice(inscritos, func(i, j int) bool {
		return inscritos[i][2].(string) < inscritos[j][2].(string)
	})

	//Colocar indices al reporte
	for i, header := range dataHeader["Indices"].([]interface{}) {
		file.SetCellValue("Hoja1", fmt.Sprintf("%v7", string(rune(65+i))), header)
	}

	//Agregar data al reporte
	var lastCell = ""
	for i, row := range inscritos {
		dataRow := i + 8
		file.SetCellValue("Hoja1", fmt.Sprintf("A%v", dataRow), i+1)
		for j, col := range row {
			file.SetCellValue("Hoja1", fmt.Sprintf("%s%d", string(rune(65+j+1)), dataRow), col)
			lastCell = fmt.Sprintf("%s%d", string(rune(65+j+1)), dataRow)
		}
		file.SetRowHeight("Hoja1", dataRow, 30)
	}

	//creación de el estilo para el excel
	style2, err := file.NewStyle(
		&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
				WrapText:   true,
			},
			Border: []excelize.Border{
				{Type: "left", Color: "00000000", Style: 1},
				{Type: "right", Color: "00000000", Style: 1},
				{Type: "top", Color: "00000000", Style: 1},
				{Type: "bottom", Color: "00000000", Style: 1},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	file.SetCellStyle("Hoja1", "A8", lastCell, style2)

	//Redimensión de el excel para que el convertidor tome todas las celdas
	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A2:%v", lastCell))
	if errDimesion != nil {
		return errEmiter(errDimesion)
	}

	//Funcion reverse columans
	for i, j := 0, len(infoReporte.Columnas)-1; i < j; i, j = i+1, j-1 {
		infoReporte.Columnas[i], infoReporte.Columnas[j] = infoReporte.Columnas[j], infoReporte.Columnas[i]
	}

	//Eliminador de columnas
	for _, columna := range infoReporte.Columnas {
		file.RemoveCol("Hoja1", columna)
	}

	//Agregar datos de la cabecera

	if infoReporte.TipoReporte == 1 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE MATRICULADOS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 2 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE ADMITIDOS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 3 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE ASPIRANTES  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 5 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE TRANSFERENCIAS INTERNAS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 6 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE TRANSFERENCIAS EXTERNAS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 7 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE REINTEGROS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	}
	file.SetCellValue("Hoja1", "A6", fmt.Sprintf("PROYECTO CURRICULAR %v ORDENADO POR NOMBRE", dataHeader["ProyectoCurricular"]))

	//Definir ancho dinamico de las columnas
	//167.5 es el ancho total del reporte
	var anchoPorColumna = float64(167.5) / float64(10-len(infoReporte.Columnas))
	file.SetColWidth("Hoja1", "A", string(rune(65+(10-len(infoReporte.Columnas)))), anchoPorColumna)

	//Insertar header Xlsx
	if err := file.AddPicture("Hoja1", "A2", "static/images/HeaderEstaticoRecortado.jpg",
		&excelize.GraphicOptions{
			ScaleX:  0.20, //Escalado en x de la imagen
			ScaleY:  0.15, //Escalado en y de la imagen
			OffsetX: 2,    //Espacio entre la celda y la imagen para x
			OffsetY: 2,    //Espacio entre la celda y la imagen para y
		},
	); err != nil {
		errEmiter(err)
	}

	//Creación plantilla base
	pdf := gofpdf.New("", "", "", "")

	excelPdf := xlsx2pdf.Excel2PDF{
		Excel:    file,
		Pdf:      pdf,
		Sheets:   make(map[string]xlsx2pdf.SheetInfo),
		FontDims: xlsx2pdf.FontDims{Size: 0.85},
		Header:   func() {},
		CustomSize: xlsx2pdf.PageFormat{
			Orientation: "L",
			Wd:          600,
			Ht:          370,
		},
	}

	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/HeaderEstaticoRecortado.jpg", 25, 25, 320, 25, false, "", 0, "")
		}
	}

	excelPdf.ConvertSheets()

	/*if err := file.SaveAs("static/templates/ModificadoInscritos.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	err = pdf.OutputFileAndClose("static/templates/ReporteInscrito.pdf") //----> Si se guarda en local el PDF se borra de el buffer y no se genera el base 64
	if err != nil {
		return errEmiter(err)
	}*/

	//Conversión a base 64

	//Excel
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return errEmiter(err)
	}

	encodedFileExcel := base64.StdEncoding.EncodeToString(buffer.Bytes())

	//PDF
	var bufferPdf bytes.Buffer
	writer := bufio.NewWriter(&bufferPdf)
	pdf.Output(writer)
	writer.Flush()
	encodedFilePdf := base64.StdEncoding.EncodeToString(bufferPdf.Bytes())

	//Enviar respuesta
	respuesta := map[string]interface{}{
		"Excel": encodedFileExcel,
		"Pdf":   encodedFilePdf,
	}

	return requestresponse.APIResponseDTO(true, 200, respuesta)
}

func ReporteCaracterizacion(idPeriodo int64, idProyecto int64) requestresponse.APIResponse {
	var listado []map[string]interface{}
	var inscripcion []map[string]interface{}

	// Consulta de inscripciones activas en el periodo y proyecto proporcionados
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,PeriodoId:%v,Opcion:%v,EstadoInscripcionId.Id:11&limit=0", idPeriodo, idProyecto), &inscripcion)
	if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
		for _, inscrip := range inscripcion {
			var tercero map[string]interface{}
			errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero/%v", inscrip["PersonaId"]), &tercero)
			if errTercero == nil && tercero["Status"] != "404" {
				// Obtener datos de identificación
				var numeroIdentificacion string
				var identificacion []interface{}
				errDatosIdentificacion := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v,Activo:true&sortby=id&order=desc&limit=1&fields=Numero", tercero["Id"]), &identificacion)
				if errDatosIdentificacion == nil && len(identificacion) > 0 {
					numeroIdentificacion = identificacion[0].(map[string]interface{})["Numero"].(string)
				}
				// Obtener teléfono
				var numeroTelefonico string
				var telefono []interface{}
				errTelefono := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.Nombre:TELEFONO,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &telefono)
				if errTelefono == nil && len(telefono) > 0 {
					var telefonos map[string]interface{}
					err := json.Unmarshal([]byte(telefono[0].(map[string]interface{})["Dato"].(string)), &telefonos)
					if err == nil {
						if _, ok := telefonos["principal"]; ok {
							numeroTelefonico = fmt.Sprintf("%.f", telefonos["principal"].(float64))
						}
					} else {
						numeroTelefonico = telefono[0].(map[string]interface{})["Dato"].(string)
					}
				}
				// Obtener género
				var lugarResidencia string
				var residenciaData []interface{}
				errResidencia := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:LUGAR_RESIDENCIA,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &residenciaData)
				if errResidencia == nil && len(residenciaData) > 0 {
					if datoResidenciaMap, ok := residenciaData[0].(map[string]interface{}); ok {
						if datoResidenciaStr, exists := datoResidenciaMap["Dato"]; exists {
							// Verificar si datoGeneroStr es un string JSON
							if str, ok := datoResidenciaStr.(string); ok {
								// Deserializar el JSON
								var datoResidencia map[string]interface{}
								if err := json.Unmarshal([]byte(str), &datoResidencia); err == nil {
									// Aquí se asume que 'dato' tiene el valor del género
									if value, exists := datoResidencia["dato"]; exists {
										lugarResidencia = fmt.Sprintf("%v", value) // Convertir a string
									}
								}
							}
						}
					}
				} else {
					lugarResidencia = "Sin datos registrados"
				}
				// Obtener Documento Tercero
				var terceroDocumento []map[string]interface{}
				errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:%v", tercero["Id"]), &terceroDocumento)
				if errTerceroDocumento != nil || len(terceroDocumento) == 0 {
					return errEmiter(errTerceroDocumento, fmt.Sprintf("%v", terceroDocumento))
				}
				var tipoDocumento string
				if len(terceroDocumento) > 0 {
					tipoDocumento = terceroDocumento[0]["TipoDocumentoId"].(map[string]interface{})["Nombre"].(string)
				}
				// Obtener Nombre del Estado de Inscripción
				estadoInscripcionNombre := inscrip["EstadoInscripcionId"].(map[string]interface{})["Nombre"]
				// Obtener discapacidad
				var discapacidad string
				var discapacidadData []interface{}
				errDiscapacidad := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:DISCAPACIDAD,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &discapacidadData)
				if errDiscapacidad == nil && len(discapacidadData) > 0 {
					if datoDiscapacidadMap, ok := discapacidadData[0].(map[string]interface{}); ok {
						if datoDiscapacidadStr, exists := datoDiscapacidadMap["Dato"]; exists {
							// Deserializar el JSON
							var datoDiscapacidad map[string]interface{}
							if err := json.Unmarshal([]byte(fmt.Sprintf("%v", datoDiscapacidadStr)), &datoDiscapacidad); err == nil {
								if dato, exists := datoDiscapacidad["dato"]; exists {
									discapacidad = fmt.Sprintf("%v", dato) // Convertir a string
								}
							}
						}
					}
				} else {
					discapacidad = "Sin datos registrados"
				}
				// Obtener nombre del colegio
				var nombreColegio string
				var colegioData []interface{}
				errColegio := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:NOMBRE_COLEGIO,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &colegioData)
				if errColegio == nil && len(colegioData) > 0 {
					if datoColegioMap, ok := colegioData[0].(map[string]interface{}); ok {
						if datoColegioStr, exists := datoColegioMap["Dato"]; exists {
							// Deserializar el JSON
							var datoColegio map[string]interface{}
							if err := json.Unmarshal([]byte(fmt.Sprintf("%v", datoColegioStr)), &datoColegio); err == nil {
								if dato, exists := datoColegio["dato"]; exists {
									nombreColegio = fmt.Sprintf("%v", dato) // Convertir a string
								}
							}
						}
					}
				} else {
					nombreColegio = "Sin datos registrados"
				}
				// Obtener género
				var tipoColegio string
				var tipoColegioData []interface{}
				errTipoColegioData := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:OFICIAL,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &tipoColegioData)
				if errTipoColegioData == nil && len(tipoColegioData) > 0 {
					if datoTipoColegioMap, ok := tipoColegioData[0].(map[string]interface{}); ok {
						if datoTipoColegioStr, exists := datoTipoColegioMap["Dato"]; exists {
							// Verificar si datoGeneroStr es un string JSON
							if str, ok := datoTipoColegioStr.(string); ok {
								// Deserializar el JSON
								var datoTipoColegio map[string]interface{}
								if err := json.Unmarshal([]byte(str), &datoTipoColegio); err == nil {
									// Aquí se asume que 'dato' tiene el valor del género
									if value, exists := datoTipoColegio["dato"]; exists {
										tipoColegio = fmt.Sprintf("%v", value) // Convertir a string
									}
								}
							}
						}
					}
				} else {
					tipoColegio = "Sin datos registrados"
				}
				// Obtener género
				var genero string
				var generoData []interface{}
				errGenero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:MASCULINO,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &generoData)
				if errGenero == nil && len(generoData) > 0 {
					if datoGeneroMap, ok := generoData[0].(map[string]interface{}); ok {
						if datoGeneroStr, exists := datoGeneroMap["Dato"]; exists {
							// Verificar si datoGeneroStr es un string JSON
							if str, ok := datoGeneroStr.(string); ok {
								// Deserializar el JSON
								var datoGenero map[string]interface{}
								if err := json.Unmarshal([]byte(str), &datoGenero); err == nil {
									// Aquí se asume que 'dato' tiene el valor del género
									if value, exists := datoGenero["dato"]; exists {
										genero = fmt.Sprintf("%v", value) // Convertir a string
									}
								}
							}
						}
					}
				} else {
					genero = "Sin datos registrados"
				}
				// Obtener estrato
				var estrato string
				var estratoData []interface{}
				errEstrato := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId:%v,InfoComplementariaId.CodigoAbreviacion:ESTRATO,Activo:true&sortby=id&order=desc&limit=1&fields=Dato", tercero["Id"]), &estratoData)
				if errEstrato == nil && len(estratoData) > 0 {
					if datoEstratoMap, ok := estratoData[0].(map[string]interface{}); ok {
						if datoEstratoStr, exists := datoEstratoMap["Dato"]; exists {
							// Verificar si datoEstratoStr es un string JSON
							if str, ok := datoEstratoStr.(string); ok {
								// Deserializar el JSON
								var datoEstrato map[string]interface{}
								if err := json.Unmarshal([]byte(str), &datoEstrato); err == nil {
									if value, exists := datoEstrato["value"]; exists {
										estrato = fmt.Sprintf("%v", value) // Convertir a string
									}
								}
							}
						}
					}
				} else {
					estrato = "Sin datos registrados"
				}
				// Concatenar Nombres y Apellidos
				nombres := fmt.Sprintf("%s %s", tercero["PrimerNombre"], tercero["SegundoNombre"])
				apellidos := fmt.Sprintf("%s %s", tercero["PrimerApellido"], tercero["SegundoApellido"])

				listado = append(listado, map[string]interface{}{
					"IdTercero":         tercero["Id"],
					"Nombres":           nombres,
					"Apellidos":         apellidos,
					"Numero":            numeroIdentificacion,
					"CorreoPersonal":    tercero["UsuarioWSO2"],
					"Telefono":          numeroTelefonico,
					"TipoDocumento":     tipoDocumento,
					"EstadoInscripcion": estadoInscripcionNombre,
					"LugarResidencia":   lugarResidencia,
					"Discapacidad":      discapacidad,
					"Colegio":           nombreColegio,
					"TipoColegio":       tipoColegio,
					"Estrato":           estrato,
					"Genero":            genero,
				})
			}
		}
		return requestresponse.APIResponseDTO(true, 200, listado)
	} else {
		return requestresponse.APIResponseDTO(false, 404, "No se encontraron datos relacionados con el periodo y proyecto proporcionados")
	}
}
