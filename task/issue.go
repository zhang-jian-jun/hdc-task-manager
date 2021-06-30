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

// openEuler regularly releases tasks in the blacklist
func EulerRelBlackTask(relblacklist string) {
	logs.Info("releases tasks in the blacklist start...")
	blackTask := toolbox.NewTask("BlackListTask", relblacklist, BlackListTask)
	toolbox.AddTask("BlackListTask", blackTask)
	logs.Info("End of releases tasks in the blacklist...")
}

func BlackListTask() error {
	taskhandler.HandBlackListTask()
	return nil
}

// The number of deleted and cancelled tasks has reached the blacklist of online users
func EulerRelUnassignTask(relunassign string) {
	logs.Info("The number of deleted and canceled tasks has reached the start of the blacklist of online users...")
	unassignTask := toolbox.NewTask("RelUnassignTask", relunassign, RelUnassignTask)
	toolbox.AddTask("RelUnassignTask", unassignTask)
	logs.Info("The number of deleted and canceled tasks has reached the end of the online user blacklist...")
}

func RelUnassignTask() error {
	taskhandler.RemoveUnassignBlacklist()
	return nil
}

// Release unsubmitted questions
func EulerRelUncompleteTask(reluncomplete string) {
	logs.Info("Release uncommitted tasks to start...")
	uncompleteTask := toolbox.NewTask("RelUncompleteTask", reluncomplete, RelUncompleteTask)
	toolbox.AddTask("RelUncompleteTask", uncompleteTask)
	logs.Info("Release uncommitted tasks to end...")
}

func RelUncompleteTask() error {
	taskhandler.RemoveUncompleteList()
	return nil
}