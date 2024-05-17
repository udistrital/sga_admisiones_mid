package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PostCriterioIcfes",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetAspirantesByPeriodoByProyecto",
            Router: "/aspirantes",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetListaAspirantesPor",
            Router: "/aspirantespor",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PutNotaFinalAspirantes",
            Router: "/calcular_nota",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "CriteriosSubcriterios",
            Router: "/criterio",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PostCuposAdmision",
            Router: "/cupos",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetDependenciaPorVinculacionTercero",
            Router: "/dependencia_vinculacion_tercero/:id_tercero",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "CambioEstadoAspiranteByPeriodoByProyecto",
            Router: "/estado",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PostEvaluacionAspirantes",
            Router: "/evaluacion",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetEvaluacionAspirantes",
            Router: "/evaluacion/:id_programa/:id_periodo/:id_requisito",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetPuntajeTotalByPeriodoByProyecto",
            Router: "/puntaje",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "SoporteConfiguracion",
            Router: "/soporte/:id_periodo/:id_nivel",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"],
        beego.ControllerComments{
            Method: "GetAdmitidos",
            Router: "/admitidos/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"],
        beego.ControllerComments{
            Method: "GuardarCodigo",
            Router: "/codigos-periodo/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:CodificacionController"],
        beego.ControllerComments{
            Method: "GenerarCodigo",
            Router: "/codigos/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
