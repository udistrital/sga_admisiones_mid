package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_admisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"golang.org/x/sync/errgroup"
)

// FUNCIONES QUE SE USAN EN PUT NOTA FINAL ASPIRANTES

func solicitudInscripcionPut(InscripcionId string, InscripcionPut map[string]interface{}, Inscripcion *[]map[string]interface{}, respuesta *[]map[string]interface{}, i int, errorGetAll bool) (APIResponseDTO requestresponse.APIResponse, err bool) {
	errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, "PUT", &InscripcionPut, (*Inscripcion)[0])
	if errInscripcionPut == nil {
		if InscripcionPut != nil && fmt.Sprintf("%v", InscripcionPut) != "map[]" {
			(*respuesta)[i] = InscripcionPut
			errorGetAll = true
			APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil)
			return APIResponseDTO, err
		} else {
			errorGetAll = true
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
			return APIResponseDTO, err
		}
	} else {
		errorGetAll = true
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errInscripcionPut)
		return APIResponseDTO, err
	}
}

func validarDetalleEvaluacionPut(DetalleEvaluacion *[]map[string]interface{}, NotaFinal float64, Inscripcion *[]map[string]interface{}, InscripcionId string, InscripcionPut map[string]interface{}, respuesta *[]map[string]interface{}, i int, errorGetAll bool) (APIResponseDTO requestresponse.APIResponse, err bool) {
	if *DetalleEvaluacion != nil && fmt.Sprintf("%v", (*DetalleEvaluacion)[0]) != "map[]" {
		NotaFinal = 0
		// Calculo de la nota Final con los criterios relacionados al proyecto
		for _, EvaluacionAux := range *DetalleEvaluacion {
			f, _ := strconv.ParseFloat(fmt.Sprintf("%v", EvaluacionAux["NotaRequisito"]), 64)
			NotaFinal = NotaFinal + f
		}
		NotaFinal = math.Round(NotaFinal*100) / 100
		(*Inscripcion)[0]["NotaFinal"] = NotaFinal

		//PUT a inscripción con la nota final calculada
		return solicitudInscripcionPut(InscripcionId, InscripcionPut, Inscripcion, respuesta, i, errorGetAll)
	} else {
		errorGetAll = true
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
		return APIResponseDTO, err
	}
}

func solicitudDetalleEvaluacionPut(InscripcionId string, ProgramaAcademicoId string, PeriodoId string, DetalleEvaluacion *[]map[string]interface{}, NotaFinal float64, Inscripcion *[]map[string]interface{}, InscripcionPut map[string]interface{}, respuesta *[]map[string]interface{}, i int, errorGetAll bool) (APIResponseDTO requestresponse.APIResponse, err bool) {
	errDetalleEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=InscripcionId:"+InscripcionId+",RequisitoProgramaAcademicoId__ProgramaAcademicoId:"+ProgramaAcademicoId+",RequisitoProgramaAcademicoId__PeriodoId:"+PeriodoId+"&limit=0", DetalleEvaluacion)
	if errDetalleEvaluacion == nil {
		return validarDetalleEvaluacionPut(DetalleEvaluacion, NotaFinal, Inscripcion, InscripcionId, InscripcionPut, respuesta, i, errorGetAll)
	} else {
		errorGetAll = true
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, errDetalleEvaluacion.Error())
		return APIResponseDTO, err
	}
}

func SolicitudIdPut(data []byte) (APIResponseDTO requestresponse.APIResponse) {

	var Evaluacion map[string]interface{}
	var Inscripcion = &[]map[string]interface{}{}
	var DetalleEvaluacion = &[]map[string]interface{}{}
	var NotaFinal float64
	var InscripcionPut map[string]interface{}
	var respuesta []map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var errorGetAll = false

	if err := json.Unmarshal(data, &Evaluacion); err == nil {
		IdPersona := Evaluacion["IdPersona"].([]interface{})
		PeriodoId := fmt.Sprintf("%v", Evaluacion["IdPeriodo"])
		ProgramaAcademicoId := fmt.Sprintf("%v", Evaluacion["IdPrograma"])
		respuesta = make([]map[string]interface{}, len(IdPersona))
		for i := 0; i < len(IdPersona); i++ {
			PersonaId := fmt.Sprintf("%v", IdPersona[i].(map[string]interface{})["Id"])

			//GET a Inscripción para obtener el ID
			errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=PersonaId:"+PersonaId+",PeriodoId:"+PeriodoId+",ProgramaAcademicoId:"+ProgramaAcademicoId, Inscripcion)
			if errInscripcion == nil {
				if Inscripcion != nil && fmt.Sprintf("%v", (*Inscripcion)[0]) != "map[]" {
					InscripcionId := fmt.Sprintf("%v", (*Inscripcion)[0]["Id"])

					//GET a detalle evaluacion
					APIResponseDTO, errorGetAll = solicitudDetalleEvaluacionPut(InscripcionId, ProgramaAcademicoId, PeriodoId, DetalleEvaluacion, NotaFinal, Inscripcion, InscripcionPut, &respuesta, i, errorGetAll)
				} else {
					errorGetAll = true
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
					return APIResponseDTO
				}
			} else {
				errorGetAll = true
				APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, errInscripcion.Error())
				return APIResponseDTO
			}
		}
		resultado["Response"] = respuesta
	} else {
		errorGetAll = true
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, err.Error())
		return APIResponseDTO
	}

	if !errorGetAll {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 200, resultado)
		return APIResponseDTO
	}
	return APIResponseDTO
}

// FUNCIONES QUE SE USAN EN GET EVALUACION ASPIRANTES

func solicitudTercerosGetEvApspirantes(Inscripcion *map[string]interface{}, Terceros *map[string]interface{}, respuestaAux *string, errorGetAll *bool) interface{} {
	TerceroId := fmt.Sprintf("%v", (*Inscripcion)["PersonaId"])
	errTerceros := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+TerceroId, Terceros)
	if errTerceros == nil {
		if *Terceros != nil && fmt.Sprintf("%v", *Terceros) != "map[]" {
			*respuestaAux = *respuestaAux + "\"Aspirantes\": " + fmt.Sprintf("%q", (*Terceros)["NombreCompleto"]) + "\n}"
			return nil
		} else {
			APIResponseDTO := requestresponse.APIResponseDTO(false, 404, nil, "No data fouund")
			return APIResponseDTO
		}
	} else {
		APIResponseDTO := requestresponse.APIResponseDTO(false, 400, nil, errTerceros.Error())
		return APIResponseDTO
	}
}

func SolicitudInscripcionGetEvApspirantes(evaluacion map[string]interface{}, Inscripcion *map[string]interface{}, Terceros *map[string]interface{}, respuestaAux *string, errorGetAll *bool) interface{} {
	InscripcionId := fmt.Sprintf("%v", evaluacion["InscripcionId"])
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, Inscripcion)
	if errInscripcion == nil {
		if *Inscripcion != nil && fmt.Sprintf("%v", *Inscripcion) != "map[]" {
			//GET a la tabla de terceros para obtener el nombre
			return solicitudTercerosGetEvApspirantes(Inscripcion, Terceros, respuestaAux, errorGetAll)
		} else {
			APIResponseDTO := requestresponse.APIResponseDTO(false, 404, nil, "No data fouund")
			return APIResponseDTO
		}
	} else {
		APIResponseDTO := requestresponse.APIResponseDTO(false, 400, nil, errInscripcion.Error())
		return APIResponseDTO
	}
}

