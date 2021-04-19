package task

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"hdc-task-manager/taskhandler"
)

// start task
func StartTask() {
	toolbox.StartTask()
}

func StopTask() {
	toolbox.StopTask()
}

//InitTask Timing task initialization
func InitTask() bool {
	// Get the original yaml data
	getIssueFlag, errxs := beego.AppConfig.Int("task::getissueflag")
	if getIssueFlag == 1 && errxs == nil {
		getIssue := beego.AppConfig.String("task::getissue")
		GetOriginIssueTask(getIssue)
	}
	// export points task
	exportPointFlag, errxs := beego.AppConfig.Int("task::exportwpointflag")
	if exportPointFlag == 1 && errxs == nil {
		exportWPoint := beego.AppConfig.String("task::exportwpoint")
		taskhandler.GetWeekPointsTask(exportWPoint)
	}
	exportMPointFlag, errxs := beego.AppConfig.Int("task::exportmpointflag")
	if exportMPointFlag == 1 && errxs == nil {
		exportMPoint := beego.AppConfig.String("task::exportmpoint")
		taskhandler.GetMonthPointsTask(exportMPoint)
	}
	return true
}