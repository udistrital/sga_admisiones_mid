package services

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/time_bogota"
)

func GetAdmitidos(idPeriodo int64, idProyecto int64, periodoValor string, proyectoCodigo string) (APIResponseDTO requestresponse.APIResponse) {

	var inscripcion []map[string]interface{}
	var listado []map[string]interface{}

	//Cambair el formato de periodo valor para comparar
	if  periodoValor[len(periodoValor)-1:] == "3" {
		periodoValor = strings.ReplaceAll(periodoValor, "-3", "2")
	} else {
		periodoValor = strings.ReplaceAll(periodoValor, "-", "")
	}

	compareCodigo := fmt.Sprintf("%v%v", periodoValor, proyectoCodigo)
	compareCodigo = strings.ReplaceAll(compareCodigo, "\"", "")
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v,EstadoInscripcionId__Nombre:ADMITIDO&sortby=NotaFinal&order=desc&limit=0", idProyecto, idPeriodo), &inscripcion)
	if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
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
				for _, cod := range codigoIdentif {
					codigo, ok := cod["Numero"].(string)
					if ok && codigo[0:8] == compareCodigo {
						datoIdentTercero["codigo"] = codigo
					} else {
						datoIdentTercero["codigo"] = ""
					}

				}

			}

			listado = append(listado, map[string]interface{}{
				"InscripcionId":     inscrip["Id"],
				"TerceroId":         inscrip["PersonaId"],
				"NumeroDocumento":   datoIdentTercero["numero"],
				"PrimerNombre":      datoIdentTercero["PrimerNombre"],
				"SegundoNombre":     datoIdentTercero["SegundoNombre"],
				"PrimerApellido":    datoIdentTercero["PrimerApellido"],
				"SegundoApellido":   datoIdentTercero["SegundoApellido"],
				"PuntajeFinal":      inscrip["NotaFinal"],
				"EstadoInscripcion": inscrip["EstadoInscripcionId"].(map[string]interface{})["Nombre"],
				"Enfasis":           enfasis["Nombre"],
				"codigo":            datoIdentTercero["codigo"],
				"codigoIndice":      compareCodigo,
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

func GenerarCodificacion(data []byte, tipo int64) (APIResponseDTO requestresponse.APIResponse) {
	var estudiantes []map[string]interface{}

	if err := json.Unmarshal(data, &estudiantes); err == nil {
		switch tipo {
		case 1:
			// Ordenar el slice por el campo "PrimerApellido"
			sort.Slice(estudiantes, func(i, j int) bool {
				// Convertir los valores del campo "nombre" a strings
				nombreI := estudiantes[i]["PrimerApellido"].(string)
				nombreJ := estudiantes[j]["PrimerApellido"].(string)
				// Comparar los apellidos lexicográficamente
				return nombreI < nombreJ
			})

		case 2:
			// Ordenar el slice por el campo "InscripcionId"
			sort.Slice(estudiantes, func(i, j int) bool {
				// Convertir los valores del campo "nombre" a strings
				idI := estudiantes[i]["InscripcionId"].(float64)
				idJ := estudiantes[j]["InscripcionId"].(float64)
				// Comparar los InscripcionId
				return idI < idJ
			})

		case 3:
			// Ordenar el slice por el campo "PuntajeFinal"
			sort.Slice(estudiantes, func(i, j int) bool {
				// Convertir los valores del campo "nombre" a strings
				puntajeI := estudiantes[i]["PuntajeFinal"].(float64)
				puntajeJ := estudiantes[j]["PuntajeFinal"].(float64)
				// Comparar los InscripcionId
				return puntajeI < puntajeJ
			})
		}

		for i, estudiante := range estudiantes {

			indice := fmt.Sprintf("%03d", i+1)
			estudiante["codigo"] = estudiante["codigoIndice"].(string) + indice
		}

		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, estudiantes)

	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
		return APIResponseDTO
	}
	return APIResponseDTO
}

func GuardarCodificacion(data []byte) (APIResponseDTO requestresponse.APIResponse) {

	//Definición código
	var codigoGuardado []map[string]interface{}
	var estudiantes []map[string]interface{}
	date := time_bogota.TiempoBogotaFormato()
	if err := json.Unmarshal(data, &estudiantes); err == nil {

		for _, estudiante := range estudiantes {

			dataIdentifi := map[string]interface{}{
				"TipoDocumentoId":   map[string]interface{}{"Id": 14},
				"TerceroId":         map[string]interface{}{"Id": estudiante["TerceroId"]},
				"Numero":            estudiante["codigo"],
				"Activo":            true,
				"FechaCreacion":     date,
				"FechaModificacion": date,
			}

			var guardado map[string]interface{}
			errGuardar := request.SendJson("http://"+beego.AppConfig.String("TercerosService")+"/datos_identificacion", "POST", &guardado, dataIdentifi)
			if errGuardar == nil {
				codigoGuardado = append(codigoGuardado, guardado)
			} else {
				codigoGuardado = append(codigoGuardado, map[string]interface{}{"Error": errGuardar.Error()})
			}

		}
		APIResponseDTO = requestresponse.APIResponseDTO(true, 200, codigoGuardado)

	} else {
		APIResponseDTO = requestresponse.APIResponseDTO(false, 400, nil, err.Error())
		return APIResponseDTO
	}

	return APIResponseDTO

}