func IterarEvaluacion(id_periodo string, id_programa string, id_requisito string) (APIResponseDTO requestresponse.APIResponse) {

	var DetalleEvaluacion []map[string]interface{}
	var DetalleEspecificoJSON []map[string]interface{}
	var Inscripcion map[string]interface{}
	var Terceros map[string]interface{}

	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var errorGetAll bool

	//GET a la tabla detalle_evaluacion
	errDetalleEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=RequisitoProgramaAcademicoId__RequisitoId__Id:"+id_requisito+",RequisitoProgramaAcademicoId__PeriodoId:"+id_periodo+",RequisitoProgramaAcademicoId__ProgramaAcademicoId:"+id_programa+"&sortby=InscripcionId&order=asc&limit=0", &DetalleEvaluacion)
	if errDetalleEvaluacion == nil {

		var formatoFecha = "2006-01-02 15:04:05.999999999 -0700 -0700"
		var InscripcionIdReciente, _ = DetalleEvaluacion[0]["InscripcionId"]
		var ids []float64
		var fechasMasReciente []time.Time
		ids = append(ids, DetalleEvaluacion[0]["InscripcionId"].(float64))

		for _, evaluacion := range DetalleEvaluacion {
			var InscripcionIdActual, _ = evaluacion["InscripcionId"]
			if InscripcionIdActual == InscripcionIdReciente {
				InscripcionIdReciente = InscripcionIdActual
			} else if InscripcionIdActual != InscripcionIdReciente {
				InscripcionIdReciente = InscripcionIdActual
				ids = append(ids, InscripcionIdReciente.(float64))
			}
		}

		for _, id := range ids {
			fechaReciente := time.Time{}
			for i, evaluacion := range DetalleEvaluacion {
				if id == DetalleEvaluacion[i]["InscripcionId"] {
					var fechaActual, _ = time.Parse(formatoFecha, evaluacion["FechaModificacion"].(string))
					if fechaActual.After(fechaReciente) {
						fechaReciente = fechaActual
					}
				}
			}
			fechasMasReciente = append(fechasMasReciente, fechaReciente)
		}

		if DetalleEvaluacion != nil && fmt.Sprintf("%v", DetalleEvaluacion[0]) != "map[]" {
			Respuesta := "[\n"

			for j, id := range ids {
				for i, evaluacion := range DetalleEvaluacion {
					evaluacionFecha, _ := time.Parse(formatoFecha, evaluacion["FechaModificacion"].(string))
					if id == evaluacion["InscripcionId"] && fechasMasReciente[j].Equal(evaluacionFecha) {
						respuestaAux := "{\n"
						var Evaluacion map[string]interface{}
						DetalleEspecifico := evaluacion["DetalleCalificacion"].(string)
						if err := json.Unmarshal([]byte(DetalleEspecifico), &Evaluacion); err == nil {
							for k := range Evaluacion["areas"].([]interface{}) {
								for k1, aux := range Evaluacion["areas"].([]interface{})[k].(map[string]interface{}) {
									if k1 != "Ponderado" {
										if k1 == "Asistencia" {
											respuestaAux = respuestaAux + fmt.Sprintf("%q", k1) + ":" + fmt.Sprintf("%t", aux) + ",\n"
										} else {
											respuestaAux = respuestaAux + fmt.Sprintf("%q", k1) + ":" + fmt.Sprintf("%q", aux) + ",\n"
										}
									}
								}
							}

							//GET a la tabla de inscripcion para saber el id del inscrito
							if resp := SolicitudInscripcionGetEvApspirantes(evaluacion, &Inscripcion, &Terceros, &respuestaAux, &errorGetAll); resp != nil {
								APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, resp)
							}

							if i+1 == len(DetalleEvaluacion) {
								Respuesta = Respuesta + respuestaAux + "\n]"
							} else {
								Respuesta = Respuesta + respuestaAux + ",\n"
							}
						}
					}

				}

			}

			if err := json.Unmarshal([]byte(Respuesta), &DetalleEspecificoJSON); err == nil {
				resultado["areas"] = DetalleEspecificoJSON
			}
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data fouund")
		}

	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data fouund")
	}

	if !errorGetAll {
		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado)
		return APIResponseDTO
	}
	return APIResponseDTO
}

// FUNCIONES QUE SE USAN EN POST EVALUACION ASPIRANTES

func validarAsistenciaPostEvaluacion(Asistencia *interface{}, PorcentajeEspJSON map[string]interface{}, k int, Ponderado *float64, DetalleCalificacion *string, aux2 interface{}, k2 string, ultimo bool) {
	if *Asistencia != nil {
		if *Asistencia == true {
			f, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{})["Porcentaje"]), 64) //Porcentaje del subcriterio
			j, _ := strconv.ParseFloat(fmt.Sprintf("%v", aux2), 64)                                                                                 //Nota subcriterio
			PonderadoAux := j * (f / 100)
			*Ponderado = *Ponderado + PonderadoAux
			if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
				*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
				ultimo = true
			} else {
				*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
			}
		} else {
			// Si el estudiante inscrito no asiste tendrá una calificación de 0
			*Ponderado = 0
			if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
				*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":\"0\", \"Ponderado\":\"0\"},\n"
				ultimo = true
			} else {
				*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":\"0\", \"Ponderado\":\"0\"},\n"
			}
		}
	} else {
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{})["Porcentaje"]), 64) //Porcentaje del subcriterio
		j, _ := strconv.ParseFloat(fmt.Sprintf("%v", aux2), 64)                                                                                 //Nota subcriterio
		PonderadoAux := j * (f / 100)
		*Ponderado = *Ponderado + PonderadoAux
		if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
			*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
			ultimo = true
		} else {
			*DetalleCalificacion = *DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
		}
	}
}

func calculoSubCriteriosPostEvaluacion(Asistencia interface{}, AspirantesData []interface{}, PorcentajeGeneral interface{}, Ponderado *float64, DetalleCalificacion *string, i int) {
	if Asistencia != nil {
		if Asistencia == true {
			f, _ := strconv.ParseFloat(fmt.Sprintf("%v", AspirantesData[i].(map[string]interface{})["Puntuacion"]), 64) //Puntaje del aspirante
			g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)                                        //Porcentaje del criterio
			*Ponderado = f * (g / 100)                                                                                  //100% del puntaje que obtuvo el aspirante
			*DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\":" + fmt.Sprintf("%q", AspirantesData[i].(map[string]interface{})["Puntuacion"]) + "}\n]\"n}"
		} else {
			// Si el estudiante inscrito no asiste tendrá una calificación de 0
			*Ponderado = 0
			*DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\": \"0\"}\n]\n}"
		}
	} else {
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", AspirantesData[i].(map[string]interface{})["Puntuacion"]), 64) //Puntaje del aspirante
		g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)                                        //Porcentaje del criterio
		*Ponderado = f * (g / 100)                                                                                  //100% del puntaje que obtuvo el aspirante
		*DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\":" + fmt.Sprintf("%q", AspirantesData[i].(map[string]interface{})["Puntuacion"]) + "}\n]\"n}"
	}
}

func solictiudDetalleEvaluacionPostEvaluacion(DetalleEvaluacion map[string]interface{}, respuesta *[]map[string]interface{}, i int, errorGetAll *bool, alertas *[]interface{}, alerta *models.Alert) interface{} {

	errDetalleEvaluacion := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion", "POST", &DetalleEvaluacion, (*respuesta)[i])
	if errDetalleEvaluacion == nil {
		if DetalleEvaluacion != nil && fmt.Sprintf("%v", DetalleEvaluacion) != "map[]" {
			//respuesta[i] = DetalleEvaluacion
			return nil
		} else {
			*errorGetAll = true
			ManejoError(alerta, alertas, errorGetAll, "No data found")
			return map[string]interface{}{"Response": *alerta}
		}
	} else {
		ManejoError(alerta, alertas, errorGetAll, "", errDetalleEvaluacion)
		return map[string]interface{}{"Response": *alerta}
	}
}

func validarPorcentajesPostEvaluacion(PorcentajeEspJSON map[string]interface{}, Ponderado *float64, DetalleCalificacion *string, Evaluacion map[string]interface{}, i int, Asistencia *interface{}, PorcentajeGeneral interface{}, AspirantesData []interface{}) {
	if PorcentajeEspJSON != nil && fmt.Sprintf("%v", PorcentajeEspJSON) != "map[]" {
		//Calculos para los criterios que cuentan con subcriterios)
		*Ponderado = 0
		*DetalleCalificacion = "{\n\"areas\":\n["
		ultimo := false

		for k := range PorcentajeEspJSON["areas"].([]interface{}) {
			for _, aux := range PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{}) {
				for k2, aux2 := range Evaluacion["Aspirantes"].([]interface{})[i].(map[string]interface{}) {
					if ultimo {
						break
					}
					if aux == k2 {
						//Si existe la columna de asistencia se hace la validación de la misma
						validarAsistenciaPostEvaluacion(Asistencia, PorcentajeEspJSON, k, Ponderado, DetalleCalificacion, aux2, k2, ultimo)
					}
				}
			}
		}
		g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)
		*Ponderado = *Ponderado * (g / 100)
		if *Asistencia == true && *Asistencia != nil {
			*DetalleCalificacion = *DetalleCalificacion + "{\"Asistencia\": true" + "}]\n}"
		} else {
			*DetalleCalificacion = *DetalleCalificacion + "{\"Asistencia\": false" + "}]\n}"
		}
	} else {
		//Calculos para los criterios que no tienen subcriterios
		//Si existe la columna de asistencia se hace la validación de la misma

		calculoSubCriteriosPostEvaluacion(*Asistencia, AspirantesData, PorcentajeGeneral, Ponderado, DetalleCalificacion, i)
	}
}

