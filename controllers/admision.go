package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
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
	c.Mapping("GetListaAspirantesPorProyectosActivos", c.GetListaAspirantesDeProyectosActivos)
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
	data := c.Ctx.Input.RequestBody

	respuesta := services.SolicitudIdPut(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
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
// @router /evaluacion/:id_programa/:id_periodo/:id_requisito [get]
func (c *AdmisionController) GetEvaluacionAspirantes() {
	defer errorhandler.HandlePanic(&c.Controller)

	id_periodo := c.Ctx.Input.Param(":id_periodo")
	id_programa := c.Ctx.Input.Param(":id_programa")
	id_requisito := c.Ctx.Input.Param(":id_requisito")

	respuesta := services.IterarEvaluacion(id_periodo, id_programa, id_requisito)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// PostEvaluacionAspirantes ...
// @Title PostEvaluacionAspirantes
// @Description Agregar la evaluacion de los aspirantes de acuerdo a los criterios
// @Param   body        body    {}  true        "body Agregar evaluacion aspirantes content"
// @Success 200 {}
// @Failure 403 body is empty
// @router /evaluacion [post]
func (c *AdmisionController) PostEvaluacionAspirantes() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	respuesta := services.RegistratEvaluaciones(data)
	fmt.Println("respuestaaa")
	fmt.Println(respuesta)
	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
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

	data := c.Ctx.Input.RequestBody

	respuesta := services.CriteriosIcfesPost(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// ConsultarPuntajeTotalByPeriodoByProyecto ...
// @Title GetPuntajeTotalByPeriodoByProyecto
// @Description get PuntajeTotalCriteio by id_periodo and id_proyecto
// @Param	body		body 	{}	true		"body for Get Puntaje total content"
// @Success 201 {int}
// @Failure 400 the request contains incorrect syntax
// @router /puntaje [post]
func (c *AdmisionController) GetPuntajeTotalByPeriodoByProyecto() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	respuesta := services.PuntajeTotal(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// PostCuposAdmision ...
// @Title PostCuposAdmision
// @Description Agregar PostCuposAdmision
// @Param   body        body    {}  true        "body Agregar PostCuposAdmision content"
// @Success 200 {}
// @Failure 403 body is empty
// @router /cupos [post]
func (c *AdmisionController) PostCuposAdmision() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	respuesta := services.CuposAdmision(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// CambioEstadoAspiranteByPeriodoByProyecto ...
// @Title CambioEstadoAspiranteByPeriodoByProyecto
// @Description post cambioestadoaspirante by id_periodo and id_proyecto
// @Param   body        body    {}  true        "body for  post cambio estadocontent"
// @Success 200 {}
// @Failure 403 body is empty
// @router /estado [post]
func (c *AdmisionController) CambioEstadoAspiranteByPeriodoByProyecto() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	respuesta := services.CambioEstadoAspirante(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// GetAspirantesByPeriodoByProyecto ...
// @Title GetAspirantesByPeriodoByProyecto
// @Description get Aspirantes by id_periodo and id_proyecto
// @Param	body		body 	{}	true		"body for Get Aspirantes content"
// @Success 201 {int}
// @Failure 400 the request contains incorrect syntax
// @router /aspirantes [post]
func (c *AdmisionController) GetAspirantesByPeriodoByProyecto() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	respuesta := services.ConsultaAspirantes(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// GetListaAspirantesPor ...
// @Title GetListaAspirantesPor
// @Description get Lista estados aspirantes by id_periodo id_nivel id_proyecto and tipo_lista
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Param	tipo_lista		query 	string	true		"tipo de lista"
// @Success 200 {}
// @Failure 404 not found resource
// @router /aspirantespor [get]
func (c *AdmisionController) GetListaAspirantesPor() {

	defer errorhandler.HandlePanic(&c.Controller)

	idPeriodo, okPeriodo := c.GetInt64("id_periodo")
	idProyecto, okProyecto := c.GetInt64("id_proyecto")
	idLista, okLista := c.GetInt64("tipo_lista")

	if okLista != nil || okProyecto != nil || okPeriodo != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "invalid params")
	} else {

		respuesta := services.ListaAspirantes(idPeriodo, idProyecto, idLista)
		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
	}
	c.ServeJSON()
}

// GetListaAspirantesDeProyectosActivos ...
// @Title GetListaAspirantesPorProyectosActivos
// @Description get Lista estados aspirantes de proyectos activos
// @Param	id-nivel	query	string	false	"Se recibe parametro Id de el nivel"
// @Param	id-periodo	query	string	false	"Se recibe parametro Id de el Periodo"
// @Param	tipo-lista	query 	string	false	"Se recibe parametro Id tipo de lista"
// @Success 200 {}
// @Failure 404 not found resource
// @router /aspirantes-de-proyectos-activos [get]
func (c *AdmisionController) GetListaAspirantesDeProyectosActivos() {

	defer errorhandler.HandlePanic(&c.Controller)

	var idPeriodo string
	var idNivel string
	var tipoLista string

	// Id de el periodo
	if v := c.GetString("id-periodo"); v != "" {
		idPeriodo = v
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Parametro id periodo vacío")
	}

	// Id de el nivel
	if v := c.GetString("id-nivel"); v != "" {
		idNivel = v
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Parametro id nivel vacío")
	}

	// Id de el nivel
	if v := c.GetString("tipo-lista"); v != "" {
		tipoLista = v
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Parametro tipo lista vacío")
	}

	resultado, err := services.GetAspirantesDeProyectosActivos(idNivel, idPeriodo, tipoLista)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 200, resultado)
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, err.Error())
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

	id_tercero_str := c.Ctx.Input.Param(":id_tercero")

	respuesta := services.DependenciaPorVinculacion(id_tercero_str)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}
