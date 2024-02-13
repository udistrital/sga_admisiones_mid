package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego"

	// "github.com/udistrital/sga_mid_admisiones/models"
	"github.com/udistrital/sga_admisiones_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

// AdmisionController ...
type AdmisionController struct {
	beego.Controller
}

// URLMapping ...
func (c *AdmisionController) URLMapping() {
	c.Mapping("PostCriterioIcfes", c.PostCriterioIcfes)
	c.Mapping("GetPuntajeTotalByPeriodoByProyecto", c.GetPuntajeTotalByPeriodoByProyecto)
	c.Mapping("PostCuposAdmision", c.PostCuposAdmision)
	c.Mapping("CambioEstadoAspiranteByPeriodoByProyecto", c.CambioEstadoAspiranteByPeriodoByProyecto)
	c.Mapping("GetAspirantesByPeriodoByProyecto", c.GetAspirantesByPeriodoByProyecto)
	c.Mapping("PostEvaluacionAspirantes", c.PostEvaluacionAspirantes)
	c.Mapping("GetEvaluacionAspirantes", c.GetEvaluacionAspirantes)
	c.Mapping("PutNotaFinalAspirantes", c.PutNotaFinalAspirantes)
	c.Mapping("GetListaAspirantesPor", c.GetListaAspirantesPor)
	c.Mapping("GetDependenciaPorVinculacionTercero", c.GetDependenciaPorVinculacionTercero)
}

// PutNotaFinalAspirantes ...
// @Title PutNotaFinalAspirantes
// @Description Se calcula la nota final de cada aspirante
// @Param   body        body    {}  true        "body Calcular nota final content"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router /calcular_nota [put]
func (c *AdmisionController) PutNotaFinalAspirantes() {
	var Evaluacion map[string]interface{}
	var Inscripcion []map[string]interface{}
	var DetalleEvaluacion []map[string]interface{}
	var NotaFinal float64
	var InscripcionPut map[string]interface{}
	var respuesta []map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var errorGetAll bool
	var message string

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &Evaluacion); err == nil {
		IdPersona := Evaluacion["IdPersona"].([]interface{})
		PeriodoId := fmt.Sprintf("%v", Evaluacion["IdPeriodo"])
		ProgramaAcademicoId := fmt.Sprintf("%v", Evaluacion["IdPrograma"])
		respuesta = make([]map[string]interface{}, len(IdPersona))
		for i := 0; i < len(IdPersona); i++ {
			PersonaId := fmt.Sprintf("%v", IdPersona[i].(map[string]interface{})["Id"])

			//GET a Inscripción para obtener el ID
			errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=PersonaId:"+PersonaId+",PeriodoId:"+PeriodoId+",ProgramaAcademicoId:"+ProgramaAcademicoId, &Inscripcion)
			if errInscripcion == nil {
				if Inscripcion != nil && fmt.Sprintf("%v", Inscripcion[0]) != "map[]" {
					InscripcionId := fmt.Sprintf("%v", Inscripcion[0]["Id"])

					//GET a detalle evaluacion
					errDetalleEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=InscripcionId:"+InscripcionId+",RequisitoProgramaAcademicoId__ProgramaAcademicoId:"+ProgramaAcademicoId+",RequisitoProgramaAcademicoId__PeriodoId:"+PeriodoId+"&limit=0", &DetalleEvaluacion)
					if errDetalleEvaluacion == nil {
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
							errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, "PUT", &InscripcionPut, Inscripcion[0])
							if errInscripcionPut == nil {
								if InscripcionPut != nil && fmt.Sprintf("%v", InscripcionPut) != "map[]" {
									respuesta[i] = InscripcionPut
								} else {
									errorGetAll = true
									message = "No data found"
								}
							} else {
								errorGetAll = true
								message = errInscripcionPut.Error()
							}
						} else {
							errorGetAll = true
							message = "No data found"
						}
					} else {
						errorGetAll = true
						message = errDetalleEvaluacion.Error()
					}
				} else {
					errorGetAll = true
					message = "No data found"
				}
			} else {
				errorGetAll = true
				message = errInscripcion.Error()
			}
		}
		resultado["Response"] = respuesta
	} else {
		errorGetAll = true
		message = err.Error()
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, message)
	}

	c.ServeJSON()
}

// GetEvaluacionAspirantes ...
// @Title GetEvaluacionAspirantes
// @Description Consultar la evaluacion de los aspirantes de acuerdo a los criterios
// @Param	id_requisito	path	int	true	"Id del requisito"
// @Param	id_periodo	path	int	true	"Id del periodo"
// @Param	id_programa	path	int	true	"Id del programa academico"
// @Success 200 {}
// @Failure 404 not found resource
// @router /consultar_evaluacion/:id_programa/:id_periodo/:id_requisito [get]
func (c *AdmisionController) GetEvaluacionAspirantes() {
	id_periodo := c.Ctx.Input.Param(":id_periodo")
	id_programa := c.Ctx.Input.Param(":id_programa")
	id_requisito := c.Ctx.Input.Param(":id_requisito")
	var DetalleEvaluacion []map[string]interface{}
	var DetalleEspecificoJSON []map[string]interface{}
	var Inscripcion map[string]interface{}
	var Terceros map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var errorGetAll bool
	var message string

	//GET a la tabla detalle_evaluacion
	errDetalleEvaluacion := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion?query=RequisitoProgramaAcademicoId__RequisitoId__Id:"+id_requisito+",RequisitoProgramaAcademicoId__PeriodoId:"+id_periodo+",RequisitoProgramaAcademicoId__ProgramaAcademicoId:"+id_programa+"&sortby=InscripcionId&order=asc", &DetalleEvaluacion)
	if errDetalleEvaluacion == nil {
		if DetalleEvaluacion != nil && fmt.Sprintf("%v", DetalleEvaluacion[0]) != "map[]" {
			Respuesta := "[\n"
			for i, evaluacion := range DetalleEvaluacion {
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
					InscripcionId := fmt.Sprintf("%v", evaluacion["InscripcionId"])
					errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+InscripcionId, &Inscripcion)
					if errInscripcion == nil {
						if Inscripcion != nil && fmt.Sprintf("%v", Inscripcion) != "map[]" {

							//GET a la tabla de terceros para obtener el nombre
							TerceroId := fmt.Sprintf("%v", Inscripcion["PersonaId"])
							errTerceros := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+TerceroId, &Terceros)
							if errTerceros == nil {
								if Terceros != nil && fmt.Sprintf("%v", Terceros) != "map[]" {
									respuestaAux = respuestaAux + "\"Aspirantes\": " + fmt.Sprintf("%q", Terceros["NombreCompleto"]) + "\n}"
								} else {
									errorGetAll = true
									message = "No data found"
								}
							} else {
								errorGetAll = true
								message = errTerceros.Error()
							}
						} else {
							errorGetAll = true
							message = "No data found"
						}
					} else {
						errorGetAll = true
						message = errInscripcion.Error()
					}

					if i+1 == len(DetalleEvaluacion) {
						Respuesta = Respuesta + respuestaAux + "\n]"
					} else {
						Respuesta = Respuesta + respuestaAux + ",\n"
					}
				}
			}
			if err := json.Unmarshal([]byte(Respuesta), &DetalleEspecificoJSON); err == nil {
				resultado["areas"] = DetalleEspecificoJSON
			}
		} else {
			errorGetAll = true
			message = "No data found"
		}

	} else {
		errorGetAll = true
		message = errDetalleEvaluacion.Error()
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, message)
	}

	c.ServeJSON()
}

