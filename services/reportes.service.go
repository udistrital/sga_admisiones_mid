package services

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sort"

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

	return requestresponse.APIResponseDTO(false, 400, nil)
}

func GenerarReporteCodigos(idPeriodo int64, idProyecto int64) requestresponse.APIResponse {
	//Mapa para guardar los admitidos
	var admitidos []map[string]interface{}

	//Obtener Datos del periodo
	var periodo map[string]interface{}
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+fmt.Sprintf("periodo/%v", idPeriodo), &periodo)
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "[map[]]" {
		return errEmiter(errPeriodo, fmt.Sprintf("%v", periodo))
	}

	//Obtener Datos del proyecto & facultad
	var facultad map[string]interface{}

	var proyecto map[string]interface{}
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_institucion/%v", idProyecto), &proyecto)
	if errProyecto != nil || fmt.Sprintf("%v", proyecto) == "map[]" {
		return errEmiter(errProyecto, fmt.Sprintf("%v", proyecto))
	} else {
		//Obtener Datos de la facultad
		errFacultad := request.GetJson("http://"+beego.AppConfig.String("OikosService")+fmt.Sprintf("dependencia/%v", proyecto["FacultadId"]), &facultad)
		if errFacultad != nil || fmt.Sprintf("%v", facultad) == "map[]" {
			return errEmiter(errFacultad, fmt.Sprintf("%v", facultad))
		}
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
	errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:"+idTercero, &terceroDocumento)
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
			telefono = fmt.Sprintf("%v", telefonoPrincipal["principal"])
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
	if err := json.Unmarshal(data, &reporte); err == nil {
		if reporte.TipoReporte != 0 {
			switch reporte.TipoReporte {
			case 1:
				return reporteInscritosPorPrograma(reporte)
			}
		}
		fmt.Println(fmt.Sprintf("TipoReporte: %v", reporte.TipoReporte))

	} else {
		fmt.Println(err.Error())
	}
	return requestresponse.APIResponseDTO(true, 200, nil)
}

//Funcion para reporte de Inscrfitos por prohrama
func reporteInscritosPorPrograma(infoReporte models.ReporteEstructura) requestresponse.APIResponse {

	var inscritosMap []map[string]interface{}

	//Inscripciones en estado inscrito
	var inscripciones []map[string]interface{}
	errInscripciones := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:INSCRITO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", infoReporte.Proyecto, infoReporte.Periodo), &inscripciones)
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "[map[]]" {
		return errEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	} else {

		for _, inscripcion := range inscripciones {

			//Datos basicos tercero
			tercero, err := obtenerInfoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil {
				return errEmiter(err)
			}

			//Obtener Documento Tercero
			terceroDocumento, err := obtenerDocumentoTercero(fmt.Sprintf("%v", inscripcion["PersonaId"]))
			if err != nil {
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

			inscritosMap = append(inscritosMap, map[string]interface{}{
				"Documento":  terceroDocumento[0]["Numero"],
				"Nombre":     tercero[0]["NombreCompleto"],
				"Telefono":   terceroTelefono,
				"Correo":     terceroCorreo,
				"Credencial": inscripcion["Id"],
				"Enfasis":    enfasis,
				"Descuento":  nombreDescuento,
				"Estado":     inscripcion["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
			})

		}

		//Abrir Plantilla Excel
		file, err := excelize.OpenFile("static/templates/ReporteInscritos.xlsx")
		if err != nil {
			log.Fatal(err)
			return errEmiter(err)
		}

		var anchoTotal float64

		//Funcion reverse columans
		for i, j := 0, len(infoReporte.Columnas)-1; i < j; i, j = i+1, j-1 {
			infoReporte.Columnas[i], infoReporte.Columnas[j] = infoReporte.Columnas[j], infoReporte.Columnas[i]
		}
		
		for _, columna := range infoReporte.Columnas {
			if width, err := file.GetColWidth("Hoja1", columna); err == nil {
				anchoTotal += width
				fmt.Println("Ancho: " + fmt.Sprintf("%v", width))
			}
			fmt.Println("Columna: " + columna)
			//file.RemoveCol("Hoja1", columna)
			file.SetColVisible("Hoja1", columna, false)
		}

		if err := file.SaveAs("static/templates/ModificadoInscritos.xlsx"); err != nil {
			log.Fatal(err)
			return errEmiter(err)
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
				pdf.Image("static/images/HeaderEstaticoRecortado.png", 25, 25, 300, 25, false, "", 0, "")
			}
		}

		excelPdf.ConvertSheets()

		err = pdf.OutputFileAndClose("static/templates/ReporteInscrito.pdf")
		if err != nil {
			return errEmiter(err)
		}
		return requestresponse.APIResponseDTO(true, 200, inscritosMap)

	}

}
