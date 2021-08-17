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
	// openEuler regularly releases tasks in the blacklist
	relblacklistflag, errxs := beego.AppConfig.Int("task::relblacklistflag")
	if relblacklistflag == 1 && errxs == nil {
		relblacklist := beego.AppConfig.String("task::relblacklist")
		EulerRelBlackTask(relblacklist)
	}
	// The number of deleted and cancelled tasks has reached the blacklist of online users
	relunassignflag, errxs := beego.AppConfig.Int("task::relunassignflag")
	if relunassignflag == 1 && errxs == nil {
		relunassign := beego.AppConfig.String("task::relunassign")
		EulerRelUnassignTask(relunassign)
	}
	// Release unsubmitted questions
	reluncompleteflag, errxs := beego.AppConfig.Int("task::reluncompleteflag")
	if reluncompleteflag == 1 && errxs == nil {
		reluncomplete := beego.AppConfig.String("task::reluncomplete")
		EulerRelUncompleteTask(reluncomplete)
	}
	// Export issue pr and number of comments
	exportissueprflag, errxs := beego.AppConfig.Int("task::exportissueprflag")
	if exportissueprflag == 1 && errxs == nil {
		exportissuepr := beego.AppConfig.String("task::exportissuepr")
		EulerIssueStatisticsTask(exportissuepr)
	}
	// Releasing the limited number of canceled tasks per month
	monthrelunassignflag, errxs := beego.AppConfig.Int("task::monthrelunassignflag")
	if monthrelunassignflag == 1 && errxs == nil {
		monthrelunassign := beego.AppConfig.String("task::monthrelunassign")
		MonthRelUnassignTask(monthrelunassign)
	}
	// Export points for the specified week
	specexportwpointflag, errxs := beego.AppConfig.Int("task::specexportwpointflag")
	if specexportwpointflag == 1 && errxs == nil {
		specexportwpoint := beego.AppConfig.String("task::specexportwpoint")
		taskhandler.GetSpecWeekPointsTask(specexportwpoint)
	}
	return true
}