// PostEvaluacionAspirantes ...
// @Title PostEvaluacionAspirantes
// @Description Agregar la evaluacion de los aspirantes de acuerdo a los criterios
// @Param   body        body    {}  true        "body Agregar evaluacion aspirantes content"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router /registrar_evaluacion [post]
func (c *AdmisionController) PostEvaluacionAspirantes() {
	var Evaluacion map[string]interface{}
	var Inscripciones []map[string]interface{}
	var Requisito []map[string]interface{}
	var DetalleCalificacion string
	var Ponderado float64
	var respuesta []map[string]interface{}
	var DetalleEvaluacion map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var errorGetAll bool
	var message string
	//Calificacion = append([]interface{}{"areas"})

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &Evaluacion); err == nil {
		AspirantesData := Evaluacion["Aspirantes"].([]interface{})
		ProgramaAcademicoId := Evaluacion["ProgramaId"]
		PeriodoId := Evaluacion["PeriodoId"]
		CriterioId := Evaluacion["CriterioId"]
		respuesta = make([]map[string]interface{}, len(AspirantesData))
		//GET para obtener el porcentaje general, especifico (si lo hay)
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
						errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion?query=PersonaId:"+fmt.Sprintf("%v", PersonaId)+",ProgramaAcademicoId:"+fmt.Sprintf("%v", ProgramaAcademicoId)+",PeriodoId:"+fmt.Sprintf("%v", PeriodoId), &Inscripciones)
						if errInscripcion == nil {
							if Inscripciones != nil && fmt.Sprintf("%v", Inscripciones[0]) != "map[]" {
								if PorcentajeEspJSON != nil && fmt.Sprintf("%v", PorcentajeEspJSON) != "map[]" {
									//Calculos para los criterios que cuentan con subcriterios)
									Ponderado = 0
									DetalleCalificacion = "{\n\"areas\":\n["
									ultimo := false

									for k := range PorcentajeEspJSON["areas"].([]interface{}) {
										for _, aux := range PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{}) {
											for k2, aux2 := range Evaluacion["Aspirantes"].([]interface{})[i].(map[string]interface{}) {
												if ultimo {
													break
												}
												if aux == k2 {
													//Si existe la columna de asistencia se hace la validación de la misma
													if Asistencia != nil {
														if Asistencia == true {
															f, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{})["Porcentaje"]), 64) //Porcentaje del subcriterio
															j, _ := strconv.ParseFloat(fmt.Sprintf("%v", aux2), 64)                                                                                 //Nota subcriterio
															PonderadoAux := j * (f / 100)
															Ponderado = Ponderado + PonderadoAux
															if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
																DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
																ultimo = true
															} else {
																DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
															}
														} else {
															// Si el estudiante inscrito no asiste tendrá una calificación de 0
															Ponderado = 0
															if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
																DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":\"0\", \"Ponderado\":\"0\"},\n"
																ultimo = true
															} else {
																DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":\"0\", \"Ponderado\":\"0\"},\n"
															}
														}
													} else {
														f, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeEspJSON["areas"].([]interface{})[k].(map[string]interface{})["Porcentaje"]), 64) //Porcentaje del subcriterio
														j, _ := strconv.ParseFloat(fmt.Sprintf("%v", aux2), 64)                                                                                 //Nota subcriterio
														PonderadoAux := j * (f / 100)
														Ponderado = Ponderado + PonderadoAux
														if k+1 == len(PorcentajeEspJSON["areas"].([]interface{})) {
															DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
															ultimo = true
														} else {
															DetalleCalificacion = DetalleCalificacion + "{" + fmt.Sprintf("%q", k2) + ":" + fmt.Sprintf("%q", aux2) + ", \"Ponderado\":" + fmt.Sprintf("%.2f", PonderadoAux) + "},\n"
														}
													}
												}
											}
										}
									}
									g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)
									Ponderado = Ponderado * (g / 100)
									if Asistencia == true && Asistencia != nil {
										DetalleCalificacion = DetalleCalificacion + "{\"Asistencia\": true" + "}]\n}"
									} else {
										DetalleCalificacion = DetalleCalificacion + "{\"Asistencia\": false" + "}]\n}"
									}
								} else {
									//Calculos para los criterios que no tienen subcriterios
									//Si existe la columna de asistencia se hace la validación de la misma
									if Asistencia != nil {
										if Asistencia == true {
											f, _ := strconv.ParseFloat(fmt.Sprintf("%v", AspirantesData[i].(map[string]interface{})["Puntuacion"]), 64) //Puntaje del aspirante
											g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)                                        //Porcentaje del criterio
											Ponderado = f * (g / 100)                                                                                   //100% del puntaje que obtuvo el aspirante
											DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\":" + fmt.Sprintf("%q", AspirantesData[i].(map[string]interface{})["Puntuacion"]) + "}\n]\n}"
										} else {
											// Si el estudiante inscrito no asiste tendrá una calificación de 0
											Ponderado = 0
											DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\": \"0\"}\n]\n}"
										}
									} else {
										f, _ := strconv.ParseFloat(fmt.Sprintf("%v", AspirantesData[i].(map[string]interface{})["Puntuacion"]), 64) //Puntaje del aspirante
										g, _ := strconv.ParseFloat(fmt.Sprintf("%v", PorcentajeGeneral), 64)                                        //Porcentaje del criterio
										Ponderado = f * (g / 100)                                                                                   //100% del puntaje que obtuvo el aspirante
										DetalleCalificacion = "{\n \"areas\": [\n {\"Puntuacion\":" + fmt.Sprintf("%q", AspirantesData[i].(map[string]interface{})["Puntuacion"]) + "}\n]\n}"
									}
								}
								// JSON para el post detalle evaluacion
								respuesta[i] = map[string]interface{}{
									"InscripcionId":                Inscripciones[0]["Id"],
									"RequisitoProgramaAcademicoId": Requisito[0],
									"Activo":                       true,
									"FechaCreacion":                time.Now(),
									"FechaModificacion":            time.Now(),
									"DetalleCalificacion":          DetalleCalificacion,
									"NotaRequisito":                Ponderado,
								}
								//Función POST a la tabla detalle_evaluación
								errDetalleEvaluacion := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion", "POST", &DetalleEvaluacion, respuesta[i])
								if errDetalleEvaluacion == nil {
									if DetalleEvaluacion != nil && fmt.Sprintf("%v", DetalleEvaluacion) != "map[]" {
										//respuesta[i] = DetalleEvaluacion
									} else {
										errorGetAll = true
										message = "No data found"
									}
								} else {
									errorGetAll = true
									message = errDetalleEvaluacion.Error()
								}
							} else {
								errorGetAll = true
								message = "No data found"
							}
						} else {
							errorGetAll = true
							message = errInscripcion.Error()
						}
					}
				}
			} else {
				errorGetAll = true
				message = "No data found"
			}
		} else {
			errorGetAll = true
			message = errRequisito.Error()
		}

		resultado["Evaluacion"] = respuesta
	} else {
		errorGetAll = true
		message = err.Error()
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, message)
	}

	c.ServeJSON()

}

