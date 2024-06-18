package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
)

// GestionCorreosController operations for GestionCorreos
type GestionCorreosController struct {
	beego.Controller
}

// URLMapping ...
func (c *GestionCorreosController) URLMapping() {
	c.Mapping("SugerenciaCorreoInstitucional", c.SugerenciaCorreoInstitucional)
}

// SugerenciaCorreoInstitucional ...
// @Title SugerenciaCorreoInstitucional
// @Description Endpoint para sugerencias de correos institucional sin homonimo
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Failure 403 :id_periodo is empty
// @Success 200 {}
// @Failure 404 not found resource
// @router /correo-sugerido [get]
func (c *GestionCorreosController) SugerenciaCorreoInstitucional() {
	idPeriodo, _ := c.GetInt64("id_periodo")
	idProyecto, _ := c.GetInt64("id_proyecto")

	resultado := services.ConsultaInscritosAdmitidos(idPeriodo, idProyecto)

	c.Data["json"] = resultado
	c.ServeJSON()
}