func solicitudInscripcionesPostEvaluacion(PersonaId interface{}, ProgramaAcademicoId interface{}, PeriodoId interface{}, Inscripciones *[]map[string]interface{}, PorcentajeEspJSON map[string]interface{}, Ponderado *float64, DetalleCalificacion *string, Evaluacion map[string]interface{}, i int, Asistencia *interface{}, PorcentajeGeneral interface{}, AspirantesData []interface{}, respuesta *[]map[string]interface{}, Requisito []map[string]interface{}, DetalleEvaluacion map[string]interface{}, errorGetAll *bool, alertas *[]interface{}, alerta *models.Alert) interface{} {

	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=PersonaId:"+fmt.Sprintf("%v", PersonaId)+",ProgramaAcademicoId:"+fmt.Sprintf("%v", ProgramaAcademicoId)+",PeriodoId:"+fmt.Sprintf("%v", PeriodoId), Inscripciones)

	if errInscripcion == nil {
		if *Inscripciones != nil && fmt.Sprintf("%v", (*Inscripciones)[0]) != "map[]" {
			validarPorcentajesPostEvaluacion(PorcentajeEspJSON, Ponderado, DetalleCalificacion, Evaluacion, i, Asistencia, PorcentajeGeneral, AspirantesData)
			// JSON para el post detalle evaluacion
			(*respuesta)[i] = map[string]interface{}{
				"InscripcionId":                (*Inscripciones)[0]["Id"],
				"RequisitoProgramaAcademicoId": Requisito[0],
				"Activo":                       true,
				"FechaCreacion":                time.Now(),
				"FechaModificacion":            time.Now(),
				"DetalleCalificacion":          *DetalleCalificacion,
				"NotaRequisito":                *Ponderado,
			}
			//Función POST a la tabla detalle_evaluación

			return solictiudDetalleEvaluacionPostEvaluacion(DetalleEvaluacion, respuesta, i, errorGetAll, alertas, alerta)
		} else {
			*errorGetAll = true
			ManejoError(alerta, alertas, errorGetAll, "No data found")
			return map[string]interface{}{"Response": *alerta}
		}
	} else {
		*errorGetAll = true
		ManejoError(alerta, alertas, errorGetAll, "", errInscripcion)
		return map[string]interface{}{"Response": *alerta}
	}
}

func SolicitudRequisitoPostEvaluacion(ProgramaAcademicoId interface{}, PeriodoId interface{}, Inscripciones *[]map[string]interface{}, Ponderado *float64, DetalleCalificacion *string, Evaluacion map[string]interface{}, AspirantesData []interface{}, respuesta *[]map[string]interface{}, Requisito []map[string]interface{}, DetalleEvaluacion map[string]interface{}, errorGetAll *bool, alertas *[]interface{}, alerta *models.Alert, CriterioId interface{}) interface{} {

	errRequisito := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico?query=ProgramaAcademicoId:"+fmt.Sprintf("%v", ProgramaAcademicoId)+",PeriodoId:"+fmt.Sprintf("%v", PeriodoId)+",RequisitoId:"+fmt.Sprintf("%v", CriterioId), &Requisito)
	if errRequisito == nil {
		if Requisito != nil && fmt.Sprintf("%v", Requisito[0]) != "map[]" {
			//Se guarda JSON con los porcentajes específicos
			var PorcentajeEspJSON map[string]interface{}
			PorcentajeGeneral := Requisito[0]["PorcentajeGeneral"]
			PorcentajeEspecifico := Requisito[0]["PorcentajeEspecifico"].(string)
			if err := json.Unmarshal([]byte(PorcentajeEspecifico), &PorcentajeEspJSON); err == nil {
				for i := 0; i < len(AspirantesData); i++ {
					PersonaId := AspirantesData[i].(map[string]interface{})["Id"]
					Asistencia := AspirantesData[i].(map[string]interface{})["Asistencia"]
					if Asistencia == "" {
						Asistencia = nil
					}

					//GET para obtener el numero de la inscripcion de la persona
					if errSolicitud := solicitudInscripcionesPostEvaluacion(PersonaId, ProgramaAcademicoId, PeriodoId, Inscripciones, PorcentajeEspJSON, Ponderado, DetalleCalificacion, Evaluacion, i, &Asistencia, PorcentajeGeneral, AspirantesData, respuesta, Requisito, DetalleEvaluacion, errorGetAll, alertas, alerta); errSolicitud != nil {
						return errSolicitud
					}
				}
			}
		} else {
			*errorGetAll = true
			ManejoError(alerta, alertas, errorGetAll, "No data found")
			return map[string]interface{}{"Response": *alerta}
		}
	} else {
		*errorGetAll = true
		ManejoError(alerta, alertas, errorGetAll, "", errRequisito)
		return map[string]interface{}{"Response": *alerta}
	}
	return nil
}

// FUNCIONES QUE SE USAN EN POST CRITERIO ICFES

