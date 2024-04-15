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
	"github.com/phpdave11/gofpdf"
	"github.com/udistrital/sga_admisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/xlsx2pdf"
	"github.com/xuri/excelize/v2"
)

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
			return errEmiter(errTercero, fmt.Sprintf("%v", tercero))
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
		return errEmiter(err)
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
		return errEmiter(errDimesion)
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
		return errEmiter(err)
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
		return errEmiter(err)
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
*/

func ReporteDinamico(data []byte) requestresponse.APIResponse {
	var reporte models.ReporteEstructura
	var respuesta requestresponse.APIResponse
	if err := json.Unmarshal(data, &reporte); err == nil {
		if reporte.TipoReporte != 0 {
			respuesta = reporteInscritosPorPrograma(reporte)
		}

	} else {
		respuesta = requestresponse.APIResponseDTO(true, 200, nil)
	}

	return respuesta
}

func columnasParaEliminar(columnasSolicitadas []string) []string {
	columnasMaximas := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}

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

//Funcion para reporte de Inscritos por programa
func reporteInscritosPorPrograma(infoReporte models.ReporteEstructura) requestresponse.APIResponse {

	var inscritos [][]interface{}

	//Definir Columnas a eliminar
	infoReporte.Columnas = columnasParaEliminar(infoReporte.Columnas)
	fmt.Println("columnas: " + fmt.Sprintf("%v", infoReporte.Columnas))

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
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:INSCRITO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", infoReporte.Proyecto, infoReporte.Periodo), &inscripciones)

	} else if infoReporte.TipoReporte == 2 {

		//Añadir headers no compartidos
		dataHeader["Indices"] = append(dataHeader["Indices"].([]interface{}),
			"Credencial",
			"Enfasis",
			"Estado inscripción",
			"Puntaje")

		//Hacer consulta especifica para estado ADMITIDO U OPCIONADO
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre__in:ADMITIDO|OPCIONADO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", infoReporte.Proyecto, infoReporte.Periodo), &inscripciones)
	}else{
		//Añadir headers no compartidos
		dataHeader["Indices"] = append(dataHeader["Indices"].([]interface{}),
			"Tipo inscripción",
			"Estado inscripción")

		//Hacer consulta especifica para aspirantes
		errInscripciones = request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", infoReporte.Proyecto, infoReporte.Periodo), &inscripciones)
	}

	//Si existen inscripciones entonces
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "[map[]]" {
		return errEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	} else {

		fmt.Println("Hay Informacion")
		for _, inscripcion := range inscripciones {
			//Datos basicos tercero
			tercero, err := obtenerInfoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
				return errEmiter(err)
			}
			fmt.Println(inscripcion["PersonaId"])

			//Obtener Documento Tercero
			terceroDocumento, err := obtenerDocumentoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil || fmt.Sprintf("%v", terceroDocumento) == "[map[]]" {
				return errEmiter(err)
			}
			fmt.Println("Hay Documento")

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
			}else {
				inscrito = append(inscrito, inscripcion["TipoInscripcionId"].(map[string]interface{})["Nombre"], inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"])
			}
			inscritos = append(inscritos, inscrito)

		}

		return generarXlsxyPdfIncripciones(infoReporte, inscritos, dataHeader)

	}

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

	

	fmt.Println("Colocar indices")
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
			fmt.Println(lastCell)
		}
		file.SetRowHeight("Hoja1", dataRow, 30)
	}
	fmt.Println("Data colocada")

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

	fmt.Println(lastCell)
	//Redimensión de el excel para que el convertidor tome todas las celdas
	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A2:%v", lastCell))
	if errDimesion != nil {
		return errEmiter(errDimesion)
	}
	fmt.Println("Redimensión")


	if infoReporte.TipoReporte == 1 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE MATRICULADOS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else if infoReporte.TipoReporte == 2 {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE ADMITIDOS  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	} else {
		file.SetCellValue("Hoja1", "A5", fmt.Sprintf("LISTADO DE ASPIRANTES  PARA EL %v SEMESTRE ACADÉMICO DEL AÑO %v", dataHeader["Semestre"], dataHeader["Año"]))
	}
	file.SetCellValue("Hoja1", "A6", fmt.Sprintf("PROYECTO CURRICULAR %v ORDENADO POR NOMBRE", dataHeader["ProyectoCurricular"]))

	//Funcion reverse columans
	for i, j := 0, len(infoReporte.Columnas)-1; i < j; i, j = i+1, j-1 {
		infoReporte.Columnas[i], infoReporte.Columnas[j] = infoReporte.Columnas[j], infoReporte.Columnas[i]
	}

	//156.5 es el ancho que abarca el reporte con todas las columnas
	var anchoTotal float64
	for _, columna := range infoReporte.Columnas {
		if width, err := file.GetColWidth("Hoja1", columna); err == nil {
			anchoTotal += width
		}
		file.RemoveCol("Hoja1", columna)
		//file.SetColVisible("Hoja1", columna, false)
	}

	//Definir ancho dinamico de las columnas
	//145 es el ancho a distribuir sin la columna A por lo tanto
	var anchoPorColumna = float64(145) / float64(8-len(infoReporte.Columnas))
	file.SetColWidth("Hoja1", "B", string(rune(65+(8-len(infoReporte.Columnas)))), anchoPorColumna)

	//Insertar header Xlsx
	if err := file.AddPicture("Hoja1", "A2", "static/images/HeaderEstaticoRecortado.jpg",
		&excelize.GraphicOptions{
			ScaleX:  0.19, //Escalado en x de la imagen
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
			Ht:          350,
		},
	}

	excelPdf.Header = func() {
		if excelPdf.PageCount == 1 {
			pdf.Image("static/images/HeaderEstaticoRecortado.jpg", 25, 25, 300, 25, false, "", 0, "")
		}
	}

	excelPdf.ConvertSheets()

	if err := file.SaveAs("static/templates/ModificadoInscritos.xlsx"); err != nil {
		log.Fatal(err)
		return errEmiter(err)
	}

	err = pdf.OutputFileAndClose("static/templates/ReporteInscrito.pdf") //----> Si se guarda en local el PDF se borra de el buffer y no se genera el base 64
	if err != nil {
		return errEmiter(err)
	}

	/*/Conversión a base 64

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
	}*/

	return requestresponse.APIResponseDTO(true, 200, nil)
}
