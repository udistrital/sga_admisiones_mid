package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_mid_admisiones/models"
	"github.com/udistrital/sga_mid_admisiones/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/request"
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
// @Failure 403 body is empty
// @router /calcular_nota [put]
func (c *AdmisionController) PutNotaFinalAspirantes() {
	defer errorhandler.HandlePanic(&c.Controller)

	var Evaluacion map[string]interface{}
	var Inscripcion []map[string]interface{}
	var DetalleEvaluacion []map[string]interface{}
	var NotaFinal float64
	var InscripcionPut map[string]interface{}
	var respuesta []map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var alerta models.Alert
	var errorGetAll bool
	alertas := append([]interface{}{})

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &Evaluacion); err == nil {
		IdPersona := Evaluacion["IdPersona"].([]interface{})
		PeriodoId := fmt.Sprintf("%v", Evaluacion["IdPeriodo"])
		ProgramaAcademicoId := fmt.Sprintf("%v", Evaluacion["IdPrograma"])
		respuesta = make([]map[string]interface{}, len(IdPersona))
		for i := 0; i < len(IdPersona); i++ {
			PersonaId := fmt.Sprintf("%v", IdPersona[i].(map[string]interface{})["Id"])

			//GET a Inscripción para obtener el ID
			if resp := services.SolicitudIdPut(PersonaId, PeriodoId, ProgramaAcademicoId, &Inscripcion, &DetalleEvaluacion, NotaFinal, InscripcionPut, &respuesta, i, &alerta, &alertas, &errorGetAll); resp != nil {
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = resp
			}
		}
		resultado["Response"] = respuesta
	} else {
		c.Ctx.Output.SetStatus(404)
		services.ManejoError(&alerta, &alertas, &errorGetAll, "", err)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		services.ManejoExito(&alertas, &alerta, resultado)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
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
// @Failure 403 body is empty
// @router /consultar_evaluacion/:id_programa/:id_periodo/:id_requisito [get]
func (c *AdmisionController) GetEvaluacionAspirantes() {
	defer errorhandler.HandlePanic(&c.Controller)

	id_periodo := c.Ctx.Input.Param(":id_periodo")
	id_programa := c.Ctx.Input.Param(":id_programa")
	id_requisito := c.Ctx.Input.Param(":id_requisito")
	var DetalleEvaluacion []map[string]interface{}
	var DetalleEspecificoJSON []map[string]interface{}
	var Inscripcion map[string]interface{}
	var Terceros map[string]interface{}
	var resultado map[string]interface{}
	resultado = make(map[string]interface{})
	var alerta models.Alert
	var errorGetAll bool
	alertas := append([]interface{}{})

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
					services.IterarEvaluacion(Evaluacion, &respuestaAux)

					//GET a la tabla de inscripcion para saber el id del inscrito
					if resp := services.SolicitudInscripcionGetEvApspirantes(evaluacion, &Inscripcion, &Terceros, &respuestaAux, &errorGetAll, &alerta, &alertas); resp != nil {
						c.Ctx.Output.SetStatus(404)
						c.Data["json"] = resp
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
			c.Ctx.Output.SetStatus(404)
			services.ManejoError(&alerta, &alertas, &errorGetAll, "No data found")
			c.Data["json"] = map[string]interface{}{"Response": alerta}
		}

	} else {
		c.Ctx.Output.SetStatus(404)
		services.ManejoError(&alerta, &alertas, &errorGetAll, "", errDetalleEvaluacion)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		services.ManejoExito(&alertas, &alerta, resultado)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
	}

	c.ServeJSON()
}

