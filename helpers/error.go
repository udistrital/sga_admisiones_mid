package helpers

import "github.com/udistrital/utils_oas/requestresponse"

func ErrEmiter(errData error, infoData ...string) requestresponse.APIResponse {
	if errData != nil {
		return requestresponse.APIResponseDTO(false, 400, nil, errData.Error())
	}

	if len(infoData) > 0 && (infoData[0] == "[map[]]" || infoData[0] == "map[]") {
		return requestresponse.APIResponseDTO(false, 404, nil, "No se encontraron datos")
	}

	return requestresponse.APIResponseDTO(false, 400, nil)
}