func ManejoCriterioCriterioIcfes(criterioProyecto *[]map[string]interface{}, CriterioIcfes map[string]interface{}, requestBod string, criterioProyectos map[string]interface{}, i int, alertas *[]interface{}, alerta *models.Alert, tipo int, criterio_existente *[]map[string]interface{}) {
	var Id_criterio_existente interface{}
	Id_criterio_existente = nil
	if tipo == 1 {

		Id_criterio_existente = (*criterio_existente)[0]["Id"]
	} else if tipo == 2 {

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

func solicitudCriterioPutIcfes(Id_criterio_existente interface{}, criterioProyecto []map[string]interface{}, i int, alerta *models.Alert, alertas *[]interface{}) {
	var resultadoPutcriterio map[string]interface{}
	errPutCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/"+fmt.Sprintf("%.f", Id_criterio_existente.(float64)), "PUT", &resultadoPutcriterio, criterioProyecto[i])
	if resultadoPutcriterio["Type"] == "error" || errPutCriterio != nil || resultadoPutcriterio["Status"] == "404" || resultadoPutcriterio["Message"] != nil {
		ManejoErrorSinGetAll(alerta, alertas, fmt.Sprintf("%v", resultadoPutcriterio))
	} else {

	}
}

func solicitudCriterioPostIcfes(criterioProyecto *[]map[string]interface{}, i int, alertas *[]interface{}, alerta *models.Alert) {
	var resultadocriterio map[string]interface{}
	errPostCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico", "POST", &resultadocriterio, (*criterioProyecto)[i])
	if resultadocriterio["Type"] == "error" || errPostCriterio != nil || resultadocriterio["Status"] == "404" || resultadocriterio["Message"] != nil {
		ManejoErrorSinGetAll(alerta, alertas, fmt.Sprintf("%v", resultadocriterio))
	} else {

	}
}

// FUNCIONES QUE SE USAN EN GET PUNTAJE TOTAL BY PERIODO BY PROYECTO

func peticionResultadoDocumentoGetPuntaje(resultado_puntaje *[]map[string]interface{}, resultado_persona map[string]interface{}, i int, id_persona float64) (interface{}, interface{}, bool) {
	(*resultado_puntaje)[i]["NombreAspirante"] = resultado_persona["NombreCompleto"]
	var resultado_documento []map[string]interface{}
	errGetDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion/?query=TerceroId.Id:"+fmt.Sprintf("%v", id_persona), &resultado_documento)
	if errGetDocumento == nil && fmt.Sprintf("%v", resultado_documento[0]) != "map[]" {
		if resultado_documento[0]["Status"] != 404 {
			(*resultado_puntaje)[i]["TipoDocumento"] = resultado_documento[0]["TipoDocumentoId"].(map[string]interface{})["CodigoAbreviacion"]
			(*resultado_puntaje)[i]["NumeroDocumento"] = resultado_documento[0]["Numero"]
			return nil, nil, true
		} else {
			if resultado_documento[0]["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_documento[0])
				return errGetDocumento, nil, false
			}
		}
	} else {
		logs.Error(resultado_documento[0])
		return errGetDocumento, nil, false
	}
}

func peticionResultadoPersonaGetPuntaje(resultado_inscripcion map[string]interface{}, resultado_puntaje *[]map[string]interface{}, i int) (interface{}, interface{}, bool) {
	id_persona := (resultado_inscripcion["PersonaId"]).(float64)

	var resultado_persona map[string]interface{}
	errGetPersona := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+fmt.Sprintf("%v", id_persona), &resultado_persona)
	if errGetPersona == nil && fmt.Sprintf("%v", resultado_persona) != "map[]" {
		if resultado_persona["Status"] != 404 {
			infoSystem, infoJson, exito := peticionResultadoDocumentoGetPuntaje(resultado_puntaje, resultado_persona, i, id_persona)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultado_persona["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_persona)
				return errGetPersona, nil, false
			}
		}
	} else {
		logs.Error(resultado_persona)
		return errGetPersona, nil, false
	}
}

func PeticionResultadoInscripcionGetPuntaje(resultado_tem map[string]interface{}, resultado_puntaje *[]map[string]interface{}, i int) (interface{}, interface{}, bool) {
	id_inscripcion := (resultado_tem["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]).(float64)

	var resultado_inscripcion map[string]interface{}
	errGetInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", id_inscripcion), &resultado_inscripcion)
	if errGetInscripcion == nil && fmt.Sprintf("%v", resultado_inscripcion) != "map[]" {
		if resultado_inscripcion["Status"] != 404 {
			infoSystem, infoJson, exito := peticionResultadoPersonaGetPuntaje(resultado_inscripcion, resultado_puntaje, i)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultado_inscripcion["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_inscripcion)
				return errGetInscripcion, nil, false
			}
		}
	} else {
		logs.Error(resultado_inscripcion)
		return errGetInscripcion, nil, false
	}
}

// FUNCIONES QUE SE USAN EN POST CUPO ADMISION

func manejoCriterioCuposAdmision(tipo int, cupos_existente []map[string]interface{}, CuposProyectos *[]map[string]interface{}, CuposAdmision map[string]interface{}, requestBod string, cupoProyectos map[string]interface{}, i int) interface{} {
	var Id_cupo_existente interface{}
	Id_cupo_existente = nil
	if tipo == 1 {

		Id_cupo_existente = cupos_existente[0]["Id"]
	} else if tipo == 2 {

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

		return nil
	}
}

func solicitudPostCuposAdmision(CuposProyectos *[]map[string]interface{}, i int) interface{} {
	var resultadocupopost map[string]interface{}
	errPostCupo := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia", "POST", &resultadocupopost, (*CuposProyectos)[i])
	if resultadocupopost["Type"] == "error" || errPostCupo != nil || resultadocupopost["Status"] == "404" || resultadocupopost["Message"] != nil {
		return map[string]interface{}{"Success": false, "Status": "400", "Message": errPostCupo, "Data": nil}
	} else {

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

func SolicituVerificacionCuposAdmision(cupoProyectos map[string]interface{}, CuposAdmision map[string]interface{}, CuposProyectos *[]map[string]interface{}, requestBod string, i int) interface{} {
	var cupos_existente []map[string]interface{}
	errCupoExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/?query=DependenciaId:"+fmt.Sprintf("%.f", cupoProyectos["Id"].(float64))+",PeriodoId:"+fmt.Sprintf("%.f", CuposAdmision["Periodo"].(map[string]interface{})["Id"].(float64)), &cupos_existente)
	if errCupoExistente == nil && fmt.Sprintf("%v", cupos_existente[0]) != "map[]" {
		if cupos_existente[0]["Status"] != 404 {
			return manejoCriterioCuposAdmision(1, cupos_existente, CuposProyectos, CuposAdmision, requestBod, cupoProyectos, i)
		} else {
			return manejoError404(cupos_existente, errCupoExistente)
		}
	} else {
		return manejoCriterioCuposAdmision(2, cupos_existente, CuposProyectos, CuposAdmision, requestBod, cupoProyectos, i)
	}
}

// FUNCIONES QUE SE USAN EN CAMBIO ESTADO ASPIRANTE BY PERIODO BY PROYECTO

func peticionEliminarResultadoCambioEstado(estadotemp map[string]interface{}, errInscripcionPut interface{}) {
	var resultado2 map[string]interface{}
	request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"/inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), "DELETE", &resultado2, nil)
	logs.Error(errInscripcionPut)
}

func peticionInscripcionPutCambioEstado(estadotemp map[string]interface{}, resultadoaspiranteinscripcion map[string]interface{}, mensaje string) (interface{}, bool) {
	var inscripcionPut map[string]interface{}
	errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%.f", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"].(float64)), "PUT", &inscripcionPut, resultadoaspiranteinscripcion)
	if errInscripcionPut == nil && fmt.Sprintf("%v", inscripcionPut) != "map[]" && inscripcionPut["Id"] != nil {
		if inscripcionPut["Status"] != 400 {

			return nil, true
		} else {
			peticionEliminarResultadoCambioEstado(estadotemp, errInscripcionPut)
			return inscripcionPut, false
		}
	} else {
		logs.Error(errInscripcionPut)
		return inscripcionPut, false
	}
}

func peticionResultadoAspitanteCambioEstado(estadotemp map[string]interface{}, mensaje string, id int) (interface{}, interface{}, bool) {
	var resultadoaspiranteinscripcion map[string]interface{}
	errinscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), &resultadoaspiranteinscripcion)
	if errinscripcion == nil && fmt.Sprintf("%v", resultadoaspiranteinscripcion) != "map[]" {
		if resultadoaspiranteinscripcion["Status"] != 404 {
			resultadoaspiranteinscripcion["EstadoInscripcionId"] = map[string]interface{}{"Id": id}

			data, exito := peticionInscripcionPutCambioEstado(estadotemp, resultadoaspiranteinscripcion, mensaje)

			if !exito {
				return data, nil, false
			} else {
				return nil, nil, true
			}
		} else {
			if resultadoaspiranteinscripcion["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultadoaspiranteinscripcion)
				return errinscripcion, nil, false
			}
		}
	} else {
		logs.Error(resultadoaspiranteinscripcion)
		return errinscripcion, nil, false
	}
}

func iteracionAspirantesCambioEstado(resultadoaspirantenota []map[string]interface{}, CuposHabilitados int64, CuposOpcionados int64) (interface{}, interface{}, bool) {
	for e, estadotemp := range resultadoaspirantenota {
		if e < (int(CuposHabilitados)) {
			infoSystem, infoJson, exito := peticionResultadoAspitanteCambioEstado(estadotemp, "Put correcto Admitido", 2)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			}
		}
		if e >= int(CuposHabilitados) && e < (int(CuposHabilitados)+int(CuposOpcionados)) {
			infoSystem, infoJson, exito := peticionResultadoAspitanteCambioEstado(estadotemp, "Put correcto OPCIONADO", 3)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			}
		}
		if e >= (int(CuposHabilitados) + int(CuposOpcionados)) {
			infoSystem, infoJson, exito := peticionResultadoAspitanteCambioEstado(estadotemp, "Put correcto NO ADMITIDO", 4)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			}
		}
	}
	return nil, nil, true
}

