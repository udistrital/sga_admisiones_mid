package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/yourusername/projectname/models/requestresponse"
	"github.com/yourusername/projectname/services"
)

// ConsultaInscritosController maneja la consulta de inscritos admitidos
type ConsultaInscritosController struct {
	beego.Controller
}

// Get consulta los inscritos admitidos para un periodo y proyecto específico
func (c *ConsultaInscritosController) Get() {
	idPeriodo, _ := c.GetInt64("id_periodo")
	idProyecto, _ := c.GetInt64("id_proyecto")

	resultado := services.ConsultaInscritosAdmitidos(idPeriodo, idProyecto)

	c.Data["json"] = resultado
	c.ServeJSON()
}
