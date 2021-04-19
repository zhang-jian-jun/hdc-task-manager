package main

import (
	"github.com/astaxie/beego"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	_ "hdc-task-manager/routers"
	"hdc-task-manager/task"
)

func init() {
	// Initialization log
	common.LogInit()
}

func main() {
	// init db
	dbOk := models.Initdb()
	if !dbOk {
		println("error: Database initialization failed")
		return
	}
	// Initialize a scheduled task
	taskOk := task.InitTask()
	if !taskOk {
		println("error: Timing task initialization failed, the program ends")
		task.StopTask()
		return
	}
	task.StartTask()
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}