// PostEvaluacionAspirantes ...
// @Title PostEvaluacionAspirantes
// @Description Agregar la evaluacion de los aspirantes de acuerdo a los criterios
// @Param   body        body    {}  true        "body Agregar evaluacion aspirantes content"
// @Success 200 {}
// @Failure 403 body is empty
// @router /registrar_evaluacion [post]
func (c *AdmisionController) PostEvaluacionAspirantes() {
	defer errorhandler.HandlePanic(&c.Controller)

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

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &Evaluacion); err == nil {
		AspirantesData := Evaluacion["Aspirantes"].([]interface{})
		ProgramaAcademicoId := Evaluacion["ProgramaId"]
		PeriodoId := Evaluacion["PeriodoId"]
		CriterioId := Evaluacion["CriterioId"]
		respuesta = make([]map[string]interface{}, len(AspirantesData))
		//GET para obtener el porcentaje general, especifico (si lo hay)
		if resp := services.SolicitudRequisitoPostEvaluacion(ProgramaAcademicoId, PeriodoId, &Inscripciones, &Ponderado, &DetalleCalificacion, Evaluacion, AspirantesData, &respuesta, Requisito, DetalleEvaluacion, &errorGetAll, &alertas, &alerta, CriterioId); resp != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = resp
		}

		resultado["Evaluacion"] = respuesta
	} else {
		c.Ctx.Output.SetStatus(404)
		services.ManejoError(&alerta, &alertas, &errorGetAll, "", err)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
	}

	if !errorGetAll {
		c.Ctx.Output.SetStatus(200)
		services.ManejoExito(&alertas, &alerta, resultado)
		c.Data["json"] = map[string]interface{}{"Response": alerta}
	}

	c.ServeJSON()
}

// PostCriterioIcfes ...
// @Title PostCriterioIcfes
// @Description Agregar CriterioIcfes
// @Param   body        body    {}  true        "body Agregar CriterioIcfes content"
// @Success 200 {}
// @Failure 403 body is empty
// @router / [post]
func (c *AdmisionController) PostCriterioIcfes() {
	defer errorhandler.HandlePanic(&c.Controller)

	var CriterioIcfes map[string]interface{}
	var alerta models.Alert
	alertas := append([]interface{}{"Response:"})
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

			var criterio_existente []map[string]interface{}
			errCriterioExistente := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"requisito_programa_academico/?query=ProgramaAcademicoId:"+fmt.Sprintf("%.f", criterioProyectos["Id"].(float64)), &criterio_existente)
			if errCriterioExistente == nil && fmt.Sprintf("%v", criterio_existente[0]) != "map[]" {
				if criterio_existente[0]["Status"] != 404 {
					services.ManejoCriterioCriterioIcfes(&criterioProyecto, CriterioIcfes, requestBod, criterioProyectos, i, &alertas, &alerta, 1, &criterio_existente)
				} else 
					c.Ctx.Output.SetStatus(404)
					if criterio_existente[0]["Message"] == "Not found resource" {
						c.Data["json"] = nil
					} else {
						logs.Error(criterio_existente)
						c.Data["system"] = errCriterioExistente
						c.Abort("404")
					}
				}
			} else {
				services.ManejoCriterioCriterioIcfes(&criterioProyecto, CriterioIcfes, requestBod, criterioProyectos, i, &alertas, &alerta, 2, &criterio_existente)
			}
		}
		c.Ctx.Output.SetStatus(200)
		alertas = append(alertas, criterioProyecto)

	} else {
		c.Ctx.Output.SetStatus(404)
		services.ManejoErrorSinGetAll(&alerta, &alertas, "", err)
	}
	c.Data["json"] = alerta
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
	defer errorhandler.HandlePanic(&c.Controller)

	var consulta map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consulta); err == nil {

		var resultado_puntaje []map[string]interface{}
		errPuntaje := request.GetJson("http://"+beego.AppConfig.String("EvaluacionInscripcionService")+"detalle_evaluacion/?query=RequisitoProgramaAcademicoId.ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",RequisitoProgramaAcademicoId.PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_puntaje)

		if errPuntaje == nil && fmt.Sprintf("%v", resultado_puntaje[0]) != "map[]" {
			if resultado_puntaje[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_puntaje)
				for i, resultado_tem := range resultado_puntaje {
					infoSystem, infoJson, exito := services.PeticionResultadoInscripcionGetPuntaje(resultado_tem, &resultado_puntaje, i)

					if !exito {
						if infoSystem != nil {
							c.Data["system"] = infoSystem
							c.Abort("404")
						} else {
							c.Data["json"] = infoJson
						}
					}

					c.Data["json"] = resultado_puntaje
				}
				c.Ctx.Output.SetStatus(200)
			} else {
				c.Ctx.Output.SetStatus(404)
				if resultado_puntaje[0]["Message"] == "Not found resource" {
					c.Data["json"] = nil
				} else {
					logs.Error(resultado_puntaje)
					c.Data["system"] = errPuntaje
					c.Abort("404")
				}
			}
		} else {
			c.Ctx.Output.SetStatus(404)
			logs.Error(resultado_puntaje)
			c.Data["system"] = errPuntaje
			c.Abort("404")
		}
	} else {
		c.Ctx.Output.SetStatus(400)
		logs.Error(err)
		c.Data["system"] = err
		c.Abort("400")
	}
	c.ServeJSON()
}

