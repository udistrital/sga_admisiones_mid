package services

import (
	"fmt"
	"math"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_mid_admisiones/models"
	"github.com/udistrital/utils_oas/request"
)

// FUNCIONES QUE SE USAN EN PUT NOTA FINAL ASPIRANTES

func solicitudInscripcionPut(InscripcionId string, InscripcionPut map[string]interface{}, Inscripcion []map[string]interface{}, respuesta *[]map[string]interface{}, i int, alerta models.Alert, alertas []interface{}, errorGetAll bool) interface{} {
	errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, "PUT", &InscripcionPut, Inscripcion[0])
	if errInscripcionPut == nil {
		if InscripcionPut != nil && fmt.Sprintf("%v", InscripcionPut) != "map[]" {
			(*respuesta)[i] = InscripcionPut
			return nil
		} else {
			ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
			return map[string]interface{}{"Response": alerta}
		}
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "", errInscripcionPut)
		return map[string]interface{}{"Response": alerta}
	}
}

func validarDetalleEvaluacionPut(DetalleEvaluacion []map[string]interface{}, NotaFinal float64, Inscripcion []map[string]interface{}, InscripcionId string, InscripcionPut map[string]interface{}, respuesta []map[string]interface{}, i int, alerta models.Alert, alertas []interface{}, errorGetAll bool) interface{} {
	if DetalleEvaluacion != nil && fmt.Sprintf("%v", DetalleEvaluacion[0]) != "map[]" {
		NotaFinal = 0
		// Calculo de la nota Final con los criterios relacionados al proyecto
		for _, EvaluacionAux := range DetalleEvaluacion {
			f, _ := strconv.ParseFloat(fmt.Sprintf("%v", EvaluacionAux["NotaRequisito"]), 64)
			NotaFinal = NotaFinal + f
		}
		NotaFinal = math.Round(NotaFinal*100) / 100
		Inscripcion[0]["NotaFinal"] = NotaFinal

		//PUT a inscripción con la nota final calculada
		return solicitudInscripcionPut(InscripcionId, InscripcionPut, Inscripcion, &respuesta, i, alerta, alertas, errorGetAll)
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
		return map[string]interface{}{"Response": alerta}
	}
}

func solicitudDetalleEvaluacionPut(InscripcionId string, ProgramaAcademicoId string, PeriodoId string, DetalleEvaluacion []map[string]interface{}, NotaFinal float64, Inscripcion []map[string]interface{}, InscripcionPut map[string]interface{}, respuesta []map[string]interface{}, i int, alerta models.Alert, alertas []interface{}, errorGetAll bool) interface{} {
	errDetalleEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=InscripcionId:"+InscripcionId+",RequisitoProgramaAcademicoId__ProgramaAcademicoId:"+ProgramaAcademicoId+",RequisitoProgramaAcademicoId__PeriodoId:"+PeriodoId+"&limit=0", &DetalleEvaluacion)
	if errDetalleEvaluacion == nil {
		return validarDetalleEvaluacionPut(DetalleEvaluacion, NotaFinal, Inscripcion, InscripcionId, InscripcionPut, respuesta, i, alerta, alertas, errorGetAll)
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "", errDetalleEvaluacion)
		return map[string]interface{}{"Response": alerta}
	}
}

func SolicitudIdPut(PersonaId string, PeriodoId string, ProgramaAcademicoId string, Inscripcion []map[string]interface{}, DetalleEvaluacion []map[string]interface{}, NotaFinal float64, InscripcionPut map[string]interface{}, respuesta []map[string]interface{}, i int, alerta models.Alert, alertas []interface{}, errorGetAll bool) interface{} {
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=PersonaId:"+PersonaId+",PeriodoId:"+PeriodoId+",ProgramaAcademicoId:"+ProgramaAcademicoId, &Inscripcion)
	if errInscripcion == nil {
		if Inscripcion != nil && fmt.Sprintf("%v", Inscripcion[0]) != "map[]" {
			InscripcionId := fmt.Sprintf("%v", Inscripcion[0]["Id"])

			//GET a detalle evaluacion
			return solicitudDetalleEvaluacionPut(InscripcionId, ProgramaAcademicoId, PeriodoId, DetalleEvaluacion, NotaFinal, Inscripcion, InscripcionPut, respuesta, i, alerta, alertas, errorGetAll)
		} else {
			ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
			return map[string]interface{}{"Response": alerta}
		}
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "", errInscripcion)
		return map[string]interface{}{"Response": alerta}
	}
}

// FUNCIONES QUE SE USAN EN GET EVALUACION ASPIRANTES

func solicitudTercerosGetEvApspirantes(Inscripcion map[string]interface{}, Terceros map[string]interface{}, respuestaAux *string, errorGetAll bool, alerta models.Alert, alertas []interface{}) interface{} {
	TerceroId := fmt.Sprintf("%v", Inscripcion["PersonaId"])
	errTerceros := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+TerceroId, &Terceros)
	if errTerceros == nil {
		if Terceros != nil && fmt.Sprintf("%v", Terceros) != "map[]" {
			*respuestaAux = *respuestaAux + "\"Aspirantes\": " + fmt.Sprintf("%q", Terceros["NombreCompleto"]) + "\n}"
			return nil
		} else {
			ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
			return map[string]interface{}{"Response": alerta}
		}
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "", errTerceros)
		return map[string]interface{}{"Response": alerta}
	}
}