// PostCriterioIcfes ...
// @Title PostCriterioIcfes
// @Description Agregar CriterioIcfes
// @Param   body        body    {}  true        "body Agregar CriterioIcfes content"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router / [post]
func (c *AdmisionController) PostCriterioIcfes() {
	var CriterioIcfes map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &CriterioIcfes); err == nil {

		criterioProyecto := make([]map[string]interface{}, 0)
		area1 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area1"])
		area2 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area2"])
		area3 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area3"])
		area4 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area4"])
		area5 := fmt.Sprintf("%v", CriterioIcfes["Especifico"].(map[string]interface{})["Area5"])
		requestBod := "{\"Area1\": \"" + area1 + "\",\"Area2\": \"" + area2 + "\",\"Area3\": \"" + area3 + "\",\"Area4\": \"" + area4 + "\",\"Area5\": \"" + area5 + "\"}"
		for i, criterioTemp := range CriterioIcfes["Proyectos"].([]interface{}) {
			criterioProyectos := criterioTemp.(map[string]interface{})

			// // Verificar que no exista registro del criterio a cada proyecto
			//fmt.Sprintf("%.f", criterioProyectos["Id"].(float64))
			var criterio_existente []map[string]interface{}
			errCriterioExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/?query=ProgramaAcademicoId:"+fmt.Sprintf("%.f", criterioProyectos["Id"].(float64)), &criterio_existente)
			if errCriterioExistente == nil && fmt.Sprintf("%v", criterio_existente[0]) != "map[]" {
				if criterio_existente[0]["Status"] != 404 {
					fmt.Println("Existe criterio")
					Id_criterio_existente := criterio_existente[0]["Id"]
					criterioProyecto = append(criterioProyecto, map[string]interface{}{
						"Activo":               true,
						"PeriodoId":            CriterioIcfes["Periodo"].(map[string]interface{})["Id"],
						"PorcentajeEspecifico": requestBod,
						"PorcentajeGeneral":    CriterioIcfes["General"],
						"ProgramaAcademicoId":  criterioProyectos["Id"],
						"RequisitoId":          map[string]interface{}{"Id": 1},
					})

					// Put a criterio Existente

					var resultadoPutcriterio map[string]interface{}
					errPutCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/"+fmt.Sprintf("%.f", Id_criterio_existente.(float64)), "PUT", &resultadoPutcriterio, criterioProyecto[i])
					if resultadoPutcriterio["Type"] == "error" || errPutCriterio != nil || resultadoPutcriterio["Status"] == "404" || resultadoPutcriterio["Message"] != nil {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPutCriterio.Error())
					} else {
						fmt.Println("Registro  PUT de criterios bien")
					}

				} else {
					if criterio_existente[0]["Message"] == "Not found resource" {
						c.Data["json"] = nil
					} else {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errCriterioExistente.Error())
					}
				}
			} else {
				fmt.Println("No Existe criterio")
				criterioProyecto = append(criterioProyecto, map[string]interface{}{
					"Activo":               true,
					"PeriodoId":            CriterioIcfes["Periodo"].(map[string]interface{})["Id"],
					"PorcentajeEspecifico": requestBod,
					"PorcentajeGeneral":    CriterioIcfes["General"],
					"ProgramaAcademicoId":  criterioProyectos["Id"],
					"RequisitoId":          map[string]interface{}{"Id": 1},
				})

				var resultadocriterio map[string]interface{}
				errPostCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico", "POST", &resultadocriterio, criterioProyecto[i])
				if resultadocriterio["Type"] == "error" || errPostCriterio != nil || resultadocriterio["Status"] == "404" || resultadocriterio["Message"] != nil {
					c.Ctx.Output.SetStatus(http.StatusBadRequest)
					c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPostCriterio.Error())
				} else {
					fmt.Println("Registro de criterios bien")
				}
			}
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, criterioProyecto)
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
	}
	c.ServeJSON()
}

