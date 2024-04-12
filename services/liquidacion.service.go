package services

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/helpers"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/time_bogota"
)

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
	//return requestresponse.APIResponseDTO(false, 400, nil, "Error final")
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