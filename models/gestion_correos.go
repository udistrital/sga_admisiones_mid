package models

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
)

func solicitudDatoIdentifGetLista(inscrip map[string]interface{}, datoIdentTercero *map[string]interface{}) {
	idTercero := inscrip["TerceroId"].(string)
	var tercero map[string]interface{}
	errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero/%v", idTercero), &tercero)
	if errTercero == nil {
		(*datoIdentTercero)["PrimerNombre"] = tercero["PrimerNombre"]
		(*datoIdentTercero)["SegundoNombre"] = tercero["SegundoNombre"]
		(*datoIdentTercero)["PrimerApellido"] = tercero["PrimerApellido"]
		(*datoIdentTercero)["SegundoApellido"] = tercero["SegundoApellido"]
		(*datoIdentTercero)["UsuarioWSO2"] = tercero["UsuarioWSO2"]
		(*datoIdentTercero)["Activo"] = tercero["Activo"]
		(*datoIdentTercero)["FechaCreacion"] = tercero["FechaCreacion"]
		(*datoIdentTercero)["FechaModificacion"] = tercero["FechaModificacion"]
	}
}
