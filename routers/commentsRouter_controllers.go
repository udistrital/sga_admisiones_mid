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
            Method: "PutNotaFinalAspirantes",
            Router: "/calcular_nota",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "CambioEstadoAspiranteByPeriodoByProyecto",
            Router: "/cambioestado",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetAspirantesByPeriodoByProyecto",
            Router: "/consulta_aspirantes",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetPuntajeTotalByPeriodoByProyecto",
            Router: "/consulta_puntaje",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "GetEvaluacionAspirantes",
            Router: "/consultar_evaluacion/:id_programa/:id_periodo/:id_requisito",
            AllowHTTPMethods: []string{"get"},
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
            Method: "GetListaAspirantesPor",
            Router: "/getlistaaspirantespor",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PostCuposAdmision",
            Router: "/postcupos",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:AdmisionController"],
        beego.ControllerComments{
            Method: "PostEvaluacionAspirantes",
            Router: "/registrar_evaluacion",
            AllowHTTPMethods: []string{"post"},
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

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "Put",
            Router: "/:id",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_admisiones_mid/controllers:ReportesController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: "/:id",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
