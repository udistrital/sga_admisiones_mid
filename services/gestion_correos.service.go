package services

import (
	"fmt"
	"strings"

	"github.com/udistrital/utils_oas/requestresponse"
)

// isEmailUnique es una función ficticia que se reemplazará por una verificación real
var existingEmails = map[string]bool{}

func isEmailUnique(email string) bool {
	_, exists := existingEmails[email]
	return !exists
}

// markEmailAsUsed marca un correo electrónico como utilizado
func markEmailAsUsed(email string) {
	existingEmails[email] = true
}

// generateUniqueEmail genera un correo electrónico único basado en la metodología especificada
func generateUniqueEmail(primer_nombre, segundo_nombre, tercer_nombre, primer_apellido, segundo_apellido string) string {
	domain := "udistrital.edu.co"
	initial1 := ""
	initial2 := ""
	initial3 := ""
	initial4 := ""

	if len(primer_nombre) > 0 {
		initial1 = string(primer_nombre[0])
	}
	if len(segundo_nombre) > 0 {
		initial2 = string(segundo_nombre[0])
	}
	if len(tercer_nombre) > 0 {
		initial3 = string(tercer_nombre[0])
	}
	if len(segundo_apellido) > 0 {
		initial4 = string(segundo_apellido[0])
	}

	email := fmt.Sprintf("%s%s%s%s%s@%s", initial1, initial2, initial3, primer_apellido, initial4, domain)
	uniqueEmail := email

	i := 1
	for !isEmailUnique(uniqueEmail) {
		if i < len(primer_nombre) {
			uniqueEmail = fmt.Sprintf("%s%s%s%s%s@%s", primer_nombre[:i+1], initial2, initial3, primer_apellido, initial4, domain)
		} else {
			uniqueEmail = fmt.Sprintf("%s%s%d@%s", primer_nombre, primer_apellido, i, domain)
		}
		i++
	}

	markEmailAsUsed(uniqueEmail)
	return strings.ToLower(uniqueEmail)
}

// ConsultaInscritosAdmitidos consulta los inscritos admitidos para un periodo y proyecto específico
func ConsultaInscritosAdmitidos(idPeriodo int64, idProyecto int64) (APIResponseDTO requestresponse.APIResponse) {
	var listado []map[string]interface{}

	// Lógica para obtener los inscritos admitidos y procesar los datos necesarios
	ManejoCasosGetLista(idPeriodo, idProyecto, 10, &listado)

	// Preparar los datos específicos requeridos: número de documento, nombre completo, teléfono y correo electrónico
	resultados := make([]map[string]interface{}, 0)
	for _, estudiante := range listado {
		datosEstudiante := map[string]interface{}{
			"NumeroDocumento":   estudiante["NumeroDocumento"],
			"PrimerNombre":      estudiante["PrimerNombre"],
			"SegundoNombre":     estudiante["SegundoNombre"],
			"PrimerApellido":    estudiante["PrimerApellido"],
			"SegundoApellido":   estudiante["SegundoApellido"],
			"Telefono":          estudiante["Telefono"],
			"CorreoElectronico": sugerirCorreoInstitucional(estudiante),
		}
		resultados = append(resultados, datosEstudiante)
	}

	// Devolver la respuesta en formato APIResponseDTO
	APIResponseDTO = requestresponse.APIResponseDTO(true, 200, resultados)
	return APIResponseDTO
}

// sugerirCorreoInstitucional genera y sugiere un correo institucional único para el estudiante
func sugerirCorreoInstitucional(estudiante map[string]interface{}) string {
	correoExistente := fmt.Sprintf("%v", estudiante["CorreoElectronico"])

	if correoExistente != "" {
		return correoExistente
	}

	// Generar correo basado en nombre y apellidos
	primerNombre := fmt.Sprintf("%v", estudiante["PrimerNombre"])
	segundoNombre := fmt.Sprintf("%v", estudiante["SegundoNombre"])
	tercerNombre := fmt.Sprintf("%v", estudiante["TercerNombre"])
	primerApellido := fmt.Sprintf("%v", estudiante["PrimerApellido"])
	segundoApellido := fmt.Sprintf("%v", estudiante["SegundoApellido"])

	// Lógica para generar correo único
	correoPropuesto := generateUniqueEmail(primerNombre, segundoNombre, tercerNombre, primerApellido, segundoApellido)

	// Verificar homónimos y ajustar si es necesario
	correoAjustado := verificarHomonimos(correoPropuesto, estudiante["NumeroDocumento"])

	return correoAjustado
}

// verificarHomonimos verifica si hay homónimos para el correo institucional y ajusta si es necesario
func verificarHomonimos(correoPropuesto string, numeroDocumento interface{}) string {
	if !isEmailUnique(correoPropuesto) {
		correoAjustado := fmt.Sprintf("%s.%v@udistrital.edu.co", correoPropuesto, numeroDocumento)
		if !isEmailUnique(correoAjustado) {
			i := 1
			for {
				correoIncremental := fmt.Sprintf("%s.%v.%d@udistrital.edu.co", correoPropuesto, numeroDocumento, i)
				if isEmailUnique(correoIncremental) {
					return correoIncremental
				}
				i++
			}
		}
		return correoAjustado
	}

	return correoPropuesto
}