// ConsultarPuntajeTotalByPeriodoByProyecto ...
// @Title GetPuntajeTotalByPeriodoByProyecto
// @Description get PuntajeTotalCriteio by id_periodo and id_proyecto
// @Param	body		body 	{}	true		"body for Get Puntaje total content"
// @Success 201 {int}
// @Failure 400 the request contains incorrect syntax
// @router /consulta_puntaje [post]
func (c *AdmisionController) GetPuntajeTotalByPeriodoByProyecto() {
	var consulta map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consulta); err == nil {

		var resultado_puntaje []map[string]interface{}
		errPuntaje := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion/?query=RequisitoProgramaAcademicoId.ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",RequisitoProgramaAcademicoId.PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_puntaje)

		if errPuntaje == nil && fmt.Sprintf("%v", resultado_puntaje[0]) != "map[]" {
			if resultado_puntaje[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_puntaje)
				for i, resultado_tem := range resultado_puntaje {
					id_inscripcion := (resultado_tem["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]).(float64)

					var resultado_inscripcion map[string]interface{}
					errGetInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", id_inscripcion), &resultado_inscripcion)
					if errGetInscripcion == nil && fmt.Sprintf("%v", resultado_inscripcion) != "map[]" {
						if resultado_inscripcion["Status"] != 404 {
							id_persona := (resultado_inscripcion["PersonaId"]).(float64)

							var resultado_persona map[string]interface{}
							errGetPersona := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+fmt.Sprintf("%v", id_persona), &resultado_persona)
							if errGetPersona == nil && fmt.Sprintf("%v", resultado_persona) != "map[]" {
								if resultado_persona["Status"] != 404 {
									resultado_puntaje[i]["NombreAspirante"] = resultado_persona["NombreCompleto"]
									var resultado_documento []map[string]interface{}
									errGetDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion/?query=TerceroId.Id:"+fmt.Sprintf("%v", id_persona), &resultado_documento)
									if errGetDocumento == nil && fmt.Sprintf("%v", resultado_documento[0]) != "map[]" {
										if resultado_documento[0]["Status"] != 404 {

											resultado_puntaje[i]["TipoDocumento"] = resultado_documento[0]["TipoDocumentoId"].(map[string]interface{})["CodigoAbreviacion"]
											resultado_puntaje[i]["NumeroDocumento"] = resultado_documento[0]["Numero"]
										} else {
											if resultado_documento[0]["Message"] == "Not found resource" {
												c.Data["json"] = nil
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetDocumento.Error())
											}
										}
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetDocumento.Error())
									}

									//hh
								} else {
									if resultado_persona["Message"] == "Not found resource" {
										c.Data["json"] = nil
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetPersona.Error())
									}
								}
							} else {
								c.Ctx.Output.SetStatus(http.StatusBadRequest)
								c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetPersona.Error())
							}
						} else {
							if resultado_inscripcion["Message"] == "Not found resource" {
								c.Data["json"] = nil
							} else {
								c.Ctx.Output.SetStatus(http.StatusBadRequest)
								c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetInscripcion.Error())
							}
						}
					} else {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetInscripcion.Error())
					}
					c.Ctx.Output.SetStatus(http.StatusOK)
					c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, resultado_puntaje)
				}

			} else {
				if resultado_puntaje[0]["Message"] == "Not found resource" {
					c.Data["json"] = nil
				} else {
					c.Ctx.Output.SetStatus(http.StatusBadRequest)
					c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPuntaje.Error())
				}
			}
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPuntaje.Error())
		}

	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, err.Error())
	}
	c.ServeJSON()
}

// PostCuposAdmision ...
// @Title PostCuposAdmision
// @Description Agregar PostCuposAdmision
// @Param   body        body    {}  true        "body Agregar PostCuposAdmision content"
// @Success 200 {}
// @Failure 400 the request contains incorrect syntax
// @router /postcupos [post]
func (c *AdmisionController) PostCuposAdmision() {
	var CuposAdmision map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &CuposAdmision); err == nil {
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
				var cupos_existente []map[string]interface{}
				errCupoExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/?query=DependenciaId:"+fmt.Sprintf("%.f", cupoProyectos["Id"].(float64))+",PeriodoId:"+fmt.Sprintf("%.f", CuposAdmision["Periodo"].(map[string]interface{})["Id"].(float64)), &cupos_existente)
				if errCupoExistente == nil && fmt.Sprintf("%v", cupos_existente[0]) != "map[]" {
					if cupos_existente[0]["Status"] != 404 {
						fmt.Println("Existe cupos para el proyecto")
						Id_cupo_existente := cupos_existente[0]["Id"]
						CuposProyectos = append(CuposProyectos, map[string]interface{}{
							"Activo":           true,
							"PeriodoId":        CuposAdmision["Periodo"].(map[string]interface{})["Id"],
							"CuposEspeciales":  requestBod,
							"CuposHabilitados": CuposAdmision["CuposAsignados"],
							"DependenciaId":    cupoProyectos["Id"],
							"CuposOpcionados":  CuposAdmision["CuposOpcionados"],
						})

						// Put a cupo Existente

						var resultadoPutcupo map[string]interface{}
						errPutCriterio := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/"+fmt.Sprintf("%.f", Id_cupo_existente.(float64)), "PUT", &resultadoPutcupo, CuposProyectos[i])
						if resultadoPutcupo["Type"] == "error" || errPutCriterio != nil || resultadoPutcupo["Status"] == "404" || resultadoPutcupo["Message"] != nil {
							c.Ctx.Output.SetStatus(http.StatusBadRequest)
							c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPutCriterio.Error())
						} else {
							fmt.Println("Registro  PUT de cupo bien")
						}

					} else {
						if cupos_existente[0]["Message"] == "Not found resource" {
							c.Ctx.Output.SetStatus(http.StatusBadRequest)
							c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errCupoExistente.Error())
						} else {
							c.Ctx.Output.SetStatus(http.StatusBadRequest)
							c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errCupoExistente.Error())
						}
					}
				} else {
					fmt.Println("No Existe cupo")
					CuposProyectos = append(CuposProyectos, map[string]interface{}{
						"Activo":           true,
						"PeriodoId":        CuposAdmision["Periodo"].(map[string]interface{})["Id"],
						"CuposEspeciales":  requestBod,
						"CuposHabilitados": CuposAdmision["CuposAsignados"],
						"DependenciaId":    cupoProyectos["Id"],
						"CuposOpcionados":  CuposAdmision["CuposOpcionados"],
					})

					var resultadocupopost map[string]interface{}
					errPostCupo := request.SendJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia", "POST", &resultadocupopost, CuposProyectos[i])
					if resultadocupopost["Type"] == "error" || errPostCupo != nil || resultadocupopost["Status"] == "404" || resultadocupopost["Message"] != nil {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errPostCupo.Error())
					} else {
						fmt.Println("Registro de cupo bien")
					}
				}
			}
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, CuposProyectos)
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, "Body is empty")
		}
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, err.Error())
	}

	c.ServeJSON()
}