func peticionAspiranteNotaCambioEstado(EstadoProyectos map[string]interface{}, Id_periodo interface{}, CuposHabilitados int64, CuposOpcionados int64) (interface{}, interface{}, bool) {
	var resultadoaspirantenota []map[string]interface{}
	errconsulta := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion/?query=RequisitoProgramaAcademicoId.ProgramaAcademicoId:"+fmt.Sprintf("%v", EstadoProyectos["Id"])+",RequisitoProgramaAcademicoId.PeriodoId:"+fmt.Sprintf("%v", Id_periodo)+"&limit=0&sortby=EvaluacionInscripcionId__NotaFinal&order=desc", &resultadoaspirantenota)
	if errconsulta == nil && fmt.Sprintf("%v", resultadoaspirantenota[0]) != "map[]" {
		if resultadoaspirantenota[0]["Status"] != 404 {
			infoSystem, infoJson, exito := iteracionAspirantesCambioEstado(resultadoaspirantenota, CuposHabilitados, CuposOpcionados)

			if !exito {
				if infoSystem != nil {
					//c.Data["system"] = infoSystem
					//c.Abort("404")
					return infoSystem, nil, false
				} else {
					//c.Data["json"] = infoJson
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultadoaspirantenota[0]["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultadoaspirantenota)
				return errconsulta, nil, false
			}
		}
	} else {
		logs.Error(resultadoaspirantenota)
		return errconsulta, nil, false
	}
}

func PeticionCuposCambioEstado(EstadoProyectos map[string]interface{}, Id_periodo interface{}) (interface{}, interface{}, bool) {
	var resultadocupo []map[string]interface{}
	errCupo := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/?query=DependenciaId:"+fmt.Sprintf("%v", EstadoProyectos["Id"])+",PeriodoId:"+fmt.Sprintf("%v", Id_periodo), &resultadocupo)

	if errCupo == nil && fmt.Sprintf("%v", resultadocupo[0]) != "map[]" {
		if resultadocupo[0]["Status"] != 404 {
			CuposHabilitados, _ := strconv.ParseInt(fmt.Sprintf("%v", resultadocupo[0]["CuposHabilitados"]), 10, 64)
			CuposOpcionados, _ := strconv.ParseInt(fmt.Sprintf("%v", resultadocupo[0]["CuposOpcionados"]), 10, 64)
			// consulta id inscripcion y nota final para cada proyecto con periodo, organiza el array de forma de descendente por el campo nota final para organizar del mayor puntaje al menor
			infoSystem, infoJson, exito := peticionAspiranteNotaCambioEstado(EstadoProyectos, Id_periodo, CuposHabilitados, CuposOpcionados)

			if !exito {
				if infoSystem != nil {
					//c.Data["system"] = infoSystem
					//c.Abort("404")
					return infoSystem, nil, false
				} else {
					//c.Data["json"] = infoJson
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultadocupo[0]["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultadocupo)
				return errCupo, nil, false
			}
		}
	} else {
		logs.Error(resultadocupo)
		return errCupo, nil, false
	}
}

// FUNCIONES QUE SE USAN EN GET ASPIRANTES BY PERIODO BY PROYECTO

func peticionResultadoDocGetAspirante(id_persona float64, resultado_aspirante *[]map[string]interface{}, i int) (interface{}, interface{}, bool) {
	var resultado_documento []map[string]interface{}
	errGetDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion/?query=TerceroId.Id:"+fmt.Sprintf("%v", id_persona), &resultado_documento)
	if errGetDocumento == nil && fmt.Sprintf("%v", resultado_documento[0]) != "map[]" {
		if resultado_documento[0]["Status"] != 404 {
			(*resultado_aspirante)[i]["TipoDocumento"] = resultado_documento[0]["TipoDocumentoId"].(map[string]interface{})["CodigoAbreviacion"]
			(*resultado_aspirante)[i]["NumeroDocumento"] = resultado_documento[0]["Numero"]
			return nil, nil, true
		} else {
			if resultado_documento[0]["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_documento[0])
				return errGetDocumento, nil, false
			}
		}
	} else {
		logs.Error(resultado_documento[0])
		return errGetDocumento, nil, false
	}
}

func peticionPersonaGetAspirante(resultado_tem map[string]interface{}, resultado_aspirante *[]map[string]interface{}, i int) (interface{}, interface{}, bool) {
	id_persona := (resultado_tem["PersonaId"]).(float64)
	var resultado_persona map[string]interface{}
	errGetPersona := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+fmt.Sprintf("%v", id_persona), &resultado_persona)
	if errGetPersona == nil && fmt.Sprintf("%v", resultado_persona) != "map[]" {
		if resultado_persona["Status"] != 404 {
			(*resultado_aspirante)[i]["NombreAspirante"] = resultado_persona["NombreCompleto"]

			infoSystem, infoJson, exito := peticionResultadoDocGetAspirante(id_persona, resultado_aspirante, i)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultado_persona["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_persona)
				return errGetPersona, nil, false
			}
		}
	} else {
		logs.Error(resultado_persona)
		return errGetPersona, nil, false
	}
}

func PeticionNotaGetAspirante(resultado_tem map[string]interface{}, resultado_aspirante *[]map[string]interface{}, i int) (interface{}, interface{}, bool) {
	id_inscripcion := (resultado_tem["Id"]).(float64)
	var resultado_nota []map[string]interface{}
	errGetNota := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"evaluacion_inscripcion/?query=InscripcionId:"+fmt.Sprintf("%v", id_inscripcion), &resultado_nota)
	if errGetNota == nil && fmt.Sprintf("%v", resultado_nota[0]) != "map[]" {
		if resultado_nota[0]["Status"] != 404 {
			(*resultado_aspirante)[i]["NotaFinal"] = resultado_nota[0]["NotaFinal"]

			infoSystem, infoJson, exito := peticionPersonaGetAspirante(resultado_tem, resultado_aspirante, i)

			if !exito {
				if infoSystem != nil {
					return infoSystem, nil, false
				} else {
					return nil, infoJson, false
				}
			} else {
				return nil, nil, true
			}
		} else {
			if resultado_nota[0]["Message"] == "Not found resource" {
				return nil, nil, false
			} else {
				logs.Error(resultado_nota)
				return errGetNota, nil, false
			}
		}
	} else {
		logs.Error(resultado_nota)
		return errGetNota, nil, false
	}
}

// FUNCIONES QUE SE USAN EN GET LISTA ASPIRANTES POR

func solicitudDatoIdentifGetLista(inscrip map[string]interface{}, datoIdentTercero *map[string]interface{}) {
	var datoIdentif []map[string]interface{}
	errDatoIdentif := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip["PersonaId"]), &datoIdentif)
	if errDatoIdentif == nil && fmt.Sprintf("%v", datoIdentif) != "[map[]]" {
		(*datoIdentTercero)["nombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["NombreCompleto"]
		(*datoIdentTercero)["numero"] = datoIdentif[0]["Numero"]
		(*datoIdentTercero)["correo"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["UsuarioWSO2"]
	} else {
		var datoIdentif_2intento []map[string]interface{}
		errDatoIdentif_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip["PersonaId"]), &datoIdentif_2intento)
		if errDatoIdentif_2intento == nil && fmt.Sprintf("%v", datoIdentif_2intento) != "[map[]]" {
			(*datoIdentTercero)["nombre"] = datoIdentif_2intento[0]["NombreCompleto"]
			(*datoIdentTercero)["numero"] = ""
			(*datoIdentTercero)["correo"] = datoIdentif_2intento[0]["UsuarioWSO2"]
		}
	}
}

func solicitudEnfasisGetLista(inscrip map[string]interface{}, enfasis *map[string]interface{}) {
	errEnfasis := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("enfasis/%v", inscrip["EnfasisId"]), enfasis)
	if errEnfasis != nil || (*enfasis)["Status"] == "404" {
		*enfasis = map[string]interface{}{
			"Nombre": "Por definir",
		}
	}
}

func solicitudTelefonoGetLista(inscrip map[string]interface{}, idTelefono string, telefonoPrincipal *string) {
	var telefono []map[string]interface{}
	errTelefono := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId.Id:%v,InfoComplementariaId.Id:%v&sortby=Id&order=desc&fields=Dato&limit=1", inscrip["PersonaId"], idTelefono), &telefono)
	if errTelefono == nil && fmt.Sprintf("%v", telefono) != "[map[]]" {
		var telefonos map[string]interface{}
		if err := json.Unmarshal([]byte(telefono[0]["Dato"].(string)), &telefonos); err == nil {
			*telefonoPrincipal = fmt.Sprintf("%.f", telefonos["principal"])
		}
	}
}