// PostCuposAdmision ...
// @Title PostCuposAdmision
// @Description Agregar PostCuposAdmision
// @Param   body        body    {}  true        "body Agregar PostCuposAdmision content"
// @Success 200 {}
// @Failure 403 body is empty
// @router /postcupos [post]
func (c *AdmisionController) PostCuposAdmision() {
	defer errorhandler.HandlePanic(&c.Controller)

	var CuposAdmision map[string]interface{}

	alertas := []interface{}{"Response:"}
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
				resultado := services.SolicituVerificacionCuposAdmision(cupoProyectos, CuposAdmision, &CuposProyectos, requestBod, i); resultado != nil {
					c.Ctx.Output.SetStatus(400)
					c.Data["json"] = resultado
					break
				}
			}

			alertas = append(alertas, CuposProyectos)
			c.Ctx.Output.SetStatus(200)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "200", "Message": "Request successful", "Data": alertas}
		} else {
			c.Ctx.Output.SetStatus(403)
			c.Data["json"] = map[string]interface{}{"Success": false, "Status": "403", "Message": "Body is empty", "Data": nil}
		}
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]interface{}{"Success": false, "Status": "400", "Message": err.Error(), "Data": nil}
	}

	c.ServeJSON()
}

// CambioEstadoAspiranteByPeriodoByProyecto ...
// @Title CambioEstadoAspiranteByPeriodoByProyecto
// @Description post cambioestadoaspirante by id_periodo and id_proyecto
// @Param   body        body    {}  true        "body for  post cambio estadocontent"
// @Success 200 {}
// @Failure 403 body is empty
// @router /cambioestado [post]
func (c *AdmisionController) CambioEstadoAspiranteByPeriodoByProyecto() {
	defer errorhandler.HandlePanic(&c.Controller)

	var consultaestado map[string]interface{}
	EstadoActulizado := "Estados Actualizados"
	var alerta models.Alert
	alertas := append([]interface{}{"Response:"})

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consultaestado); err == nil {
		Id_periodo := consultaestado["Periodo"].(map[string]interface{})["Id"]
		for _, proyectotemp := range consultaestado["Proyectos"].([]interface{}) {
			EstadoProyectos := proyectotemp.(map[string]interface{})

			infoSystem, infoJson, exito := services.PeticionCuposCambioEstado(EstadoProyectos, Id_periodo)

			if !exito {
				c.Ctx.Output.SetStatus(404)
				if infoSystem != nil {
					c.Data["system"] = infoSystem
					c.Abort("404")
				} else {
					c.Data["json"] = infoJson
				}
			}
		}
		c.Ctx.Output.SetStatus(200)
		alertas = append(alertas, EstadoActulizado)

	} else {
		logs.Error(err)
		//c.Data["development"] = map[string]interface{}{"Code": "400", "Body": err.Error(), "Type": "error"}
		c.Ctx.Output.SetStatus(400)
		c.Data["system"] = err
		c.Abort("400")
	}

	alerta.Body = alertas
	c.Data["json"] = alerta
	c.ServeJSON()
}

