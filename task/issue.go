package task

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"hdc-task-manager/common"
	"hdc-task-manager/taskhandler"
	"os"
	"strings"
)

// Get the original issue data on gitee
func GetOriginIssueTask(getIssue string) {
	logs.Info("Get the original issue data and start...")
	issueTask := toolbox.NewTask("GetOriginIssue", getIssue, GetOriginIssue)
	toolbox.AddTask("GetOriginIssue", issueTask)
	logs.Info("End of obtaining the original data of the issue...")
}

// Get community issues
func GetOriginIssue() error {
	accessToken := os.Getenv("GITEE_TOKEN")
	if accessToken == "" || len(accessToken) < 1 {
		logs.Error("GetOriginIssue, issue token Get failed, current time: ", common.GetCurTime())
		return errors.New("Failed to get token in environment variable")
	}
	ownerStr := beego.AppConfig.String("repo::owner")
	if len(ownerStr) < 1 {
		logs.Error("No community issues can be obtained")
		return errors.New("Invalid value")
	}
	taskLabels := beego.AppConfig.String("tasklabel")
	ownerSlice := strings.Split(ownerStr, ",")
	for _, owner := range ownerSlice {
		orErr := taskhandler.GetOriginIssue(owner, accessToken, taskLabels)
		if orErr != nil {
			logs.Error("Failed to get issue, owner: ", owner)
			continue
		}
	}
	return nil
}
