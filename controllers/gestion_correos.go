package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
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
	defer errorhandler.HandlePanic(&c.Controller)

	idPeriodo, _ := c.GetInt64("id_periodo")

	if idPeriodo <= 0 {
		resultado := requestresponse.APIResponseDTO(false, 403, "Id periodo incorrecto")
		c.Ctx.Output.SetStatus(resultado.Status)
		c.Data["json"] = resultado
	} else {
		resultado := services.SugerenciaCorreosUD(idPeriodo)
		c.Ctx.Output.SetStatus(resultado.Status)
		c.Data["json"] = resultado
	}

	c.ServeJSON()
}