// CambioEstadoAspiranteByPeriodoByProyecto ...
// @Title CambioEstadoAspiranteByPeriodoByProyecto
// @Description post cambioestadoaspirante by id_periodo and id_proyecto
// @Param   body        body    {}  true        "body for  post cambio estadocontent"
// @Success 200 {}
// @Failure 400 the request content incorrect syntax
// @router /cambioestado [post]
func (c *AdmisionController) CambioEstadoAspiranteByPeriodoByProyecto() {
	var consultaestado map[string]interface{}
	EstadoActulizado := "Estados Actualizados"

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consultaestado); err == nil {
		Id_periodo := consultaestado["Periodo"].(map[string]interface{})["Id"]
		for _, proyectotemp := range consultaestado["Proyectos"].([]interface{}) {
			EstadoProyectos := proyectotemp.(map[string]interface{})

			var resultadocupo []map[string]interface{}
			errCupo := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"cupos_por_dependencia/?query=DependenciaId:"+fmt.Sprintf("%v", EstadoProyectos["Id"])+",PeriodoId:"+fmt.Sprintf("%v", Id_periodo), &resultadocupo)

			if errCupo == nil && fmt.Sprintf("%v", resultadocupo[0]) != "map[]" {
				if resultadocupo[0]["Status"] != 404 {
					CuposHabilitados, _ := strconv.ParseInt(fmt.Sprintf("%v", resultadocupo[0]["CuposHabilitados"]), 10, 64)
					CuposOpcionados, _ := strconv.ParseInt(fmt.Sprintf("%v", resultadocupo[0]["CuposOpcionados"]), 10, 64)
					// consulta id inscripcion y nota final para cada proyecto con periodo, organiza el array de forma de descendente por el campo nota final para organizar del mayor puntaje al menor
					var resultadoaspirantenota []map[string]interface{}
					errconsulta := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion/?query=RequisitoProgramaAcademicoId.ProgramaAcademicoId:"+fmt.Sprintf("%v", EstadoProyectos["Id"])+",RequisitoProgramaAcademicoId.PeriodoId:"+fmt.Sprintf("%v", Id_periodo)+"&limit=0&sortby=EvaluacionInscripcionId__NotaFinal&order=desc", &resultadoaspirantenota)
					if errconsulta == nil && fmt.Sprintf("%v", resultadoaspirantenota[0]) != "map[]" {
						if resultadoaspirantenota[0]["Status"] != 404 {

							for e, estadotemp := range resultadoaspirantenota {
								if e < (int(CuposHabilitados)) {

									// Se realiza get a la informacion del inscrito
									var resultadoaspiranteinscripcion map[string]interface{}
									errinscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), &resultadoaspiranteinscripcion)
									if errinscripcion == nil && fmt.Sprintf("%v", resultadoaspiranteinscripcion) != "map[]" {
										if resultadoaspiranteinscripcion["Status"] != 404 {

											//Actualiza el estado de inscripcio id =2 = ADMITIDO
											resultadoaspiranteinscripcion["EstadoInscripcionId"] = map[string]interface{}{"Id": 2}

											var inscripcionPut map[string]interface{}
											errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%.f", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"].(float64)), "PUT", &inscripcionPut, resultadoaspiranteinscripcion)
											if errInscripcionPut == nil && fmt.Sprintf("%v", inscripcionPut) != "map[]" && inscripcionPut["Id"] != nil {
												if inscripcionPut["Status"] != 400 {
													fmt.Println("Put correcto Admitido")

												} else {
													var resultado2 map[string]interface{}
													request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"/inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), "DELETE", &resultado2, nil)
													c.Ctx.Output.SetStatus(http.StatusBadRequest)
													c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
												}
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
											}

										} else {
											if resultadoaspiranteinscripcion["Message"] == "Not found resource" {
												c.Data["json"] = nil
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
											}
										}
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
									}

								}
								if e >= int(CuposHabilitados) && e < (int(CuposHabilitados)+int(CuposOpcionados)) {

									var resultadoaspiranteinscripcion map[string]interface{}
									errinscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), &resultadoaspiranteinscripcion)
									if errinscripcion == nil && fmt.Sprintf("%v", resultadoaspiranteinscripcion) != "map[]" {
										if resultadoaspiranteinscripcion["Status"] != 404 {

											//Actualiza el estado de inscripcio id =3 = OPCIONADO
											resultadoaspiranteinscripcion["EstadoInscripcionId"] = map[string]interface{}{"Id": 3}

											var inscripcionPut map[string]interface{}
											errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%.f", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"].(float64)), "PUT", &inscripcionPut, resultadoaspiranteinscripcion)
											if errInscripcionPut == nil && fmt.Sprintf("%v", inscripcionPut) != "map[]" && inscripcionPut["Id"] != nil {
												if inscripcionPut["Status"] != 400 {
													fmt.Println("Put correcto OPCIONADO")

												} else {
													var resultado2 map[string]interface{}
													request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"/inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), "DELETE", &resultado2, nil)
													c.Ctx.Output.SetStatus(http.StatusBadRequest)
													c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
												}
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
											}

										} else {
											if resultadoaspiranteinscripcion["Message"] == "Not found resource" {
												c.Data["json"] = nil
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
											}
										}
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
									}
								}
								if e >= (int(CuposHabilitados) + int(CuposOpcionados)) {

									var resultadoaspiranteinscripcion map[string]interface{}
									errinscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), &resultadoaspiranteinscripcion)
									if errinscripcion == nil && fmt.Sprintf("%v", resultadoaspiranteinscripcion) != "map[]" {
										if resultadoaspiranteinscripcion["Status"] != 404 {

											//Actualiza el estado de inscripcio id =4 = NOADMITIDO
											resultadoaspiranteinscripcion["EstadoInscripcionId"] = map[string]interface{}{"Id": 4}

											var inscripcionPut map[string]interface{}
											errInscripcionPut := request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/"+fmt.Sprintf("%.f", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"].(float64)), "PUT", &inscripcionPut, resultadoaspiranteinscripcion)
											if errInscripcionPut == nil && fmt.Sprintf("%v", inscripcionPut) != "map[]" && inscripcionPut["Id"] != nil {
												if inscripcionPut["Status"] != 400 {
													fmt.Println("Put correcto NO ADMITIDO")

												} else {
													var resultado2 map[string]interface{}
													request.SendJson("http://"+beego.AppConfig.String("InscripcionService")+"/inscripcion/"+fmt.Sprintf("%v", estadotemp["EvaluacionInscripcionId"].(map[string]interface{})["InscripcionId"]), "DELETE", &resultado2, nil)
													c.Ctx.Output.SetStatus(http.StatusBadRequest)
													c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
												}
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errInscripcionPut.Error())
											}

										} else {
											if resultadoaspiranteinscripcion["Message"] == "Not found resource" {
												c.Data["json"] = nil
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
											}
										}
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errinscripcion.Error())
									}
								}

							}

						} else {
							if resultadoaspirantenota[0]["Message"] == "Not found resource" {
								c.Data["json"] = nil
							} else {
								c.Ctx.Output.SetStatus(http.StatusBadRequest)
								c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errconsulta.Error())
							}
						}
					} else {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errconsulta.Error())
					}

				} else {
					if resultadocupo[0]["Message"] == "Not found resource" {
						c.Data["json"] = nil
					} else {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errCupo.Error())
					}
				}
			} else {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errCupo.Error())
			}
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, EstadoActulizado)

	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, err.Error())
	}
	c.ServeJSON()
}

