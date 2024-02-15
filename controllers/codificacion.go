package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
)

// CodificacionController operations for Codificacion
type CodificacionController struct {
	beego.Controller
}

// URLMapping ...
func (c *CodificacionController) URLMapping() {
	c.Mapping("GetOne", c.GetAdmitidos)
}

// GetAdmitidos ...
// @Title GetAdmitidos
// @Description get admitidos por id de proyecto y periodo
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Failure 403 :id_periodo is empty
// @Failure 403 :id_proyecto is empty
// @Success 200 {}
// @Failure 404 not found resource
// @router /getAdmitidos/:id_periodo/:id_proyecto [get]
func (c *CodificacionController) GetAdmitidos() {

	defer errorhandler.HandlePanic(&c.Controller)

	//Id del periodo
	idPeriodo, errPeriodo := c.GetInt64("id_periodo")
	//Id del proyecto
	idProyecto, errProyecto := c.GetInt64("id_proyecto")

	if (errPeriodo == nil && errProyecto == nil) {

		respuesta := services.GetAdmitidos(idPeriodo, idProyecto)
	
		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
		c.ServeJSON()
	}else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = "Invalid data"
		c.ServeJSON()
	}


}

