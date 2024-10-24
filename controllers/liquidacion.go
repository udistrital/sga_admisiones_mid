package controllers

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_admisiones_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
)

// LiquidacionController operations for Liquidacion
type LiquidacionController struct {
	beego.Controller
}

// URLMapping ...
func (c *LiquidacionController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("PostInformePregrado", c.PostInformePregrado)
	c.Mapping("PostInformeposgrado", c.PostInformeposgrado)
	c.Mapping("GetLiquidacion", c.GetLiquidacion)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
	//c.Mapping("GetInformeLiquidacionPregrado", c.GetInformeLiquidacionPregrado)
}

// Post ...
// @Title Create
// @Description create Liquidacion
// @Param	body		body 	models.Liquidacion	true		"body for Liquidacion content"
// @Success 201 {object} models.Liquidacion
// @Failure 403 body is empty
// @router / [post]
func (c *LiquidacionController) Post() {
	defer errorhandler.HandlePanic(&c.Controller)
	data := c.Ctx.Input.RequestBody
	respuesta := services.CrearLiquidacion(data)
	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// GetLiquidacion ...
// @Title GetLiquidacion
// @Description get Liquidacion by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Liquidacion
// @Failure 403 :id is empty
// @router / [get]
func (c *LiquidacionController) GetLiquidacion() {
	defer errorhandler.HandlePanic(&c.Controller)
	fmt.Println("Llegó")
	//Id del periodo
	idPeriodo, errPeriodo := c.GetInt64("id_periodo")
	//Id del proyecto
	idProyecto, errProyecto := c.GetInt64("id_proyecto")

	if errPeriodo == nil && errProyecto == nil {
		respuesta := services.ListarLiquidacionEstudiantes(idPeriodo, idProyecto)

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
// @Description get Liquidacion
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.Liquidacion
// @Failure 403
// @router /lista [get]
func (c *LiquidacionController) GetAll() {
	fmt.Println("Llegó")
	defer errorhandler.HandlePanic(&c.Controller)
	respuesta := services.GetAllLiquidaciones()
	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()
}

// Put ...
// @Title Put
// @Description update the Liquidacion
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	models.Liquidacion	true		"body for Liquidacion content"
// @Success 200 {object} models.Liquidacion
// @Failure 403 :id is not int
// @router /:id [put]
func (c *LiquidacionController) Put() {

}

// Delete ...
// @Title Delete
// @Description delete the Liquidacion
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *LiquidacionController) Delete() {

}

// PostInformePregrado ...
// @Title Create
// @Description create InformeLiquidacionPregrado
// @Param	body		body 	models.Liquidacion	true		"body for Liquidacion content"
// @Success 201 {object} models.Liquidacion
// @Failure 403 body is empty
// @router /informe/pregrado [post]
func (c *LiquidacionController) PostInformePregrado() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody
	respuesta := services.InformeLiquidacionPregrado(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()

}

// PostInformeposgrado ...
// @Title Create
// @Description create InformeLiquidacionPosgrado
// @Param	body		body 	models.Liquidacion	true		"body for Liquidacion content"
// @Success 201 {object} models.Liquidacion
// @Failure 403 body is empty
// @router /informe/posgrado [post]
func (c *LiquidacionController) PostInformeposgrado() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody
	respuesta := services.InformeLiquidacionPosgrado(data)

	c.Ctx.Output.SetStatus(respuesta.Status)
	c.Data["json"] = respuesta
	c.ServeJSON()

}