func SolicitudInscripcionGetEvApspirantes(evaluacion map[string]interface{}, Inscripcion map[string]interface{}, Terceros map[string]interface{}, respuestaAux string, errorGetAll bool, alerta models.Alert, alertas []interface{}) interface{} {
	InscripcionId := fmt.Sprintf("%v", evaluacion["InscripcionId"])
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, &Inscripcion)
	if errInscripcion == nil {
		if Inscripcion != nil && fmt.Sprintf("%v", Inscripcion) != "map[]" {
			//GET a la tabla de terceros para obtener el nombre
			return solicitudTercerosGetEvApspirantes(Inscripcion, Terceros, &respuestaAux, errorGetAll, alerta, alertas)
		} else {
			ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
			return map[string]interface{}{"Response": alerta}
		}
	} else {
		ManejoError(&alerta, &alertas, &errorGetAll, "", errInscripcion)
		return map[string]interface{}{"Response": alerta}
	}
}

func IterarEvaluacion(Evaluacion map[string]interface{}, respuestaAux *string) {
	for k := range Evaluacion["areas"].([]interface{}) {
		for k1, aux := range Evaluacion["areas"].([]interface{})[k].(map[string]interface{}) {
			if k1 != "Ponderado" {
				if k1 == "Asistencia" {
					*respuestaAux = *respuestaAux + fmt.Sprintf("%q", k1) + ":" + fmt.Sprintf("%t", aux) + ",\n"
				} else {
					*respuestaAux = *respuestaAux + fmt.Sprintf("%q", k1) + ":" + fmt.Sprintf("%q", aux) + ",\n"
				}
			}
		}
	}
}

// FUNCIONES QUE SE USAN EN POST CRITERIO ICFES

func ManejoCriterioCriterioIcfes(criterioProyecto *[]map[string]interface{}, CriterioIcfes map[string]interface{}, requestBod string, criterioProyectos map[string]interface{}, i int, alertas []interface{}, alerta models.Alert, tipo int, criterio_existente *[]map[string]interface{}) {
	var Id_criterio_existente interface{}
	Id_criterio_existente = nil
	if tipo == 1 {
		fmt.Println("Existe criterio")
		Id_criterio_existente = (*criterio_existente)[0]["Id"]
	} else if tipo == 2 {
		fmt.Println("No Existe criterio")
	}
	*criterioProyecto = append(*criterioProyecto, map[string]interface{}{
		"Activo":               true,
		"PeriodoId":            CriterioIcfes["Periodo"].(map[string]interface{})["Id"],
		"PorcentajeEspecifico": requestBod,
		"PorcentajeGeneral":    CriterioIcfes["General"],
		"ProgramaAcademicoId":  criterioProyectos["Id"],
		"RequisitoId":          map[string]interface{}{"Id": 1},
	})

	if tipo == 1 {
		solicitudCriterioPutIcfes(Id_criterio_existente, *criterioProyecto, i, alerta, alertas)
	} else if tipo == 2 {
		solicitudCriterioPostIcfes(criterioProyecto, i, alertas, alerta)
	}
}

func solicitudCriterioPutIcfes(Id_criterio_existente interface{}, criterioProyecto []map[string]interface{}, i int, alerta models.Alert, alertas []interface{}) {
	var resultadoPutcriterio map[string]interface{}
	errPutCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/"+fmt.Sprintf("%.f", Id_criterio_existente.(float64)), "PUT", &resultadoPutcriterio, criterioProyecto[i])
	if resultadoPutcriterio["Type"] == "error" || errPutCriterio != nil || resultadoPutcriterio["Status"] == "404" || resultadoPutcriterio["Message"] != nil {
		ManejoErrorSinGetAll(&alerta, &alertas, fmt.Sprintf("%v", resultadoPutcriterio))
	} else {
		fmt.Println("Registro  PUT de criterios bien")
	}
}

func solicitudCriterioPostIcfes(criterioProyecto *[]map[string]interface{}, i int, alertas []interface{}, alerta models.Alert) {
	var resultadocriterio map[string]interface{}
	errPostCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico", "POST", &resultadocriterio, (*criterioProyecto)[i])
	if resultadocriterio["Type"] == "error" || errPostCriterio != nil || resultadocriterio["Status"] == "404" || resultadocriterio["Message"] != nil {
		ManejoErrorSinGetAll(&alerta, &alertas, fmt.Sprintf("%v", resultadocriterio))
	} else {
		fmt.Println("Registro de criterios bien")
	}
}

// FUNCIONES QUE SE USAN EN POST CUPO ADMISION

