package models

type Tag struct {
	Selected bool
	Required bool
}

var TagsInscripcionPrograma = map[string]Tag{
	"info_persona": {
		Selected: false,
		Required: false,
	},
	"formacion_academica": {
		Selected: false,
		Required: false,
	},
	"idiomas": {
		Selected: false,
		Required: false,
	},
	"experiencia_laboral": {
		Selected: false,
		Required: false,
	},
	"produccion_academica": {
		Selected: false,
		Required: false,
	},
	"documento_programa": {
		Selected: false,
		Required: false,
	},
	"descuento_matricula": {
		Selected: false,
		Required: false,
	},
	"propuesta_grado": {
		Selected: false,
		Required: false,
	},
	"perfil": {
		Selected: false,
		Required: false,
	},
}
