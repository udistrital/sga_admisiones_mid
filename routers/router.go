// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/udistrital/sga_admisiones_mid/controllers"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/errorhandler"
)

func init() {

	beego.ErrorController(&errorhandler.ErrorHandlerController{})

	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/admision",
			beego.NSInclude(
				&controllers.AdmisionController{},
			),
		),
		beego.NSNamespace("/codificacion",
			beego.NSInclude(
				&controllers.CodificacionController{},
			),
		),
		beego.NSNamespace("/reporte",
			beego.NSInclude(
				&controllers.ReportesController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
