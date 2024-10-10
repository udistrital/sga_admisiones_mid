package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

// ReportesController operations for Reportes
type ReportesController struct {
	beego.Controller
}

// URLMapping ...
func (c *ReportesController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
	c.Mapping("ReporteCaracterizacion", c.ReporteCaracterizacion)
}

// PostReportes ...
// @Title GenerarReportes
// @Description Crear reportes dinamicos
// @Param   body        body    {}  true        "body con la información de las filas a eliminar el proeycto y el periodo"
// @Success 201 {object} models.Reportes
// @Failure 403 body is empty
// @router / [post]
func (c *ReportesController) Post() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	if data != nil {
		respuesta := services.ReporteDinamico(data)
		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
		c.ServeJSON()

	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Datos erroneos")
		c.ServeJSON()
	}

}

// GetCodificaciones ...
// @Title GetAll
// @Description get Reportes
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Success 200 {object} models.Reportes
// @Failure 403
// @router / [get]
func (c *ReportesController) GetAll() {
	defer errorhandler.HandlePanic(&c.Controller)
	//Id del periodo
	idPeriodo, errPeriodo := c.GetInt64("id_periodo")
	//Id del proyecto
	idProyecto, errProyecto := c.GetInt64("id_proyecto")

	if errPeriodo == nil && errProyecto == nil {
		respuesta := services.GenerarReporteCodigos(idPeriodo, idProyecto)

		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = "Invalid data"
		c.ServeJSON()
	}
}

// Put ...
// @Title Put
// @Description update the Reportes
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	models.Reportes	true		"body for Reportes content"
// @Success 200 {object} models.Reportes
// @Failure 403 :id is not int
// @router /:id [put]
func (c *ReportesController) Put() {

}

// Delete ...
// @Title Delete
// @Description delete the Reportes
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *ReportesController) Delete() {

}

// ReporteCaracterizacion ...
// @Title ReporteCaracterizacion
// @Description Reportes de Caraterización
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Success 200 {object} models.Reportes
// @Failure 403
// @router /reporte-caracterizacion [get]
func (c *ReportesController) ReporteCaracterizacion() {
	defer errorhandler.HandlePanic(&c.Controller)

	// Id del periodo
	idPeriodo, errPeriodo := c.GetInt64("id_periodo")
	// Id del proyecto
	idProyecto, errProyecto := c.GetInt64("id_proyecto")

	// Datos de ejemplo del JSON `data_proceso`
	dataProceso := []map[string]interface{}{
		{"Id": 1, "Nombre": "Inscripciones"},
		{"Id": 2, "Nombre": "Admisiones"},
	}

	// Validación de que `id_proyecto` existe en `data_proceso`
	proyectoValido := false
	for _, proceso := range dataProceso {
		if proceso["Id"] == idProyecto {
			proyectoValido = true
			break
		}
	}

	if errPeriodo == nil && errProyecto == nil && proyectoValido {
		respuesta := services.ReporteCaracterizacion(idPeriodo, idProyecto)

		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = "Invalid data"
	}
	c.ServeJSON()
}