func manejoCriterioCuposAdmision(tipo int, cupos_existente []map[string]interface{}, CuposProyectos *[]map[string]interface{}, CuposAdmision map[string]interface{}, requestBod string, cupoProyectos map[string]interface{}, i int) interface{} {
	var Id_cupo_existente interface{}
	Id_cupo_existente = nil
	if tipo == 1 {
		fmt.Println("Existe cupos para el proyecto")
		Id_cupo_existente = cupos_existente[0]["Id"]
	} else if tipo == 2 {
		fmt.Println("No Existe cupo")
	}

	*CuposProyectos = append(*CuposProyectos, map[string]interface{}{
		"Activo":           true,
		"PeriodoId":        CuposAdmision["Periodo"].(map[string]interface{})["Id"],
		"CuposEspeciales":  requestBod,
		"CuposHabilitados": CuposAdmision["CuposAsignados"],
		"DependenciaId":    cupoProyectos["Id"],
		"CuposOpcionados":  CuposAdmision["CuposOpcionados"],
	})

	if tipo == 1 {
		return solicitudPutCuposAdmision(Id_cupo_existente, CuposProyectos, i)
	} else if tipo == 2 {
		return solicitudPostCuposAdmision(CuposProyectos, i)
	} else {
		return nil
	}
}

func solicitudPutCuposAdmision(Id_cupo_existente interface{}, CuposProyectos *[]map[string]interface{}, i int) interface{} {
	var resultadoPutcupo map[string]interface{}
	errPutCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/"+fmt.Sprintf("%.f", Id_cupo_existente.(float64)), "PUT", &resultadoPutcupo, (*CuposProyectos)[i])
	if resultadoPutcupo["Type"] == "error" || errPutCriterio != nil || resultadoPutcupo["Status"] == "404" || resultadoPutcupo["Message"] != nil {
		return map[string]interface{}{"Success": false, "Status": "400", "Message": resultadoPutcupo, "Data": nil}
	} else {
		fmt.Println("Registro  PUT de cupo bien")
		return nil
	}
}

func solicitudPostCuposAdmision(CuposProyectos *[]map[string]interface{}, i int) interface{} {
	var resultadocupopost map[string]interface{}
	errPostCupo := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia", "POST", &resultadocupopost, (*CuposProyectos)[i])
	if resultadocupopost["Type"] == "error" || errPostCupo != nil || resultadocupopost["Status"] == "404" || resultadocupopost["Message"] != nil {
		return map[string]interface{}{"Success": false, "Status": "400", "Message": errPostCupo, "Data": nil}
	} else {
		fmt.Println("Registro de cupo bien")
		return nil
	}
}

func manejoError404(cupos_existente []map[string]interface{}, errCupoExistente interface{}) interface{} {
	if cupos_existente[0]["Message"] == "Not found resource" {
		return map[string]interface{}{"Success": false, "Status": "400", "Message": cupos_existente[0]["Message"], "Data": nil}
	} else {
		return map[string]interface{}{"Success": false, "Status": "404", "Message": errCupoExistente, "Data": nil}
	}
}

func SolicituVerificacionCuposAdmisio(cupoProyectos map[string]interface{}, CuposAdmision map[string]interface{}, CuposProyectos []map[string]interface{}, requestBod string, i int) interface{} {
	var cupos_existente []map[string]interface{}
	errCupoExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/?query=DependenciaId:"+fmt.Sprintf("%.f", cupoProyectos["Id"].(float64))+",PeriodoId:"+fmt.Sprintf("%.f", CuposAdmision["Periodo"].(map[string]interface{})["Id"].(float64)), &cupos_existente)
	if errCupoExistente == nil && fmt.Sprintf("%v", cupos_existente[0]) != "map[]" {
		if cupos_existente[0]["Status"] != 404 {
			return manejoCriterioCuposAdmision(1, cupos_existente, &CuposProyectos, CuposAdmision, requestBod, cupoProyectos, i)
		} else {
			return manejoError404(cupos_existente, errCupoExistente)
		}
	} else {
		return manejoCriterioCuposAdmision(2, cupos_existente, &CuposProyectos, CuposAdmision, requestBod, cupoProyectos, i)
	}
}

// FUNCIONES QUE SE USAN VARIAS FUNCIONES

func ManejoError(alerta *models.Alert, alertas *[]interface{}, errorGetAll *bool, mensaje string, err ...error) {
	var msj string
	if len(err) > 0 && err[0] != nil {
		msj = mensaje + err[0].Error()
	} else {
		msj = mensaje
	}
	*errorGetAll = true
	*alertas = append(*alertas, msj)
	(*alerta).Body = *alertas
	(*alerta).Type = "error"
	(*alerta).Code = "400"
}

func ManejoErrorSinGetAll(alerta *models.Alert, alertas *[]interface{}, mensaje string, err ...error) {
	var msj string
	if len(err) > 0 && err[0] != nil {
		msj = mensaje + err[0].Error()
	} else {
		msj = mensaje
	}
	*alertas = append(*alertas, msj)
	(*alerta).Body = *alertas
	(*alerta).Type = "error"
	(*alerta).Code = "400"
}

func ManejoExito(alertas *[]interface{}, alerta *models.Alert, resultado map[string]interface{}) {
	*alertas = append(*alertas, resultado)
	(*alerta).Body = *alertas
	(*alerta).Code = "200"
	(*alerta).Type = "OK"
}
