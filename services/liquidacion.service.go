package services

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func ListarLiquidacionEstudiantes(idPeriodo string, idProyecto string, semestre string) (APIResponseDTO requestresponse.APIResponse) {
	//Mapa para guardar los estudiantes
	var estudiantes []map[string]interface{}

	//Obtener Datos del periodo
	var periodo map[string]interface{}
	errPeriodo := request.GetJson("http://"+beego.AppConfig.String("ParametrosService")+fmt.Sprintf("periodo/%v", idPeriodo), &periodo)
	if errPeriodo != nil || fmt.Sprintf("%v", periodo) == "[map[]]" {
		return ErrEmiter(errPeriodo, fmt.Sprintf("%v", periodo))
	}

	//Obtener Datos del proyecto

	var proyecto map[string]interface{}
	errProyecto := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+fmt.Sprintf("proyecto_academico_institucion/%v", idProyecto), &proyecto)
	if errProyecto != nil || fmt.Sprintf("%v", proyecto) == "map[]" {
		return ErrEmiter(errProyecto, fmt.Sprintf("%v", proyecto))
	}

	//Inscripciones de admitidos esto solo es util en el semestre 1 para otros semestres se debe considerar si continuan activos
	var inscripciones []map[string]interface{}
	errInscripciones := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=EstadoInscripcionId__Nombre:ADMITIDO,Activo:true,ProgramaAcademicoId:%v,PeriodoId:%v", idProyecto, idPeriodo), &inscripciones)
	if errInscripciones != nil || fmt.Sprintf("%v", inscripciones) == "map[]" {
		return ErrEmiter(errInscripciones, fmt.Sprintf("%v", inscripciones))
	}

	//Base para la comparación de codigo
	if (periodo["Data"].(map[string]interface{})["Ciclo"]) == "3" {
		periodo["Data"].(map[string]interface{})["Ciclo"] = "2"
	}
	codigoBase := fmt.Sprintf("%v%v%v", periodo["Data"].(map[string]interface{})["Year"], periodo["Data"].(map[string]interface{})["Ciclo"], proyecto["Codigo"])

	for _, inscripcion := range inscripciones {

		//Obtener Datos basicos Tercero
		var tercero []map[string]interface{}
		errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero?query=Id:%v", inscripcion["PersonaId"]), &tercero)
		if errTercero != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return ErrEmiter(errTercero, fmt.Sprintf("%v", tercero))
		}

		//Obtener Documento Tercero
		var terceroDocumento []map[string]interface{}
		errTerceroDocumento := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CC,Activo:true,TerceroId:%v", inscripcion["PersonaId"]), &terceroDocumento)
		if errTerceroDocumento != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return ErrEmiter(errTerceroDocumento, fmt.Sprintf("%v", terceroDocumento))
		}

		//Obtener Codigo Tercero
		var terceroCodigo []map[string]interface{}
		errTerceroCodigo := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("datos_identificacion?query=TipoDocumentoId__CodigoAbreviacion:CODE,Activo:true,TerceroId:%v,Numero__contains:%v", inscripcion["PersonaId"], codigoBase), &terceroCodigo)
		if errTerceroCodigo != nil || fmt.Sprintf("%v", tercero) == "[map[]]" {
			return ErrEmiter(errTerceroCodigo, fmt.Sprintf("%v", terceroCodigo))
		}

		//Obtener conceptos

		//base
		var vigencia = 1 //actual? según el semestre?

		var derechoPecuniario []map[string]interface{}
		errDerechoPecuniario := request.GetJson("http://"+beego.AppConfig.String("DerechosService")+fmt.Sprintf("/derechos-pecuniarios/vigencias/?query=Id:%v", vigencia), &derechoPecuniario)
		if errTercero != nil || fmt.Sprintf("%v", derechoPecuniario) == "[map[]]" {
			return ErrEmiter(errDerechoPecuniario, fmt.Sprintf("%v", derechoPecuniario))
		}
		//descuentos

		estudiantes = append(estudiantes, map[string]interface{}{
			"Nombre":          fmt.Sprintf("%v %v", tercero[0]["PrimerNombre"], tercero[0]["SegundoNombre"]),
			"PrimerApellido":  tercero[0]["PrimerApellido"],
			"SegundoApellido": tercero[0]["SegundoApellido"],
			"Documento":       terceroDocumento[0]["Numero"],
			"Codigo":          terceroCodigo[0]["Numero"],
		})
		return requestresponse.APIResponseDTO(true, 200, estudiantes)
	}

	return requestresponse.APIResponseDTO(false, 400, nil, "Error")
}