func solicitudReciboGetLista(ReciboInscripcion string, recibo *map[string]interface{}, Estado *string) {
	errRecibo := request.GetJsonWSO2("http://"+beego.AppConfig.String("ConsultarReciboJbpmService")+"consulta_recibo/"+ReciboInscripcion, recibo)
	if errRecibo == nil {
		if *recibo != nil && fmt.Sprintf("%v", *recibo) != "map[reciboCollection:map[]]" && fmt.Sprintf("%v", *recibo) != "map[]" {
			//Fecha límite de pago extraordinario
			FechaLimite := (*recibo)["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["fecha_extraordinario"].(string)
			EstadoRecibo := (*recibo)["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["estado"].(string)
			PagoRecibo := (*recibo)["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["pago"].(string)
			//Verificación si el recibo de pago se encuentra activo y pago
			if EstadoRecibo == "A" && PagoRecibo == "S" {
				*Estado = "Pago"
			} else {
				//Verifica si el recibo está vencido o no
				ATiempo, err := models.VerificarFechaLimite(FechaLimite)
				if err == nil {
					if ATiempo {
						*Estado = "Pendiente pago"
					} else {
						*Estado = "Vencido"
					}
				} else {
					*Estado = "Vencido"
				}
			}
		}
	}
}

func caso1Inscripcion1GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	var inscripcion1 []map[string]interface{}
	errInscripcion1 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:5,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", id_proyecto, id_periodo), &inscripcion1)
	if errInscripcion1 == nil && fmt.Sprintf("%v", inscripcion1) != "[map[]]" {
		for _, inscrip1 := range inscripcion1 {
			var datoIdentif1 []map[string]interface{}
			errDatoIdentif1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip1["PersonaId"]), &datoIdentif1)
			if errDatoIdentif1 == nil && fmt.Sprintf("%v", datoIdentif1) != "[map[]]" {
				*listado = append(*listado, map[string]interface{}{
					"Credencial":     inscrip1["Id"],
					"Identificacion": datoIdentif1[0]["Numero"],
					"Nombre":         datoIdentif1[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
					"Estado":         inscrip1["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				})
			} else {
				var datoIdentif1_2intento []map[string]interface{}
				errDatoIdentif1_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip1["PersonaId"]), &datoIdentif1_2intento)
				if errDatoIdentif1_2intento == nil && fmt.Sprintf("%v", datoIdentif1_2intento) != "[map[]]" {
					*listado = append(*listado, map[string]interface{}{
						"Credencial":     inscrip1["Id"],
						"Identificacion": "",
						"Nombre":         datoIdentif1_2intento[0]["NombreCompleto"],
						"Estado":         inscrip1["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
					})
				}
			}
		}
	}
}

func caso1Inscripcion2GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	var inscripcion2 []map[string]interface{}
	errInscripcion2 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:2,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", id_proyecto, id_periodo), &inscripcion2)
	if errInscripcion2 == nil && fmt.Sprintf("%v", inscripcion2) != "[map[]]" {
		for _, inscrip2 := range inscripcion2 {
			var datoIdentif2 []map[string]interface{}
			errDatoIdentif2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip2["PersonaId"]), &datoIdentif2)
			if errDatoIdentif2 == nil && fmt.Sprintf("%v", datoIdentif2) != "[map[]]" {
				*listado = append(*listado, map[string]interface{}{
					"Credencial":     inscrip2["Id"],
					"Identificacion": datoIdentif2[0]["Numero"],
					"Nombre":         datoIdentif2[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
					"Estado":         inscrip2["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				})
			} else {
				var datoIdentif2_2intento []map[string]interface{}
				errDatoIdentif2_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip2["PersonaId"]), &datoIdentif2_2intento)
				if errDatoIdentif2_2intento == nil && fmt.Sprintf("%v", datoIdentif2_2intento) != "[map[]]" {
					*listado = append(*listado, map[string]interface{}{
						"Credencial":     inscrip2["Id"],
						"Identificacion": "",
						"Nombre":         datoIdentif2_2intento[0]["NombreCompleto"],
						"Estado":         inscrip2["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
					})
				}
			}
		}
	}
}

func caso1Inscripcion3GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	var inscripcion3 []map[string]interface{}
	errInscripcion3 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:6,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", id_proyecto, id_periodo), &inscripcion3)
	if errInscripcion3 == nil && fmt.Sprintf("%v", inscripcion3) != "[map[]]" {
		for _, inscrip3 := range inscripcion3 {
			var datoIdentif3 []map[string]interface{}
			errDatoIdentif3 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip3["PersonaId"]), &datoIdentif3)
			if errDatoIdentif3 == nil && fmt.Sprintf("%v", datoIdentif3) != "[map[]]" {
				*listado = append(*listado, map[string]interface{}{
					"Credencial":     inscrip3["Id"],
					"Identificacion": datoIdentif3[0]["Numero"],
					"Nombre":         datoIdentif3[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
					"Estado":         inscrip3["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				})
			} else {
				var datoIdentif3_2intento []map[string]interface{}
				errDatoIdentif3_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip3["PersonaId"]), &datoIdentif3_2intento)
				if errDatoIdentif3_2intento == nil && fmt.Sprintf("%v", datoIdentif3_2intento) != "[map[]]" {
					*listado = append(*listado, map[string]interface{}{
						"Credencial":     inscrip3["Id"],
						"Identificacion": "",
						"Nombre":         datoIdentif3_2intento[0]["NombreCompleto"],
						"Estado":         inscrip3["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
					})
				}
			}
		}
	}
}

func caso2Inscripcion1GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	var inscripcion1 []map[string]interface{}
	errInscripcion1 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:5,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", id_proyecto, id_periodo), &inscripcion1)
	if errInscripcion1 == nil && fmt.Sprintf("%v", inscripcion1) != "[map[]]" {
		for _, inscrip1 := range inscripcion1 {
			var datoIdentif1 []map[string]interface{}
			errDatoIdentif1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip1["PersonaId"]), &datoIdentif1)
			if errDatoIdentif1 == nil && fmt.Sprintf("%v", datoIdentif1) != "[map[]]" {
				*listado = append(*listado, map[string]interface{}{
					"Id":         inscrip1["PersonaId"],
					"Aspirantes": datoIdentif1[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
				})
			} else {
				var datoIdentif1_2intento []map[string]interface{}
				errDatoIdentif1_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip1["PersonaId"]), &datoIdentif1_2intento)
				if errDatoIdentif1_2intento == nil && fmt.Sprintf("%v", datoIdentif1_2intento) != "[map[]]" {
					*listado = append(*listado, map[string]interface{}{
						"Id":         inscrip1["PersonaId"],
						"Aspirantes": datoIdentif1_2intento[0]["NombreCompleto"],
					})
				}
			}
		}
	}
}

func caso2Inscripcion2GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	var inscripcion2 []map[string]interface{}
	errInscripcion2 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:2,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", id_proyecto, id_periodo), &inscripcion2)
	if errInscripcion2 == nil && fmt.Sprintf("%v", inscripcion2) != "[map[]]" {
		for _, inscrip2 := range inscripcion2 {
			var datoIdentif2 []map[string]interface{}
			errDatoIdentif2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip2["PersonaId"]), &datoIdentif2)
			if errDatoIdentif2 == nil && fmt.Sprintf("%v", datoIdentif2) != "[map[]]" {
				*listado = append(*listado, map[string]interface{}{
					"Id":         inscrip2["PersonaId"],
					"Aspirantes": datoIdentif2[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
				})
			} else {
				var datoIdentif2_2intento []map[string]interface{}
				errDatoIdentif2_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip2["PersonaId"]), &datoIdentif2_2intento)
				if errDatoIdentif2_2intento == nil && fmt.Sprintf("%v", datoIdentif2_2intento) != "[map[]]" {
					*listado = append(*listado, map[string]interface{}{
						"Id":         inscrip2["PersonaId"],
						"Aspirantes": datoIdentif2_2intento[0]["NombreCompleto"],
					})
				}
			}
		}
	}
}

func caso3GetLista(id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	if idTelefono, ok := models.IdInfoCompTercero("10", "TELEFONO"); ok {
		var inscripcion []map[string]interface{}
		errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=NotaFinal&order=desc&limit=0", id_proyecto, id_periodo), &inscripcion)
		if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
			for _, inscrip := range inscripcion {

				datoIdentTercero := map[string]interface{}{
					"nombre": "",
					"numero": "",
					"correo": "",
				}

				solicitudDatoIdentifGetLista(inscrip, &datoIdentTercero)

				var enfasis map[string]interface{}
				solicitudEnfasisGetLista(inscrip, &enfasis)

				var telefonoPrincipal string = ""
				solicitudTelefonoGetLista(inscrip, idTelefono, &telefonoPrincipal)

				ReciboInscripcion := inscrip["ReciboInscripcion"].(string)
				var recibo map[string]interface{}
				var Estado string
				if ReciboInscripcion != "0/<nil>" {
					solicitudReciboGetLista(ReciboInscripcion, &recibo, &Estado)
				}

				*listado = append(*listado, map[string]interface{}{
					"Inscripcion":         inscrip,
					"NumeroDocumento":     datoIdentTercero["numero"],
					"NombreAspirante":     datoIdentTercero["nombre"],
					"Telefono":            telefonoPrincipal,
					"Email":               datoIdentTercero["correo"],
					"NotaFinal":           inscrip["NotaFinal"],
					"TipoInscripcionId":   inscrip["TipoInscripcionId"],
					"TipoInscripcion":     inscrip["TipoInscripcionId"].(map[string]interface{})["Nombre"],
					"EstadoInscripcionId": inscrip["EstadoInscripcionId"],
					"EstadoRecibo":        Estado,
					"EnfasisId":           enfasis,
					"Enfasis":             enfasis["Nombre"],
				})

			}
		}
	}
}