// GetAspirantesByPeriodoByProyecto ...
// @Title GetAspirantesByPeriodoByProyecto
// @Description get Aspirantes by id_periodo and id_proyecto
// @Param	body		body 	{}	true		"body for Get Aspirantes content"
// @Success 201 {int}
// @Failure 404 not found resource
// @router /consulta_aspirantes [post]
func (c *AdmisionController) GetAspirantesByPeriodoByProyecto() {
	var consulta map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consulta); err == nil {

		var resultado_aspirante []map[string]interface{}
		errAspirante := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/?query=ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_aspirante)
		if errAspirante == nil && fmt.Sprintf("%v", resultado_aspirante[0]) != "map[]" {
			if resultado_aspirante[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_aspirante)
				for i, resultado_tem := range resultado_aspirante {

					id_inscripcion := (resultado_tem["Id"]).(float64)
					var resultado_nota []map[string]interface{}
					errGetNota := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"evaluacion_inscripcion/?query=InscripcionId:"+fmt.Sprintf("%v", id_inscripcion), &resultado_nota)
					if errGetNota == nil && fmt.Sprintf("%v", resultado_nota[0]) != "map[]" {
						if resultado_nota[0]["Status"] != 404 {
							resultado_aspirante[i]["NotaFinal"] = resultado_nota[0]["NotaFinal"]

							id_persona := (resultado_tem["PersonaId"]).(float64)

							var resultado_persona map[string]interface{}
							errGetPersona := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"tercero/"+fmt.Sprintf("%v", id_persona), &resultado_persona)
							if errGetPersona == nil && fmt.Sprintf("%v", resultado_persona) != "map[]" {
								if resultado_persona["Status"] != 404 {
									resultado_aspirante[i]["NombreAspirante"] = resultado_persona["NombreCompleto"]
									var resultado_documento []map[string]interface{}
									errGetDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"datos_identificacion/?query=TerceroId.Id:"+fmt.Sprintf("%v", id_persona), &resultado_documento)
									if errGetDocumento == nil && fmt.Sprintf("%v", resultado_documento[0]) != "map[]" {
										if resultado_documento[0]["Status"] != 404 {
											resultado_aspirante[i]["TipoDocumento"] = resultado_documento[0]["TipoDocumentoId"].(map[string]interface{})["CodigoAbreviacion"]
											resultado_aspirante[i]["NumeroDocumento"] = resultado_documento[0]["Numero"]
										} else {
											if resultado_documento[0]["Message"] == "Not found resource" {
												c.Data["json"] = nil
											} else {
												c.Ctx.Output.SetStatus(http.StatusBadRequest)
												c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetDocumento.Error())
											}
										}
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetDocumento.Error())
									}

									//hh
								} else {
									if resultado_persona["Message"] == "Not found resource" {
										c.Data["json"] = nil
									} else {
										c.Ctx.Output.SetStatus(http.StatusBadRequest)
										c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetPersona.Error())
									}
								}
							} else {
								c.Ctx.Output.SetStatus(http.StatusBadRequest)
								c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetPersona.Error())
							}
							//ojo
						} else {
							if resultado_nota[0]["Message"] == "Not found resource" {
								c.Data["json"] = nil
							} else {
								c.Ctx.Output.SetStatus(http.StatusBadRequest)
								c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetNota.Error())
							}
						}
					} else {
						c.Ctx.Output.SetStatus(http.StatusBadRequest)
						c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errGetNota.Error())
					}

					c.Ctx.Output.SetStatus(http.StatusOK)
					c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, resultado_aspirante)
				}

			} else {
				if resultado_aspirante[0]["Message"] == "Not found resource" {
					c.Data["json"] = nil
				} else {
					c.Ctx.Output.SetStatus(http.StatusBadRequest)
					c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errAspirante.Error())
				}
			}
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errAspirante.Error())
		}

	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, err.Error())
	}
	c.ServeJSON()
}

