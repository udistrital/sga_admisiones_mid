package models

type ReporteEstructura struct {
	Proyecto int64    `json:"Proyecto"`
	Periodo  int64    `json:"Periodo"`
	TipoReporte  int64    `json:"Reporte"`
	Columnas []string `json:"Columnas"`
}