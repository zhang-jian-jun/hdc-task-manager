package routers

import (
	"hdc-task-manager/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
	beego.Router("/issue/hook/event", &controllers.HookEventControllers{})
	beego.Router("/gauss/issue/hook/event", &controllers.GaussHookEventControllers{})
	beego.Router("/healthz/readiness", &controllers.HealthzReadController{})
	beego.Router("/healthz/liveness", &controllers.HealthzLiveController{})
}