// GetListaAspirantesPor ...
// @Title GetListaAspirantesPor
// @Description get Lista estados aspirantes by id_periodo id_nivel id_proyecto and tipo_lista
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_nivel		query 	int	true		"Id del nivel"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Param	tipo_lista		query 	string	true		"tipo de lista"
// @Success 200 {}
// @Failure 404 not found resource
// @router /getlistaaspirantespor [get]
func (c *AdmisionController) GetListaAspirantesPor() {

	const (
		id_periodo int8 = iota
		//id_nivel
		id_proyecto
		tipo_lista
	)

	type Params struct {
		valor int64
		err   error
	}

	var params [3]Params

	params[id_periodo].valor, params[id_periodo].err = c.GetInt64("id_periodo")
	//params[id_nivel].valor, params[id_nivel].err = c.GetInt64("id_nivel")
	params[id_proyecto].valor, params[id_proyecto].err = c.GetInt64("id_proyecto")
	params[tipo_lista].valor, params[tipo_lista].err = c.GetInt64("tipo_lista")

	var outputErrorInfo map[string]interface{}
	var ExistError bool = false

	var listado []map[string]interface{}

	for i, p := range params {
		if p.err != nil {
			outputErrorInfo = map[string]interface{}{"Success": false, "Status": "404", "Message": "Error service GetListaAspirantesPor: " + fmt.Sprintf("param[%v] ", i) + fmt.Sprintf("%v", p.err)}
			ExistError = true
			break
		}
		if p.valor <= 0 {
			outputErrorInfo = map[string]interface{}{"Success": false, "Status": "404", "Message": "Error service GetListaAspirantesPor: " + fmt.Sprintf("param[%v] ", i) + fmt.Sprintf("value <= 0: %v", p.valor)}
			ExistError = true
			break
		}
	}

	if !ExistError {

		switch params[tipo_lista].valor {
		case 1:
			var inscripcion1 []map[string]interface{}
			errInscripcion1 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:5,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion1)
			if errInscripcion1 == nil && fmt.Sprintf("%v", inscripcion1) != "[map[]]" {
				for _, inscrip1 := range inscripcion1 {
					var datoIdentif1 []map[string]interface{}
					errDatoIdentif1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip1["PersonaId"]), &datoIdentif1)
					if errDatoIdentif1 == nil && fmt.Sprintf("%v", datoIdentif1) != "[map[]]" {
						listado = append(listado, map[string]interface{}{
							"Credencial":     inscrip1["Id"],
							"Identificacion": datoIdentif1[0]["Numero"],
							"Nombre":         datoIdentif1[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
							"Estado":         inscrip1["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
						})
					} else {
						var datoIdentif1_2intento []map[string]interface{}
						errDatoIdentif1_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip1["PersonaId"]), &datoIdentif1_2intento)
						if errDatoIdentif1_2intento == nil && fmt.Sprintf("%v", datoIdentif1_2intento) != "[map[]]" {
							listado = append(listado, map[string]interface{}{
								"Credencial":     inscrip1["Id"],
								"Identificacion": "",
								"Nombre":         datoIdentif1_2intento[0]["NombreCompleto"],
								"Estado":         inscrip1["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
							})
						}
					}
				}
			}
			var inscripcion2 []map[string]interface{}
			errInscripcion2 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:2,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion2)
			if errInscripcion2 == nil && fmt.Sprintf("%v", inscripcion2) != "[map[]]" {
				for _, inscrip2 := range inscripcion2 {
					var datoIdentif2 []map[string]interface{}
					errDatoIdentif2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip2["PersonaId"]), &datoIdentif2)
					if errDatoIdentif2 == nil && fmt.Sprintf("%v", datoIdentif2) != "[map[]]" {
						listado = append(listado, map[string]interface{}{
							"Credencial":     inscrip2["Id"],
							"Identificacion": datoIdentif2[0]["Numero"],
							"Nombre":         datoIdentif2[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
							"Estado":         inscrip2["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
						})
					} else {
						var datoIdentif2_2intento []map[string]interface{}
						errDatoIdentif2_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip2["PersonaId"]), &datoIdentif2_2intento)
						if errDatoIdentif2_2intento == nil && fmt.Sprintf("%v", datoIdentif2_2intento) != "[map[]]" {
							listado = append(listado, map[string]interface{}{
								"Credencial":     inscrip2["Id"],
								"Identificacion": "",
								"Nombre":         datoIdentif2_2intento[0]["NombreCompleto"],
								"Estado":         inscrip2["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
							})
						}
					}
				}
			}
			var inscripcion3 []map[string]interface{}
			errInscripcion3 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:6,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion3)
			if errInscripcion3 == nil && fmt.Sprintf("%v", inscripcion3) != "[map[]]" {
				for _, inscrip3 := range inscripcion3 {
					var datoIdentif3 []map[string]interface{}
					errDatoIdentif3 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip3["PersonaId"]), &datoIdentif3)
					if errDatoIdentif3 == nil && fmt.Sprintf("%v", datoIdentif3) != "[map[]]" {
						listado = append(listado, map[string]interface{}{
							"Credencial":     inscrip3["Id"],
							"Identificacion": datoIdentif3[0]["Numero"],
							"Nombre":         datoIdentif3[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
							"Estado":         inscrip3["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
						})
					} else {
						var datoIdentif3_2intento []map[string]interface{}
						errDatoIdentif3_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip3["PersonaId"]), &datoIdentif3_2intento)
						if errDatoIdentif3_2intento == nil && fmt.Sprintf("%v", datoIdentif3_2intento) != "[map[]]" {
							listado = append(listado, map[string]interface{}{
								"Credencial":     inscrip3["Id"],
								"Identificacion": "",
								"Nombre":         datoIdentif3_2intento[0]["NombreCompleto"],
								"Estado":         inscrip3["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
							})
						}
					}
				}
			}

		case 2:
			var inscripcion1 []map[string]interface{}
			errInscripcion1 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:5,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion1)
			if errInscripcion1 == nil && fmt.Sprintf("%v", inscripcion1) != "[map[]]" {
				for _, inscrip1 := range inscripcion1 {
					var datoIdentif1 []map[string]interface{}
					errDatoIdentif1 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip1["PersonaId"]), &datoIdentif1)
					if errDatoIdentif1 == nil && fmt.Sprintf("%v", datoIdentif1) != "[map[]]" {
						listado = append(listado, map[string]interface{}{
							"Id":         inscrip1["PersonaId"],
							"Aspirantes": datoIdentif1[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
						})
					} else {
						var datoIdentif1_2intento []map[string]interface{}
						errDatoIdentif1_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip1["PersonaId"]), &datoIdentif1_2intento)
						if errDatoIdentif1_2intento == nil && fmt.Sprintf("%v", datoIdentif1_2intento) != "[map[]]" {
							listado = append(listado, map[string]interface{}{
								"Id":         inscrip1["PersonaId"],
								"Aspirantes": datoIdentif1_2intento[0]["NombreCompleto"],
							})
						}
					}
				}
			}
			var inscripcion2 []map[string]interface{}
			errInscripcion2 := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Id:2,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=Id&order=asc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion2)
			if errInscripcion2 == nil && fmt.Sprintf("%v", inscripcion2) != "[map[]]" {
				for _, inscrip2 := range inscripcion2 {
					var datoIdentif2 []map[string]interface{}
					errDatoIdentif2 := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip2["PersonaId"]), &datoIdentif2)
					if errDatoIdentif2 == nil && fmt.Sprintf("%v", datoIdentif2) != "[map[]]" {
						listado = append(listado, map[string]interface{}{
							"Id":         inscrip2["PersonaId"],
							"Aspirantes": datoIdentif2[0]["TerceroId"].(map[string]interface{})["NombreCompleto"],
						})
					} else {
						var datoIdentif2_2intento []map[string]interface{}
						errDatoIdentif2_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip2["PersonaId"]), &datoIdentif2_2intento)
						if errDatoIdentif2_2intento == nil && fmt.Sprintf("%v", datoIdentif2_2intento) != "[map[]]" {
							listado = append(listado, map[string]interface{}{
								"Id":         inscrip2["PersonaId"],
								"Aspirantes": datoIdentif2_2intento[0]["NombreCompleto"],
							})
						}
					}
				}
			}

		case 3:
			if idTelefono, ok := models.IdInfoCompTercero("10", "TELEFONO"); ok {
				var inscripcion []map[string]interface{}
				errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v&sortby=NotaFinal&order=desc&limit=0", params[id_proyecto].valor, params[id_periodo].valor), &inscripcion)
				if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
					for _, inscrip := range inscripcion {

						datoIdentTercero := map[string]interface{}{
							"nombre": "",
							"numero": "",
							"correo": "",
						}

						var datoIdentif []map[string]interface{}
						errDatoIdentif := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip["PersonaId"]), &datoIdentif)
						if errDatoIdentif == nil && fmt.Sprintf("%v", datoIdentif) != "[map[]]" {
							datoIdentTercero["nombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["NombreCompleto"]
							datoIdentTercero["numero"] = datoIdentif[0]["Numero"]
							datoIdentTercero["correo"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["UsuarioWSO2"]
						} else {
							var datoIdentif_2intento []map[string]interface{}
							errDatoIdentif_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip["PersonaId"]), &datoIdentif_2intento)
							if errDatoIdentif_2intento == nil && fmt.Sprintf("%v", datoIdentif_2intento) != "[map[]]" {
								datoIdentTercero["nombre"] = datoIdentif_2intento[0]["NombreCompleto"]
								datoIdentTercero["numero"] = ""
								datoIdentTercero["correo"] = datoIdentif_2intento[0]["UsuarioWSO2"]
							}
						}

						var enfasis map[string]interface{}
						errEnfasis := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("enfasis/%v", inscrip["EnfasisId"]), &enfasis)
						if errEnfasis != nil || enfasis["Status"] == "404" {
							enfasis = map[string]interface{}{
								"Nombre": "Por definir",
							}
						}

						var telefono []map[string]interface{}
						var telefonoPrincipal string = ""
						errTelefono := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("info_complementaria_tercero?query=TerceroId.Id:%v,InfoComplementariaId.Id:%v&sortby=Id&order=desc&fields=Dato&limit=1", inscrip["PersonaId"], idTelefono), &telefono)
						if errTelefono == nil && fmt.Sprintf("%v", telefono) != "[map[]]" {
							var telefonos map[string]interface{}
							if err := json.Unmarshal([]byte(telefono[0]["Dato"].(string)), &telefonos); err == nil {
								telefonoPrincipal = fmt.Sprintf("%.f", telefonos["principal"])
							}
						}

						ReciboInscripcion := inscrip["ReciboInscripcion"].(string)
						var recibo map[string]interface{}
						var Estado string
						if ReciboInscripcion != "0/<nil>" {
							errRecibo := request.GetJsonWSO2("http://"+beego.AppConfig.String("ConsultarReciboJbpmService")+"consulta_recibo/"+ReciboInscripcion, &recibo)
							if errRecibo == nil {
								if recibo != nil && fmt.Sprintf("%v", recibo) != "map[reciboCollection:map[]]" && fmt.Sprintf("%v", recibo) != "map[]" {
									//Fecha límite de pago extraordinario
									FechaLimite := recibo["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["fecha_extraordinario"].(string)
									EstadoRecibo := recibo["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["estado"].(string)
									PagoRecibo := recibo["reciboCollection"].(map[string]interface{})["recibo"].([]interface{})[0].(map[string]interface{})["pago"].(string)
									//Verificación si el recibo de pago se encuentra activo y pago
									if EstadoRecibo == "A" && PagoRecibo == "S" {
										Estado = "Pago"
									} else {
										//Verifica si el recibo está vencido o no
										ATiempo, err := models.VerificarFechaLimite(FechaLimite)
										if err == nil {
											if ATiempo {
												Estado = "Pendiente pago"
											} else {
												Estado = "Vencido"
											}
										} else {
											Estado = "Vencido"
										}
									}
								}
							}
						}

						listado = append(listado, map[string]interface{}{
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

		if len(listado) > 0 {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, listado)
		} else {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Error service GetListaAspirantesPor: no data found, length is 0")
		}

	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, outputErrorInfo)
	}

	c.ServeJSON()
}

