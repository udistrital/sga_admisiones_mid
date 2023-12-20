package models

import (
	"fmt"
	"github.com/udistrital/utils_oas/time_bogota"
	"strings"
	"time"
)

func VerificarFechaLimite(FechaLimite string) (EnTiempo bool, Error error) {
	EnTiempo = false
	Error = nil

	FechaActual := time_bogota.Tiempo_bogota().Format(time.RFC3339)
	FechaActual = strings.Replace(fmt.Sprintf("%v", FechaActual), "+", "-", -1)
	layoutActual := "2006-01-02T15:04:05-05:00"
	FechaActualFormato, errFA := time.Parse(layoutActual, FechaActual)

	FechaLimite = strings.Replace(fmt.Sprintf("%v", FechaLimite), "+", "-", -1)
	layoutLimite := "2006-01-02T15:04:05.000-05:00"
	FechaLimiteFormato, errFL := time.Parse(layoutLimite, FechaLimite)

	if errFA != nil {
		Error = errFA
	} else if errFL != nil {
		Error = errFL
	} else {
		FechaLimiteMas23h59m59s := FechaLimiteFormato.AddDate(0, 0, 1)
		EnTiempo = FechaActualFormato.Before(FechaLimiteMas23h59m59s)
	}

	return EnTiempo, Error
}
