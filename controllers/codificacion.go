package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

// CodificacionController operations for Codificacion
type CodificacionController struct {
	beego.Controller
}

// URLMapping ...
func (c *CodificacionController) URLMapping() {
	c.Mapping("GetAdmitidos", c.GetAdmitidos)
	c.Mapping("GenerarCodigo", c.GenerarCodigo)
	c.Mapping("GuardarCodigo", c.GuardarCodigo)
}

// GetAdmitidos ...
// @Title GetAdmitidos
// @Description get admitidos por id de proyecto y periodo
// @Param	id_periodo		query 	int	true		"Id del periodo"
// @Param	id_proyecto		query 	int	true		"Id del proyecto"
// @Param	valor_periodo		query 	string	true		"Valor del periodo"
// @Param	codigo_proyecto		query 	string	true		"codigo del proyecto"
// @Failure 403 :id_periodo is empty
// @Failure 403 :id_proyecto is empty
// @Success 200 {}
// @Failure 404 not found resource
// @router /admitidos/ [get]
func (c *CodificacionController) GetAdmitidos() {

	defer errorhandler.HandlePanic(&c.Controller)

	//Id del periodo
	idPeriodo, errPeriodo := c.GetInt64("id_periodo")
	//Id del proyecto
	idProyecto, errProyecto := c.GetInt64("id_proyecto")
	//Id del periodo
	valorPeriodo := c.GetString("valor_periodo")
	//Id del proyecto
	codigoProyecto := c.GetString("codigo_proyecto")

	if errPeriodo == nil && errProyecto == nil {

		respuesta := services.GetAdmitidos(idPeriodo, idProyecto, valorPeriodo, codigoProyecto)

		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = "Invalid data"
		c.ServeJSON()
	}

}

// PostGenerarCodigos ...
// @Title PostGenerarCodigos
// @Description Generar códigos
// @Param   body        body    {}  true        "body para guardar código"
// @Param	tipo_sort		query 	int	true		"Id del sort 1, 2 o 3"
// @Success 200 {}
// @Failure 403 body is empty
// @router /codigos/ [post]
func (c *CodificacionController) GenerarCodigo() {

	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	//Id del proyecto
	sortTipo, errTipo := c.GetInt64("tipo_sort")

	if errTipo == nil {
		respuesta := services.GenerarCodificacion(data, sortTipo)
		c.Ctx.Output.SetStatus(respuesta.Status)

		c.Data["json"] = respuesta
		c.ServeJSON()

	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Datos erroneos")
		c.ServeJSON()
	}

}

// PostGuardarCodigos ...
// @Title PostGuardarCodigos
// @Description Guardar códigos
// @Param   body        body    {}  true        "body para guardar código"
// @Success 200 {}
// @Failure 403 body is empty
// @router /codigos-periodo/ [post]
func (c *CodificacionController) GuardarCodigo() {

	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	if data != nil {
		respuesta := services.GuardarCodificacion(data)
		c.Ctx.Output.SetStatus(respuesta.Status)
		c.Data["json"] = respuesta
		c.ServeJSON()

	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Datos erroneos")
		c.ServeJSON()
	}

}
