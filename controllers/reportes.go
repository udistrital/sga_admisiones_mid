package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
)

// ReportesController operations for Reportes
type ReportesController struct {
	beego.Controller
}

// URLMapping ...
func (c *ReportesController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
}

// Post ...
// @Title Create
// @Description create Reportes
// @Param	body		body 	models.Reportes	true		"body for Reportes content"
// @Success 201 {object} models.Reportes
// @Failure 403 body is empty
// @router / [post]
func (c *ReportesController) Post() {

}

// GetOne ...
// @Title GetOne
// @Description get Reportes by idPeriodo & idProyecto
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Success 200 {object} models.Reportes
// @Failure 403 :id is empty
// @router /:id [get]
func (c *ReportesController) GetOne() {
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

// GetAll ...
// @Title GetAll
// @Description get Reportes
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.Reportes
// @Failure 403
// @router / [get]
func (c *ReportesController) GetAll() {

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