func ManejoCasosGetLista(tipo_lista int64, id_periodo int64, id_proyecto int64, listado *[]map[string]interface{}) {
	switch tipo_lista {
	case 1:
		caso1Inscripcion1GetLista(id_periodo, id_proyecto, listado)

		caso1Inscripcion2GetLista(id_periodo, id_proyecto, listado)

		caso1Inscripcion3GetLista(id_periodo, id_proyecto, listado)
	case 2:
		caso2Inscripcion1GetLista(id_periodo, id_proyecto, listado)

		caso2Inscripcion2GetLista(id_periodo, id_proyecto, listado)
	case 3:
		caso3GetLista(id_periodo, id_proyecto, listado)
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

func RegistratEvaluaciones(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var Evaluacion map[string]interface{}
	var Inscripciones []map[string]interface{}
	var Requisito []map[string]interface{}
	var DetalleCalificacion string
	var Ponderado float64
	var respuesta []map[string]interface{}
	var DetalleEvaluacion map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var alerta models.Alert
	var errorGetAll bool
	alertas := append([]interface{}{"Response:"})
	//Calificacion = append([]interface{}{"areas"})

	if err := json.Unmarshal(data, &Evaluacion); err == nil {
		AspirantesData := Evaluacion["Aspirantes"].([]interface{})
		ProgramaAcademicoId := Evaluacion["ProgramaId"]
		PeriodoId := Evaluacion["PeriodoId"]
		CriterioId := Evaluacion["CriterioId"]
		respuesta = make([]map[string]interface{}, len(AspirantesData))
		//GET para obtener el porcentaje general, especifico (si lo hay)

		if resp := SolicitudRequisitoPostEvaluacion(ProgramaAcademicoId, PeriodoId, &Inscripciones, &Ponderado, &DetalleCalificacion, Evaluacion, AspirantesData, &respuesta, Requisito, DetalleEvaluacion, &errorGetAll, &alertas, &alerta, CriterioId); resp != nil {
			errorGetAll = true
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, resp)
		}
		resultado["Evaluacion"] = respuesta
	} else {
		errorGetAll = true
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	}

	if !errorGetAll {
		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado)
		return APIResponseDTO
	}
	return APIResponseDTO
}

func CriteriosIcfesPost(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var CriterioIcfes map[string]interface{}
	var alerta models.Alert
	alertas := append([]interface{}{"Response:"})
	if err := json.Unmarshal(data, &CriterioIcfes); err == nil {

		criterioProyecto := make([]map[string]interface{}, 0)
		area1 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area1"])
		area2 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area2"])
		area3 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area3"])
		area4 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area4"])
		area5 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area5"])
		requestBod := "{\"Area1\": \"" + area1 + "\",\"Area2\": \"" + area2 + "\",\"Area3\": \"" + area3 + "\",\"Area4\": \"" + area4 + "\",\"Area5\": \"" + area5 + "\"}"
		for i, criterioTemp := range CriterioIcfes["Proyectos"].([]interface{}) {
			criterioProyectos := criterioTemp.(map[string]interface{})

			var criterio_existente []map[string]interface{}
			errCriterioExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/?query=ProgramaAcademicoId:"+fmt.Sprintf("%.f", criterioProyectos["Id"].(float64)), &criterio_existente)
			if errCriterioExistente == nil && fmt.Sprintf("%v", criterio_existente[0]) != "map[]" {
				if criterio_existente[0]["Status"] != 404 {
					ManejoCriterioCriterioIcfes(&criterioProyecto, CriterioIcfes, requestBod, criterioProyectos, i, &alertas, &alerta, 1, &criterio_existente)
				} else {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
					if criterio_existente[0]["Message"] == "Not found resource" {
						APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
					} else {
						logs.Error(criterio_existente)
						APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errCriterioExistente)
						return APIResponseDTO
					}
				}
			} else {
				ManejoCriterioCriterioIcfes(&criterioProyecto, CriterioIcfes, requestBod, criterioProyectos, i, &alertas, &alerta, 2, &criterio_existente)
			}
		}
		alertas = append(alertas, criterioProyecto)
		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, alerta)
		return APIResponseDTO

	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
	}
	return APIResponseDTO
}

func PuntajeTotal(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var consulta map[string]interface{}

	if err := json.Unmarshal(data, &consulta); err == nil {

		var resultado_puntaje []map[string]interface{}
		errPuntaje := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion/?query=RequisitoProgramaAcademicoId.ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",RequisitoProgramaAcademicoId.PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_puntaje)

		if errPuntaje == nil && fmt.Sprintf("%v", resultado_puntaje[0]) != "map[]" {
			if resultado_puntaje[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_puntaje)
				for i, resultado_tem := range resultado_puntaje {
					infoSystem, infoJson, exito := PeticionResultadoInscripcionGetPuntaje(resultado_tem, &resultado_puntaje, i)

					if !exito {
						if infoSystem != nil {
							APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
							return APIResponseDTO
						} else {
							APIResponseDTO = requestresponse.APIResponseDTO(true, 200, infoJson)
						}
					}

					APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado_puntaje)
				}
				APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado_puntaje)
			} else {
				if resultado_puntaje[0]["Message"] == "Not found resource" {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
				} else {
					logs.Error(resultado_puntaje)
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errPuntaje.Error())
					return APIResponseDTO
				}
			}
		} else {
			logs.Error(resultado_puntaje)
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errPuntaje.Error())
			return APIResponseDTO
		}
	} else {
		logs.Error(err)
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
		return APIResponseDTO
	}
	return APIResponseDTO
}

func CuposAdmision(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var CuposAdmision map[string]interface{}

	alertas := []interface{}{"Response:"}
	if err := json.Unmarshal(data, &CuposAdmision); err == nil {
		if fmt.Sprintf("%v", CuposAdmision) != "map[]" {
			CuposProyectos := make([]map[string]interface{}, 0)
			ComunidadesNegras := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["ComunidadesNegras"])
			DesplazadosVictimasConflicto := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["DesplazadosVictimasConflicto"])
			ComunidadesIndiginas := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["ComunidadesIndiginas"])
			MejorBachiller := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["MejorBachiller"])
			Ley1084 := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["Ley1084"])
			ProgramaReincorporacion := fmt.Sprintf("%v", CuposAdmision["CuposEspeciales"].(map[string]interface{})["ProgramaReincorporacion"])
			requestBod := "{\"ComunidadesNegras\": \"" + ComunidadesNegras + "\",\"DesplazadosVictimasConflicto\": \"" + DesplazadosVictimasConflicto + "\",\"ComunidadesIndiginas\": \"" + ComunidadesIndiginas + "\",\"MejorBachiller\": \"" + MejorBachiller + "\",\"Ley1084\": \"" + Ley1084 + "\",\"ProgramaReincorporacion\": \"" + ProgramaReincorporacion + "\"}"

			for i, cupoTemp := range CuposAdmision["Proyectos"].([]interface{}) {
				cupoProyectos := cupoTemp.(map[string]interface{})

				// // Verificar que no exista registro del cupo a cada proyecto
				if resultado := SolicituVerificacionCuposAdmision(cupoProyectos, CuposAdmision, &CuposProyectos, requestBod, i); resultado != nil {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, resultado)
					break
				}
			}

			alertas = append(alertas, CuposProyectos)
			APIResponseDTO = requestresponse.APIResponseDTO(true, 200, alertas)
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 403, nil)
		}
	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	}
	return APIResponseDTO
}

func CambioEstadoAspirante(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var consultaestado map[string]interface{}
	EstadoActulizado := "Estados Actualizados"
	alertas := append([]interface{}{"Response:"})

	if err := json.Unmarshal(data, &consultaestado); err == nil {
		Id_periodo := consultaestado["Periodo"].(map[string]interface{})["Id"]
		for _, proyectotemp := range consultaestado["Proyectos"].([]interface{}) {
			EstadoProyectos := proyectotemp.(map[string]interface{})

			infoSystem, infoJson, exito := PeticionCuposCambioEstado(EstadoProyectos, Id_periodo)

			if !exito {
				if infoSystem != nil {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
					return APIResponseDTO
				} else {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, infoJson)
				}
			}
		}
		alertas = append(alertas, EstadoActulizado)
		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, alertas)

	} else {
		logs.Error(err)
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	}

	return APIResponseDTO
}