// GetDependenciaPorVinculacionTercero ...
// @Title GetDependenciaPorVinculacionTercero
// @Description get DependenciaId por Vinculacion de tercero, verificando cargo
// @Param	id_tercero	path	int	true	"Id del tercero"
// @Success 200 {}
// @Failure 404 not found resource
// @router /dependencia_vinculacion_tercero/:id_tercero [get]
func (c *AdmisionController) GetDependenciaPorVinculacionTercero() {
	/*
		definition de respuestas
	*/
	successAns := map[string]interface{}{"Success": true, "Status": "200", "Message": "Query successful", "Data": nil}
	/*
		check validez de id tercero
	*/
	id_tercero_str := c.Ctx.Input.Param(":id_tercero")
	id_tercero, errId := strconv.ParseInt(id_tercero_str, 10, 64)
	if errId != nil || id_tercero <= 0 {
		if errId == nil {
			errId = fmt.Errorf("id_tercero: %d <= 0", id_tercero)
		}
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, errId.Error())
		c.ServeJSON()
		return
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
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = requestresponse.APIResponseDTO(false, http.StatusBadRequest, nil, estadoVinculacionErr.Error())
		c.ServeJSON()
		return
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

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = requestresponse.APIResponseDTO(true, http.StatusOK, successAns)
	c.ServeJSON()
}