// GetAspirantesByPeriodoByProyecto ...
// @Title GetAspirantesByPeriodoByProyecto
// @Description get Aspirantes by id_periodo and id_proyecto
// @Param	body		body 	{}	true		"body for Get Aspirantes content"
// @Success 201 {int}
// @Failure 400 the request contains incorrect syntax
// @router /consulta_aspirantes [post]
func (c *AdmisionController) GetAspirantesByPeriodoByProyecto() {
	defer errorhandler.HandlePanic(&c.Controller)

	var consulta map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &consulta); err == nil {

		var resultado_aspirante []map[string]interface{}
		errAspirante := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+"inscripcion/?query=ProgramaAcademicoId:"+fmt.Sprintf("%v", consulta["Id_proyecto"])+",PeriodoId:"+fmt.Sprintf("%v", consulta["Id_periodo"]), &resultado_aspirante)
		if errAspirante == nil && fmt.Sprintf("%v", resultado_aspirante[0]) != "map[]" {
			if resultado_aspirante[0]["Status"] != 404 {
				// formatdata.JsonPrint(resultado_aspirante)
				for i, resultado_tem := range resultado_aspirante {
					infoSystem, infoJson, exito := services.PeticionNotaGetAspirante(resultado_tem, &resultado_aspirante, i)

					if !exito {
						c.Ctx.Output.SetStatus(404)
						if infoSystem != nil {
							c.Data["system"] = infoSystem
							c.Abort("404")
						} else {
							c.Data["json"] = infoJson
						}
					}

					c.Data["json"] = resultado_aspirante
				}
				c.Ctx.Output.SetStatus(200)
			} else {
				c.Ctx.Output.SetStatus(404)
				if resultado_aspirante[0]["Message"] == "Not found resource" {
					c.Data["json"] = nil
				} else {
					logs.Error(resultado_aspirante)
					//c.Data["development"] = map[string]interface{}{"Code": "404", "Body": err.Error(), "Type": "error"}
					c.Data["system"] = errAspirante
					c.Abort("404")
				}
			}
		} else {
			c.Ctx.Output.SetStatus(404)
			logs.Error(resultado_aspirante)
			//c.Data["development"] = map[string]interface{}{"Code": "404", "Body": err.Error(), "Type": "error"}
			c.Data["system"] = errAspirante
			c.Abort("404")

		}

	} else {
		c.Ctx.Output.SetStatus(400)
		logs.Error(err)
		//c.Data["development"] = map[string]interface{}{"Code": "400", "Body": err.Error(), "Type": "error"}
		c.Data["system"] = err
		c.Abort("400")
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
	defer errorhandler.HandlePanic(&c.Controller)

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
		services.ManejoCasosGetLista(params[tipo_lista].valor, params[id_periodo].valor, params[id_proyecto].valor, &listado)

		if len(listado) > 0 {
			c.Ctx.Output.SetStatus(200)
			c.Data["json"] = map[string]interface{}{"Success": true, "Status": "200", "Message": "Query successful", "Data": listado}
		} else {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]interface{}{"Success": false, "Status": "404", "Message": "Error service GetListaAspirantesPor: no data found, length is 0"}
		}

	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = outputErrorInfo
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
	defer errorhandler.HandlePanic(&c.Controller)

	/*
		definition de respuestas
	*/
	failureAsn := map[string]interface{}{"Success": false, "Status": "404",
		"Message": "Error service GetDependenciaPorVinculacionTercero: The request contains an incorrect parameter or no record exist", "Data": nil}
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
		logs.Error(errId.Error())
		c.Ctx.Output.SetStatus(404)
		failureAsn["Data"] = errId.Error()
		c.Data["json"] = failureAsn
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
		logs.Error(estadoVinculacionErr.Error())
		c.Ctx.Output.SetStatus(404)
		failureAsn["Data"] = estadoVinculacionErr.Error()
		c.Data["json"] = failureAsn
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
	c.Ctx.Output.SetStatus(200)
	successAns["Data"] = map[string]interface{}{
		"DependenciaId": dependencias,
	}
	c.Data["json"] = successAns
	c.ServeJSON()
}
