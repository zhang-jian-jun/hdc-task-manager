package taskhandler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"os"
	"strings"
)

func ForciblyCancelClaim(owner, eulerToken, gitUserId string, eoi models.EulerOriginIssue, userId int64) {
	// Determine whether the user denies the task
	eu := models.EulerIssueUser{OrId: eoi.OrId, UserId: userId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId", "UserId")
	if eu.Id == 0 {
		logs.Error("Failed to give up the task, have not received this task,euErr: ", euErr)
	} else {
		if eu.Status == 2 && eoi.IssueState == "closed" {
			logs.Error("The task has been reviewed and completed, this task cannot be cancelled,eu: ", eu)
		} else if eu.Status == 1 && eoi.IssueState != "closed" {
			// give up task
			delErr := models.DeleteEulerIssueUser(&eu, "UserId", "OrId")
			if delErr == nil {
				// Edit label
				hdcTask := beego.AppConfig.String("hdc_task")
				eoi.IssueLabel = hdcTask
				EditLabel(eoi.RepoPath, eoi.IssueNumber, hdcTask, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueForciGiveUpSuccess, gitUserId)
				AddCommentToIssue(is, eoi.IssueNumber, owner, eoi.RepoPath, eulerToken)
				eir := models.QueryEulerIssueUserRecordset(userId, eoi.OrId, 2)
				if len(eir) < 1 {
					iss := fmt.Sprintf(IssueBlackSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, gitUserId)
				}
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: eoi.IssueNumber,
					RepoPath: eoi.RepoPath, Owner: owner, Status: 2}
				EulerIssueUserRecord(et)
			}
		}
	}
}

func GetEulerUserInfo(gitId string) (userId int64) {
	eu := models.EulerUser{GitUserId: gitId}
	euErr := models.QueryEulerUser(&eu, "GitUserId")
	if euErr != nil {
		logs.Error("GetEulerUserInfo, euErr: ", euErr)
		return 0
	} else {
		return eu.UserId
	}
}

func GetIssueInfo(orId int64) (models.EulerOriginIssue, error) {
	eoi := models.EulerOriginIssue{OrId: orId}
	eiErr := models.QueryEulerOriginIssue(&eoi, "OrId")
	if eiErr != nil {
		logs.Error(eiErr)
		return eoi, eiErr
	}
	return eoi, eiErr
}

func ProcEulerUserIssue(userId int64, status int8, eulerToken, ownerStr, gitUserId string) {
	eiu := models.QueryEulerIssueUserset(userId, status)
	if len(eiu) > 0 {
		for _, eu := range eiu {
			eoi, eiErr := GetIssueInfo(eu.OrId)
			if eiErr != nil {
				continue
			}
			ForciblyCancelClaim(eoi.Owner, eulerToken, gitUserId, eoi, userId)
		}
	}
}

func HandBlackListTask() {
	eulerToken := os.Getenv("GITEE_TOKEN")
	ownerStr := beego.AppConfig.String("repo::owner")
	if len(ownerStr) < 1 {
		logs.Error("No community issues can be obtained")
		return
	}
	// Query blacklist task
	ebu := models.QueryEulerBlackUserAll(1)
	if len(ebu) == 0 {
		return
	}
	// Release the problem claimed by the developer
	for _, gitId := range ebu {
		// Find user information
		userId := GetEulerUserInfo(gitId.GitUserId)
		if userId > 0 {
			// Find tasks that the user has already claimed
			ProcEulerUserIssue(userId, 2, eulerToken, ownerStr, gitId.GitUserId)
		}
		gitId.Status = 2
		upErr := models.UpdateEulerBlackUser(&gitId, "Status")
		logs.Info("upErr: ", upErr)
	}
}

func RemoveUnassignBlacklist() {
	// Query the data to be deleted
	euu := models.QueryEulerUnassignUserAll(common.GetCurTime())
	if len(euu) > 0 {
		for _, eu := range euu {
			models.DelEulerUnassignBlack(eu.Id)
		}
	}
}

func RemoveUncompleteList() {
	eulerToken := os.Getenv("GITEE_TOKEN")
	ownerStr := beego.AppConfig.String("repo::owner")
	if len(ownerStr) < 1 {
		logs.Error("No community issues can be obtained")
		return
	}
	eiu := models.QueryEulerUncompleteUserAll(common.GetCurTime())
	if len(eiu) > 0 {
		for _, ei := range eiu {
			eu := models.EulerUser{UserId: ei.UserId}
			euErr := models.QueryEulerUser(&eu, "UserId")
			if euErr != nil {
				logs.Error("euler User information query failed, euErr: ", euErr)
				continue
			}
			eoi, eiErr := GetIssueInfo(ei.OrId)
			if eiErr != nil {
				logs.Error("issue information query failed, eiErr: ", eiErr)
				continue
			}
			ForciblyCancelClaim(eoi.Owner, eulerToken, eu.GitUserId, eoi, eu.UserId)
		}
	}
}

func RemoveUncompleteHistoryList() {
	eiu := models.QueryEulerUncompleteUserHistory()
	if len(eiu) > 0 {
		for _, ei := range eiu {
			eoi, eiErr := GetIssueInfo(ei.OrId)
			if eiErr != nil {
				logs.Error("issue information query failed, eiErr: ", eiErr)
				continue
			}
			labBool := AddUserAssignTime(eoi.IssueLabel)
			if !labBool {
				ei.AssignTime = ei.CreateTime
				ei.UpdateTime = common.GetCurTime()
				eiErr := models.UpdateEulerIssueUser(&ei, "AssignTime", "UpdateTime")
				if eiErr != nil {
					logs.Error("eiErr: ", eiErr)
				}
			}
		}
	}
}

func AddUserAssignTime(label string) bool {
	if len(label) > 1 {
		labelList := strings.Split(label, ",")
		if len(labelList) > 0 {
			for _, la := range labelList {
				if la == "hdc-task-rewiew" {
					return true
				}
			}
		}
	}
	return false
}
