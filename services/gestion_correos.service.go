package services

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

var existingEmails = map[string]bool{}

func isEmailUnique(email string, listado []map[string]interface{}) bool {
	for _, elemento := range listado {
		if elemento["usuario_sugerido"] == email {
			return false
		}
	}
	return true
}

func markEmailAsUsed(email string) {
	existingEmails[email] = true
}

func generateUniqueEmail(primer_nombre, segundo_nombre, primer_apellido, segundo_apellido string, listado []map[string]interface{}) string {
	domain := "udistrital.edu.co"
	initial1 := ""
	initial2 := ""
	initial4 := ""

	if len(primer_nombre) > 0 {
		initial1 = string(primer_nombre[0])
	}
	if len(segundo_nombre) > 0 {
		initial2 = string(segundo_nombre[0])
	}
	if len(segundo_apellido) > 0 {
		initial4 = string(segundo_apellido[0])
	}

	email := fmt.Sprintf("%s%s%s%s@%s", initial1, initial2, primer_apellido, initial4, domain)
	uniqueEmail := strings.ToLower(email)

	i := 1
	for !isEmailUnique(uniqueEmail, listado) {
		if i < len(primer_nombre) {
			uniqueEmail = fmt.Sprintf("%s%s%s%s@%s", primer_nombre[:i+1], initial2, primer_apellido, initial4, domain)
		} else {
			uniqueEmail = fmt.Sprintf("%s%s%d@%s", primer_nombre, primer_apellido, i, domain)
		}
		i++
	}
	return uniqueEmail
}

func SugerenciaCorreosUD(idPeriodo int64) requestresponse.APIResponse {
	var listado []map[string]interface{}
	var inscripcion []map[string]interface{}
	errInscripcion := request.GetJson("http://"+beego.AppConfig.String("InscripcionService")+fmt.Sprintf("inscripcion?query=Activo:true,PeriodoId:%v,EstadoInscripcionId.Id:2&limit=0&fields=Id,PersonaId,ProgramaAcademicoId", idPeriodo), &inscripcion)
	if errInscripcion == nil && fmt.Sprintf("%v", inscripcion) != "[map[]]" {
		for _, inscrip := range inscripcion {
			fmt.Println(inscrip["PersonaId"])
			var tercero map[string]interface{}
			errTercero := request.GetJson("http://"+beego.AppConfig.String("TercerosService")+fmt.Sprintf("tercero/%v", inscrip["PersonaId"]), &tercero)
			if errTercero == nil {
				if verificarCorreoUd(tercero["UsuarioWSO2"].(string)) {
					listado = append(listado, map[string]interface{}{
						"PrimerNombre":    tercero["PrimerNombre"],
						"SegundoNombre":   tercero["SegundoNombre"],
						"PrimerApellido":  tercero["PrimerApellido"],
						"SegundoApellido": tercero["SegundoApellido"],
						"usuarioSugerio":  tercero["UsuarioWSO2"],
						"correo_asignado": tercero["UsuarioWSO2"],
						"correo_existe":   "Si",
					})
				} else {
					primerNombre := tercero["PrimerNombre"].(string)
					segundoNombre := ""
					if tercero["SegundoNombre"] != nil {
						segundoNombre = tercero["SegundoNombre"].(string)
					}
					primerApellido := tercero["PrimerApellido"].(string)
					segundoApellido := ""
					if tercero["SegundoApellido"] != nil {
						segundoApellido = tercero["SegundoApellido"].(string)
					}
					correoSugerido := generateUniqueEmail(primerNombre, segundoNombre, primerApellido, segundoApellido, listado)

					listado = append(listado, map[string]interface{}{
						"PrimerNombre":    primerNombre,
						"SegundoNombre":   segundoNombre,
						"PrimerApellido":  primerApellido,
						"SegundoApellido": segundoApellido,
						"usuarioSugerio":  correoSugerido,
						"correo_asignado": correoSugerido,
						"correo_existe":   "Generado",
					})
				}
			}
		}
		return requestresponse.APIResponseDTO(true, 200, listado)
	} else {
		return requestresponse.APIResponseDTO(false, 404, "No se encontraron datos relacionados con el periodo")
	}
}

func verificarCorreoUd(usuariowso2 string) bool {
	return strings.Contains(usuariowso2, "@udistrital.edu.co")
}
