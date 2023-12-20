package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
)

func IdInfoCompTercero(grupo string, codAbrev string) (Id string, ok bool) {
	var resp []map[string]interface{}
	errResp := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+"info_complementaria?query=GrupoInfoComplementariaId__Id:"+grupo+",CodigoAbreviacion:"+codAbrev+"&fields=Id", &resp)
	if errResp == nil && fmt.Sprintf("%v", resp) != "[map[]]" {
		Id = fmt.Sprintf("%v", resp[0]["Id"].(float64))
		ok = true
	} else {
		Id = "0"
		ok = false
	}
	return Id, ok
}
