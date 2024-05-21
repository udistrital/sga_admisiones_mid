package services

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/phpdave11/gofpdf"
	"github.com/udistrital/sga_admisiones_mid/helpers"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/time_bogota"
	"github.com/udistrital/utils_oas/xlsx2pdf"
	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)

func InformeLiquidacionPregrado(idPeriodo int64, idProyecto int64) (APIResponseDTO requestresponse.APIResponse) {

	admitidos := ListarLiquidacionEstudiantes(idPeriodo, idProyecto).Data.([]map[string]interface{})
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
		dataRow := i + 8
		numeroRegistros := 1 + i
		indx = dataRow
		file.SetCellValue("Hoja1", "A"+strconv.Itoa(dataRow), numeroRegistros)
		file.SetCellValue("Hoja1", "B"+strconv.Itoa(dataRow), row["PrimerApellido"])
		file.SetCellValue("Hoja1", "C"+strconv.Itoa(dataRow), row["SegundoApellido"])
		file.SetCellValue("Hoja1", "D"+strconv.Itoa(dataRow), row["Nombre"])
		file.SetCellValue("Hoja1", "E"+strconv.Itoa(dataRow), row["Estado"])
		file.SetCellValue("Hoja1", "F"+strconv.Itoa(dataRow), row["Documento"])
		file.SetCellValue("Hoja1", "G"+strconv.Itoa(dataRow), row["Codigo"])
		file.SetCellValue("Hoja1", "H"+strconv.Itoa(dataRow), row["Descuentos"])
		file.SetCellStyle("Hoja1", "A"+strconv.Itoa(dataRow), "H"+strconv.Itoa(dataRow), styleID)

	}

	errDimesion := file.SetSheetDimension("Hoja1", fmt.Sprintf("A1:Af%d", indx))
	if errDimesion != nil {
		return helpers.ErrEmiter(errDimesion)
	}

	if err := file.SaveAs("static/templates/TemplateInformePregradoModificado.xlsx"); err != nil {
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

func ListarLiquidacionEstudiantes(idPeriodo int64, idProyecto int64) (APIResponseDTO requestresponse.APIResponse) {
	//Mapa para guardar los admitidos
	var admitidos []map[string]interface{}

	//Obtener Datos del periodo
	var periodo map[string]interface{}
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+fmt.Sprintf("periodo/%v", idPeriodo), &periodo)
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "[map[]]" {
		return helpers.ErrEmiter(errPeriodo, fmt.Sprintf("%v", periodo))
	}

	//Obtener Datos del proyecto & facultad
	var facultad map[string]interface{}

	var proyecto map[string]interface{}
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_institucion/%v", idProyecto), &proyecto)
	if errProyecto != nil || fmt.Sprintf("%v", proyecto) == "map[]" {
		return helpers.ErrEmiter(errProyecto, fmt.Sprintf("%v", proyecto))
	} else {
		//Obtener Datos de la facultad
		errFacultad := request.GetJson("http://"+beego.AppConfig.String("OikosService")+fmt.Sprintf("dependencia/%v", proyecto["FacultadId"]), &facultad)
		if errFacultad != nil || fmt.Sprintf("%v", facultad) == "map[]" {
			return helpers.ErrEmiter(errFacultad, fmt.Sprintf("%v", facultad))
		}
	}

	//Inscripciones de admitidos
	var inscripciones []map[string]interface{}
	errInscripciones := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:ADMITIDO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", idProyecto, idPeriodo), &inscripciones)
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "map[]" {
		return helpers.ErrEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	}

	//Base para la comparación de codigo
	if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "3" {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "2"
	}
	codigoBase := fmt.Sprintf("%v%v%v", periodo["Data"].(map[string]interface{})["Year"], periodo["Data"].(map[string]interface{})["Ciclo"], proyecto["Codigo"])

	// Procesar cada inscripción
	for _, inscripcion := range inscripciones {
		// Variables para almacenar los datos de cada inscripción
		var tercero []map[string]interface{}
		var terceroDocumento []map[string]interface{}
		var terceroCodigo []map[string]interface{}
		var descuentos []map[string]interface{}

		// Obtener datos básicos del tercero
		errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscripcion["PersonaId"]), &tercero)
		if errTercero != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return helpers.ErrEmiter(errTercero, fmt.Sprintf("%v", tercero))
		}

		// Obtener documento del tercero
		errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:%v", inscripcion["PersonaId"]), &terceroDocumento)
		if errTerceroDocumento != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return helpers.ErrEmiter(errTerceroDocumento, fmt.Sprintf("%v", terceroDocumento))
		}

		// Obtener código del tercero
		errTerceroCodigo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CODE,Activo:true,TerceroId:%v,Numero__contains:%v", inscripcion["PersonaId"], codigoBase), &terceroCodigo)
		if errTerceroCodigo != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return helpers.ErrEmiter(errTerceroCodigo, fmt.Sprintf("%v", terceroCodigo))
		}

		// Obtener descuentos
		fmt.Println("http://" + beego.AppConfig.String("DescuentosService") + fmt.Sprintf("solicitud_descuento?query=Activo:true,TerceroId:%v,PeriodoId:%v", inscripcion["PersonaId"], idPeriodo))
		errDescuentos := request.GetJson("http://"+beego.AppConfig.String("DescuentosService")+fmt.Sprintf("solicitud_descuento?query=Activo:true,TerceroId:%v,PeriodoId:%v", inscripcion["PersonaId"], idPeriodo), &descuentos)
		if errDescuentos != nil || fmt.Sprintf("%v", descuentos) == "[map[]]" {
			return helpers.ErrEmiter(errDescuentos, fmt.Sprintf("%v", descuentos))
		}

		var descuentosInfo []interface{}

		for _, descuento := range descuentos {
			tipoDescuento := descuento["DescuentosDependenciaId"].(map[string]interface{})["TipoDescuentoId"].(map[string]interface{})
			id := tipoDescuento["Id"]
			descuentosInfo = append(descuentosInfo, id)
		}

		// Agregar la información del admitido al slice admitidos
		admitidos = append(admitidos, map[string]interface{}{
			"Nombre":          fmt.Sprintf("%v %v", tercero[0]["PrimerNombre"], tercero[0]["SegundoNombre"]),
			"PrimerApellido":  tercero[0]["PrimerApellido"],
			"SegundoApellido": tercero[0]["SegundoApellido"],
			"Estado":          "Admitido",
			"Documento":       terceroDocumento[0]["Numero"],
			"Codigo":          terceroCodigo[0]["Numero"],
			"Descuentos":      descuentosInfo, // Se almacenan todos los descuentos para esta inscripción
		})
	}

	return requestresponse.APIResponseDTO(false, 200, admitidos, "Admitidos")
}
func CrearLiquidacion(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	//Almacena la nueva noticia
	var nuevaLiquidacion map[string]interface{}
	var nuevoConcepto map[string]interface{}
	var nuevoRecibo map[string]interface{}
	var errSaveAll bool
	//respuesta a la petición
	var respuesta map[string]interface{}
	//timestamp
	date := time_bogota.TiempoBogotaFormato()

	if err := json.Unmarshal(data, &nuevaLiquidacion); err == nil {

		dataLiquidacion := map[string]interface{}{
			"activo":                true,
			"fecha_creacion":        date,
			"fecha_modificacion":    date,
			"tercero_id":            nuevaLiquidacion["tercero_id"].(float64),
			"periodo_id":            nuevaLiquidacion["periodo_id"].(float64),
			"programa_academico_id": nuevaLiquidacion["programa_academico_id"].(float64),
			"tipo_programa_id":      nuevaLiquidacion["tipo_programa_id"].(float64),
		}
		//var guardada map[string]interface{}
		nuevaLiquidacion = dataLiquidacion

		errLiquidacion := request.SendJson("http://"+beego.AppConfig.String("liquidacionService")+"liquidacion/", "POST", &nuevaLiquidacion, dataLiquidacion)
		if errLiquidacion == nil {
			fmt.Println(dataLiquidacion)

			liqId := nuevaLiquidacion["Data"].(map[string]interface{})["_id"].(string)
			fmt.Println("---------------------------------")
			fmt.Println(liqId)
			fmt.Println("---------------------------------")
			if err := json.Unmarshal(data, &nuevaLiquidacion); err == nil {
				//fmt.Println(nuevaNoticia["Contenido"])

				concepto, conceptoExist := nuevaLiquidacion["liqDetalle"]
				if conceptoExist {

					conceptos := concepto.([]interface{})
					fmt.Println(conceptos)

					for _, c := range conceptos {
						contenidoMap := c.(map[string]interface{})
						dataConcepto := map[string]interface{}{
							"tipo_concepto_id":   contenidoMap["tipo_concepto_id"].(float64),
							"valor":              contenidoMap["valor"].(float64),
							"Activo":             true,
							"fecha_creacion":     date,
							"fecha_modificacion": date,
							"liquidacion_id":     liqId,
						}
						fmt.Println(dataConcepto)

						errConcepto := request.SendJson("http://"+beego.AppConfig.String("liquidacionService")+"liquidacion-detalle/", "POST", &nuevoConcepto, dataConcepto)
						if errConcepto != nil {
							//errSaveAll = true
						}
					}
					if !errSaveAll {
						if err := json.Unmarshal(data, &nuevaLiquidacion); err == nil {

							dataRecibo := map[string]interface{}{
								"Activo":             true,
								"fecha_creacion":     date,
								"fecha_modificacion": date,
								"liquidacion_id":     liqId,
								"recibo_id":          nuevaLiquidacion["recibo_id"].(float64),
							}

							errEtiqueta := request.SendJson("http://"+beego.AppConfig.String("liquidacionService")+"liquidacion-recibo/", "POST", &nuevoRecibo, dataRecibo)
							if errEtiqueta != nil {
								errSaveAll = true
							}

							if !errSaveAll {

								APIResponseDTO = requestresponse.APIResponseDTO(true, 200, nuevoRecibo)
								return APIResponseDTO
							}

						}
						APIResponseDTO = requestresponse.APIResponseDTO(true, 200, nuevaLiquidacion)
						return APIResponseDTO
					}
				}
			}
			APIResponseDTO = requestresponse.APIResponseDTO(true, 200, respuesta, nuevaLiquidacion)
			return APIResponseDTO
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 500, respuesta, nuevaLiquidacion)
			return APIResponseDTO
		}
	}

	APIResponseDTO = requestresponse.APIResponseDTO(true, 200, respuesta, nuevaLiquidacion)
	return APIResponseDTO
}

