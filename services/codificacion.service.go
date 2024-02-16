package services

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func GetAdmitidos(idPeriodo int64, idProyecto int64, periodoValor string, proyectoCodigo string) (APIResponseDTO requestresponse.APIResponse) {

	var inscripcion []map[string]interface{}
	var listado []map[string]interface{}

	//Cambair el formato de periodo valor apra comparar
	fmt.Println(periodoValor)
	if periodoValor[len(periodoValor)-1:] == "3" {
		periodoValor = strings.ReplaceAll(periodoValor, "-3", "2")
	} else {
		periodoValor = strings.ReplaceAll(periodoValor, "-1", "1")
	}

	compareCodigo := periodoValor + proyectoCodigo
	fmt.Println("http://" + beego.AppConfig.String("InscripcionService") + fmt.Sprintf("/inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,EstadoInscripcionId__Nombre:ADMITIDO&sortby=NotaFinal&order=desc&limit=0", idProyecto, idPeriodo))
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,EstadoInscripcionId__Nombre:ADMITIDO&sortby=NotaFinal&order=desc&limit=0", idProyecto, idPeriodo), &inscripcion)
	if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
		fmt.Println("Pasó de el if")
		for _, inscrip := range inscripcion {
			datoIdentTercero := map[string]interface{}{
				"PrimerNombre":    "",
				"SegundoNombre":   "",
				"PrimerApellido":  "",
				"SegundoApellido": "",
				"numero":          "",
				"codigo":          "",
			}

			var datoIdentif []map[string]interface{}
			errDatoIdentif := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId:%v", inscrip["PersonaId"]), &datoIdentif)
			if errDatoIdentif == nil && fmt.Sprintf("%v", datoIdentif) != "[map[]]" {
				datoIdentTercero["PrimerNombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["PrimerNombre"]
				datoIdentTercero["SegundoNombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["SegundoNombre"]
				datoIdentTercero["PrimerApellido"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["PrimerApellido"]
				datoIdentTercero["SegundoApellido"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["SegundoApellido"]
				datoIdentTercero["numero"] = datoIdentif[0]["Numero"]
			} else {
				var datoIdentif_2intento []map[string]interface{}
				errDatoIdentif_2intento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscrip["PersonaId"]), &datoIdentif_2intento)
				if errDatoIdentif_2intento == nil && fmt.Sprintf("%v", datoIdentif_2intento) != "[map[]]" {
					datoIdentTercero["PrimerNombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["PrimerNombre"]
					datoIdentTercero["SegundoNombre"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["SegundoNombre"]
					datoIdentTercero["PrimerApellido"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["PrimerApellido"]
					datoIdentTercero["SegundoApellido"] = datoIdentif[0]["TerceroId"].(map[string]interface{})["SegundoApellido"]
					datoIdentTercero["numero"] = ""
				}
			}

			//Definición enfasis
			var enfasis map[string]interface{}
			errEnfasis := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("enfasis/%v", inscrip["EnfasisId"]), &enfasis)
			if errEnfasis != nil || enfasis["Status"] == "404" {
				enfasis = map[string]interface{}{
					"Nombre": "Por definir",
				}
			}

			//Definición código
			var codigoIdentif []map[string]interface{}
			errCodigoIdentif := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TerceroId__Id:%v,TipoDocumentoId__Id:14", inscrip["PersonaId"]), &codigoIdentif)
			if errCodigoIdentif == nil && fmt.Sprintf("%v", datoIdentif) != "[map[]]" {
				fmt.Println(compareCodigo)
				for _, cod := range codigoIdentif {
					codigo, ok := cod["Numero"].(string)
					if(ok && codigo[0:7] == compareCodigo){
						datoIdentTercero["codigo"] = cod["Numero"]
					}else{
						datoIdentTercero["codigo"] =""
					}
					
				}

			}

			listado = append(listado, map[string]interface{}{

				"NumeroDocumento":   datoIdentTercero["numero"],
				"PrimerNombre":      datoIdentTercero["PrimerNombre"],
				"SegundoNombre":     datoIdentTercero["SegundoNombre"],
				"PrimerApellido":    datoIdentTercero["PrimerApellido"],
				"SegundoApellido":   datoIdentTercero["SegundoApellido"],
				"PuntajeFinal":      inscrip["NotaFinal"],
				"EstadoInscripcion": inscrip["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				"Enfasis":           enfasis["Nombre"],
				"codigo":            datoIdentTercero["codigo"],
			})

		}

		if len(listado) > 0 {
			APIResponseDTO = requestresponse.APIResponseDTO(true, 200, listado)
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
		}
	} else {
		if errInscripcion == nil {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil, "No data found")
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, errInscripcion.Error())
		}

	}

	return APIResponseDTO
}