func ConsultaAspirantes(data []byte) (APIResponseDTO requestresponse.APIResponse) {
	var consulta map[string]interface{}

	if err := json.Unmarshal(data, &consulta); err == nil {

		var resultado_aspirante []map[string]interface{}
		errAspirante := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/?query=ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_aspirante)
		if errAspirante == nil && fmt.Sprintf("%v", resultado_aspirante[0]) != "map[]" {
			if resultado_aspirante[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_aspirante)
				for i, resultado_tem := range resultado_aspirante {
					infoSystem, infoJson, exito := PeticionNotaGetAspirante(resultado_tem, &resultado_aspirante, i)

					if !exito {
						if infoSystem != nil {
							APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
							return APIResponseDTO
						} else {
							APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, infoJson)
						}
					}

					APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado_aspirante)
				}
				APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultado_aspirante)
			} else {
				if resultado_aspirante[0]["Message"] == "Not found resource" {
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
				} else {
					logs.Error(resultado_aspirante)
					//c.Data["development"] = map[string]interface{}{"Code": "404", "Body": err.Error(), "Type": "error"}
					APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errAspirante.Error())
					return APIResponseDTO
				}
			}
		} else {
			logs.Error(resultado_aspirante)
			//c.Data["development"] = map[string]interface{}{"Code": "404", "Body": err.Error(), "Type": "error"}
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, errAspirante.Error())
			return APIResponseDTO
		}

	} else {
		logs.Error(err)
		//c.Data["development"] = map[string]interface{}{"Code": "400", "Body": err.Error(), "Type": "error"}
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	}
	return APIResponseDTO
}

func ListaAspirantes(idPeriodo int64, idProyecto int64, tipoLista int64) (APIResponseDTO requestresponse.APIResponse) {
	const (
		id_periodo int8 = iota
		id_proyecto
		tipo_lista
	)

	type Params struct {
		valor int64
		err   error
	}

	var params [3]Params

	params[id_periodo].valor = idPeriodo
	params[id_proyecto].valor = int64(idProyecto)
	params[tipo_lista].valor = tipoLista

	var outputErrorInfo map[string]interface{}
	var ExistError bool = false

	var listado []map[string]interface{}

	for i, p := range params {
		if p.err != nil {
			outputErrorInfo = map[string]interface{}{"Success": false, "Status": "404", "Message": "Error service GetListaAspirantesPor: " + fmt.Sprintf("%v", params[i]) + fmt.Sprintf("%v", p.err)}
			ExistError = true
			break
		}
		if p.valor <= 0 {
			outputErrorInfo = map[string]interface{}{"Success": false, "Status": "404", "Message": "Error service GetListaAspirantesPor: " + fmt.Sprintf("%v", params[i]) + fmt.Sprintf("value <= 0: %v", p.valor)}
			ExistError = true
			break
		}
	}

	if !ExistError {
		ManejoCasosGetLista(params[tipo_lista].valor, params[id_periodo].valor, params[id_proyecto].valor, &listado)

		if len(listado) > 0 {
			APIResponseDTO = requestresponse.APIResponseDTO(true, 200, listado)
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
		}

	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, outputErrorInfo)
	}
	return APIResponseDTO
}

func DependenciaPorVinculacion(id_tercero_str string) (APIResponseDTO requestresponse.APIResponse) {
	/*
		definition de respuestas
	*/
	failureAsn := map[string]interface{}{"Success": false, "Status": "404",
		"Message": "Error service GetDependenciaPorVinculacionTercero: The request contains an incorrect parameter or no record exist", "Data": nil}
	successAns := map[string]interface{}{"Success": true, "Status": "200", "Message": "Query successful", "Data": nil}
	/*
		check validez de id tercero
	*/

	id_tercero, errId := strconv.ParseInt(id_tercero_str, 10, 64)
	if errId != nil || id_tercero <= 0 {
		if errId == nil {
			errId = fmt.Errorf("id_tercero: %d <= 0", id_tercero)
		}
		logs.Error(errId.Error())
		failureAsn["Data"] = errId.Error()
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, failureAsn)
		return APIResponseDTO
	}
	/*
		consulta vinculación tercero and check resultado válido
		DependenciaId__gt:0 -> que tenga id mayor que cero
		CargoId__in:312|320 -> parametrosId: 312: JEFE OFICINA, 320: Asistente Dependencia
	*/
	var estadoVinculacion []map[string]interface{}
	estadoVinculacionErr := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+
		fmt.Sprintf("vinculacion?query=Activo:true,DependenciaId__gt:0,CargoId__in:312|320,tercero_principal_id:%v", id_tercero_str), &estadoVinculacion)
	if estadoVinculacionErr != nil || fmt.Sprintf("%v", estadoVinculacion) == "[map[]]" {
		if estadoVinculacionErr == nil {
			estadoVinculacionErr = fmt.Errorf("vinculacion is empty: %v", estadoVinculacion)
		}
		logs.Error(estadoVinculacionErr.Error())
		failureAsn["Data"] = estadoVinculacionErr.Error()
		APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, failureAsn)
		return APIResponseDTO
	}
	/*
		preparar lista de dependencias, normalmente será una, pero se espera soportar varias por tercero
	*/
	var dependencias []int64
	for _, vinculacion := range estadoVinculacion {
		dependencias = append(dependencias, int64(vinculacion["DependenciaId"].(float64)))
	}
	/*
		entrega de respuesta existosa :)
	*/
	successAns["Data"] = map[string]interface{}{
		"DependenciaId": dependencias,
	}
	APIResponseDTO = requestresponse.APIResponseDTO(true, 200, successAns)
	return APIResponseDTO
}

func GetAspirantesDeProyectosActivos(idNiv string, idPer string, tipoLista string) (interface{}, error) {
	var proyectosP []map[string]interface{}
	var proyectosH []map[string]interface{}
	var proyectosArrMap []map[string]interface{}
	wge := new(errgroup.Group)
	var mutex sync.Mutex // Mutex para proteger el acceso a resultados


	// Obtenemos los proyectos padres
	errProyectosP := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosP)

	if errProyectosP != nil {
		logs.Error(errProyectosP.Error())
		return nil, errors.New("error del servicio GetCalendarProject: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}

	// Obtenemos los proyectos hijos
	errProyectosH := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.NivelFormacionPadreId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosH)

	if errProyectosH != nil {
		logs.Error(errProyectosH.Error())
		return nil, errors.New("error del servicio GetCalendarProject: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}

	// Combinamos los proyectos padres e hijos
	proyectos := append(proyectosP, proyectosH...)

	// Construimos la lista de proyectos con solo los campos necesarios
	wge.SetLimit(-1)
	for _, proyecto := range proyectos {
		proyecto := proyecto
		wge.Go(func () error{

			proyectoInfo := map[string]interface{}{
				"ProyectoId":     proyecto["Id"],
				"NombreProyecto": proyecto["Nombre"],
				"Aspirantes":     []map[string]interface{}{}, // Inicializamos la lista de aspirantes como vacía
			}
	
			// Obtener lista de aspirantes para el proyecto actual
			idPerInt64, _ := strconv.Atoi(idPer)
			tipoListaInt64, _ := strconv.Atoi(tipoLista)
			idProyecto := int64(proyecto["Id"].(float64)) // Convertir Id a int64
	
			listaAspirantesResponse := ListaAspirantes(int64(idPerInt64), idProyecto, int64(tipoListaInt64))
	
			// Verificar si ocurrió un error al obtener la lista de aspirantes
			if listaAspirantesResponse.Success {
				// Obtener la lista de aspirantes de la respuesta
				listaAspirantes := listaAspirantesResponse.Data.([]map[string]interface{})
	
				// Agregar la lista de aspirantes al objeto del proyecto
				proyectoInfo["Aspirantes"] = listaAspirantes
			} else {
				// Si hay un error, dejar la lista de aspirantes vacía para este proyecto
				logs.Error("No hay aspirantes para el proyecto de id: ", idProyecto)
			}
			
			mutex.Lock()
			proyectosArrMap = append(proyectosArrMap, proyectoInfo)
			mutex.Unlock()

			return nil
		})
		if err := wge.Wait(); err != nil {
			return requestresponse.APIResponseDTO(false, 400, proyectosArrMap), err
		}
	}

	return requestresponse.APIResponseDTO(true, 200, proyectosArrMap), nil
}