func GetAllLiquidaciones() (APIResponseDTO requestresponse.APIResponse) {
	fmt.Println("GetAll")
	var liquidacion interface{}
	wge := new(errgroup.Group)
	var mutex sync.Mutex // Mutex para proteger el acceso a resultados

	errLiquidacion := request.GetJson("http://"+beego.AppConfig.String("liquidacionService")+fmt.Sprintf("liquidacion?=activo:true&limit=0"), &liquidacion)
	if errLiquidacion == nil {

		if data, ok := liquidacion.(map[string]interface{}); ok {

			if liquidaciones, ok := data["Data"].([]interface{}); ok {
				var liquidacionesSlice []interface{}
				wge.SetLimit(-1)
				for _, l := range liquidaciones {
					l := l
					wge.Go(func() error {

						if liquidacionData, ok := l.(map[string]interface{}); ok {
							liquidacionInfo := make(map[string]interface{})
							liquidacionInfo["_id"] = liquidacionData["_id"]
							liquidacionInfo["tercero_id"] = liquidacionData["tercero_id"]
							liquidacionInfo["periodo_id"] = liquidacionData["periodo_id"]
							liquidacionInfo["programa_academico_id"] = liquidacionData["programa_academico_id"]
							liquidacionInfo["tipo_programa_id"] = liquidacionData["tipo_programa_id"]

							// Obtener detalles de liquidación para esta liquidación
							var liqDetalles interface{}
							errLiqDetalle := request.GetJson("http://"+beego.AppConfig.String("liquidacionService")+fmt.Sprintf("liquidacion-detalle?liquidacion_id=%v", liquidacionData["_id"]), &liqDetalles)
							if errLiqDetalle == nil {
								//fmt.Println("Detalles de liquidación obtenidos con éxito:", liqDetalles)

								if data, ok := liqDetalles.(map[string]interface{}); ok {
									//fmt.Println("Data obtenida:", data)

									if detalles, ok := data["Data"].([]interface{}); ok {
										var detallesFiltrados []interface{}
										for _, detalle := range detalles {
											detalleMap, ok := detalle.(map[string]interface{})
											if !ok {
												continue // Salta este detalle si no es un mapa
											}
											liquidacionID, ok := detalleMap["liquidacion_id"].(string)
											if !ok || liquidacionID == "" {
												continue // Salta este detalle si liquidacion_id no es un string o está vacío
											}
											if liquidacionID == liquidacionData["_id"] {
												detallesFiltrados = append(detallesFiltrados, detalleMap)
											}
										}
										liquidacionInfo["detalles"] = detallesFiltrados
									} else {
										return errors.New("No se encontraron detalles en la respuesta")
									}
								} else {
									return errors.New("La respuesta JSON no es un objeto")
								}
							} else {
								return errLiqDetalle
							}

							// Obtener recibo de liquidación para esta liquidación
							var liqRecibo interface{}
							errLiqRecibo := request.GetJson("http://"+beego.AppConfig.String("liquidacionService")+fmt.Sprintf("liquidacion-recibo?liquidacion_id=%v", liquidacionData["_id"]), &liqRecibo)
							if errLiqRecibo == nil {
								if data, ok := liqRecibo.(map[string]interface{}); ok {
									if recibos, ok := data["Data"].([]interface{}); ok {

										var reciboFiltrado []interface{}
										for _, recibo := range recibos {
											reciboMap, ok := recibo.(map[string]interface{})
											if !ok {
												continue // Salta este recibo si no es un mapa
											}
											liquidacionID, ok := reciboMap["liquidacion_id"].(string)
											if !ok || liquidacionID == "" {
												continue // Salta este recibo si liquidacion_id no es un string o está vacío
											}
											if liquidacionID == liquidacionData["_id"] {
												reciboFiltrado = append(reciboFiltrado, reciboMap)
											}
										}
										liquidacionInfo["recibo"] = reciboFiltrado
									}
								}
							} else {
								return errLiqRecibo
							}

							mutex.Lock()
							liquidacionesSlice = append(liquidacionesSlice, liquidacionInfo)
							mutex.Unlock()
						}
						return nil
					})
				}
				if err := wge.Wait(); err != nil {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err)
					return APIResponseDTO
				}
				APIResponseDTO = requestresponse.APIResponseDTO(true, 200, liquidacionesSlice)
			} else {
				APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, "No se encontró la clave 'Data' en la respuesta JSON")
			}
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, "La respuesta JSON no es un objeto")
		}
	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, errLiquidacion.Error())
	}
	return APIResponseDTO
}
