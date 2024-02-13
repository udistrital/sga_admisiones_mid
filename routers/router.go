// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	// "github.com/udistrital/sga_mid_admisiones/controllers"
	"github.com/udistrital/sga_admisiones_mid/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/admision",
			beego.NSInclude(
				&controllers.AdmisionController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
