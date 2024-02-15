package services

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func GetAdmitidos(idPeriodo int64, idProyecto int64) (APIResponseDTO requestresponse.APIResponse) {

	var inscripcion []map[string]interface{}
	var listado []map[string]interface{}
	fmt.Println("http://" + beego.AppConfig.String("InscripcionService") + fmt.Sprintf("/inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,EstadoInscripcionId__Nombre:ADMITIDO&sortby=NotaFinal&order=desc&limit=0", idProyecto, idPeriodo))
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,EstadoInscripcionId__Nombre:ADMITIDO&sortby=NotaFinal&order=desc&limit=0", idProyecto, idPeriodo), &inscripcion)
	if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {

		for _, inscrip := range inscripcion {
			datoIdentTercero := map[string]interface{}{
				"PrimerNombre":    "",
				"SegundoNombre":   "",
				"PrimerApellido":  "",
				"SegundoApellido": "",
				"numero":          "",
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

			var enfasis map[string]interface{}
			errEnfasis := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("enfasis/%v", inscrip["EnfasisId"]), &enfasis)
			if errEnfasis != nil || enfasis["Status"] == "404" {
				enfasis = map[string]interface{}{
					"Nombre": "Por definir",
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
			})

		}

		if len(listado) > 0 {
			APIResponseDTO = requestresponse.APIResponseDTO(true, 200, listado)
		} else {
			APIResponseDTO = requestresponse.APIResponseDTO(false, 404, nil)
		}
	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, errInscripcion.Error())
	}

	return APIResponseDTO
}
