package taskhandler

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"hdc-task-manager/util"
	"os"
	"strconv"
	"strings"
	"sync"
)

var pointLock sync.Mutex
var gaussLock sync.Mutex

type EulerIssueUserRecordTp struct {
	UserId      int64
	OrId        int64
	IssueNumber string
	RepoPath    string
	Owner       string
	Status      int8
}

func CreateIssueBody(eulerToken, path, statusName string, eoi models.EulerOriginIssue) string {
	requestBody := ""
	body := eoi.IssueBody
	requestBody = fmt.Sprintf(`{
			"access_token": "%s",
			"repo": "%s", 
			"title": "%s",
			"state": "%s",
			"body": "%s",
			"assignee": "%s",
			"labels": "%s",
			"security_hole": "false"
			}`, eulerToken, path, eoi.Title, statusName, body, eoi.IssueAssignee, eoi.IssueLabel)
	return requestBody
}

func UpdateIssueToGit(eulerToken, owner, path, issueState string, eoi models.EulerOriginIssue) error {
	if eulerToken != "" && owner != "" && path != "" {
		url := "https://gitee.com/api/v5/repos/" + owner + "/issues/" + eoi.IssueNumber
		statusName := IssueStateRev(issueState)
		requestBody := CreateIssueBody(eulerToken, path, statusName, eoi)
		logs.Info("UpdateIssueToGit, isssue_body: ", requestBody)
		if requestBody != "" && len(requestBody) > 1 {
			resp, err := util.HTTPPatch(url, requestBody)
			if err != nil {
				logs.Error("UpdateIssueToGit, Update issue failed, issueNum: ", eoi.IssueNumber, "err: ", err)
				return errors.New("Failed to call gitee to update the issue interface")
			}
			if _, ok := resp["id"]; !ok {
				logs.Error("UpdateIssueToGit, Failed to create issue, err: ", ok, "url: ", url)
				return errors.New("Failed to call gitee to update the issue interface")
			}
			// Update the status of issue in db
			if eoi.IssueState != statusName {
				eoi.IssueState = statusName
				upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
				if upIssueErr != nil {
					logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
				}
			}
		}
	}
	return nil
}

// Calculate the points earned by users
func CalculateUserPoints(eulerToken string, eoi models.EulerOriginIssue) {
	// Query user information
	eiu := models.EulerIssueUser{OrId: eoi.OrId, Status: 1}
	eiuErr := models.QueryEulerIssueUser(&eiu, "OrId", "Status")
	if eiuErr != nil {
		logs.Error("Points cannot be calculated or points have already been calculated, eiuErr: ", eiuErr)
		return
	}
	eiu.Status = 2
	upeiuErr := models.UpdateEulerIssueUser(&eiu, "Status")
	if upeiuErr != nil {
		logs.Error("UpdateEulerIssueUser, upeiuErr: ", upeiuErr)
		return
	}
	pointValue := int64(eoi.DifficultValue * eoi.EmergencyValue)
	eid := models.EulerUserIntegDetail{UserId: eiu.UserId, OrId: eoi.OrId}
	eidErr := models.QueryEulerUserIntegDetail(&eid, "UserId", "OrId")
	if eidErr == nil {
		logs.Info("The user has already calculated the points, eid: ", eid)
		return
	} else {
		eid = models.EulerUserIntegDetail{UserId: eiu.UserId, OrId: eoi.OrId,
			IntegralValue: pointValue, CreateTime: common.GetCurTime()}
		id, indErr := models.InsertEulerUserIntegDetail(&eid)
		if id > 0 {
			eic := models.EulerUserIntegCount{UserId: eiu.UserId}
			eicErr := models.QueryEulerUserIntegCount(&eic, "UserId")
			if eicErr != nil {
				eic = models.EulerUserIntegCount{UserId: eiu.UserId, IntegralValue: pointValue, CreateTime: common.GetCurTime()}
				eicId, ineicErr := models.InsertEulerUserIntegCount(&eic)
				if ineicErr != nil {
					logs.Error("InsertEulerUserIntegCount, ineicErr: ", ineicErr, eicId)
				}
			} else {
				eic.IntegralValue += pointValue
				upicErr := models.UpdateEulerUserIntegCount(&eic, "IntegralValue")
				if upicErr != nil {
					logs.Error("UpdateEulerUserIntegCount, upicErr: ", upicErr)
				}
			}
			// After earning points, send a private message
			eu := models.EulerUser{UserId: eiu.UserId}
			euErr := models.QueryEulerUser(&eu, "UserId")
			if euErr != nil {
				logs.Error("QueryEulerUser, euErr: ", euErr)
			} else {
				iss := fmt.Sprintf(IssuePointSend, eoi.GitUrl, pointValue)
				SendPrivateLetters(eulerToken, iss, eu.GitUserId)
			}
		} else {
			logs.Error("InsertEulerUserIntegDetail, indErr:", indErr)
		}
	}
	eiuc := models.EulerIssueUserComplate{UserId: eiu.UserId, OrId: eoi.OrId}
	eiucErr := models.QueryEulerIssueUserComplate(&eiuc, "UserId", "OrId")
	if eiucErr != nil {
		eiuc = models.EulerIssueUserComplate{UserId: eiu.UserId, OrId: eoi.OrId,
			IssueNumber: eoi.IssueNumber, RepoPath: eoi.RepoPath, Owner: eoi.Owner,
			Status: 1, IntegralValue: pointValue, CreateTime: common.GetCurTime()}
		ucId, ucErr := models.InsertEulerIssueUserComplate(&eiuc)
		if ucErr != nil {
			logs.Error("InsertEulerIssueUserComplate, ucErr: ", ucErr, ucId)
		}
	}
}

// Entry function for handling issue status
func HandleIssueStateChange(issueHook *models.IssuePayload) error {
	eulerToken := os.Getenv("GITEE_TOKEN")
	issueId := issueHook.Issue.Id
	issueTitle := common.TrimString(issueHook.Issue.Title)
	issueType := common.TrimString(issueHook.Issue.TypeName)
	issueNumber := common.TrimString(issueHook.Issue.Number)
	repoPath := common.TrimString(issueHook.Repository.Path)
	owner := common.TrimString(issueHook.Repository.NameSpace)
	if issueType == CIssueType || issueTitle == CIssueType {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueId, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("QueryEulerOriginIssue, Data does not exist, eiErr: ", eiErr)
			return errors.New("no data")
		}
		switch issueHook.State {
		case IssueOpenState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.IssueAssignee != issueHook.Issue.User.Login {
				is := fmt.Sprintf(IssueStateProc, issueHook.Issue.User.Login)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueOpenState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
		case IssueProgressState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.IssueAssignee != issueHook.Issue.User.Login {
				is := fmt.Sprintf(IssueStateProc, issueHook.Issue.User.Login)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueProgressState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
		case IssueCloseState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.IssueAssignee != issueHook.Issue.User.Login {
				is := fmt.Sprintf(IssueStateProc, issueHook.Issue.User.Login)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, "open", eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			// Calculate the points earned by users
			pointLock.Lock()
			CalculateUserPoints(eulerToken, eoi)
			pointLock.Unlock()
			// Modify data status
			eoi.IssueState = IssueCloseState
			upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
			}
		case IssueRejectState:
			// Non-reviewers, cannot modify the status of the issue
			if eoi.IssueAssignee != issueHook.Issue.User.Login {
				is := fmt.Sprintf(IssueStateProc, issueHook.Issue.User.Login)
				AddCommentToIssue(is, issueHook.Issue.Number, owner, issueHook.Repository.Path, eulerToken)
				upErr := UpdateIssueToGit(eulerToken, owner, repoPath, eoi.IssueState, eoi)
				if upErr != nil {
					logs.Error("UpdateIssueToGit, upErr: ", upErr)
				}
				return errors.New("No operation authority")
			}
			upErr := UpdateIssueToGit(eulerToken, owner, repoPath, IssueRejectState, eoi)
			if upErr != nil {
				logs.Error("UpdateIssueToGit, upErr: ", upErr)
			}
		}
	}
	return nil
}

// Parse issue comments
func HandleIssueComment(payload models.CommentPayload) {
	if payload.Issue == nil || payload.Comment == nil {
		return
	}
	if payload.Comment.User == nil {
		return
	}
	// The default timeout for receiving hooks
	logs.Info("payload.Comment: ", payload.Comment, ", Number: ", payload.Issue.Number, "id: ", payload.Issue.Id)
	issueNum := payload.Issue.Number           //issue number string
	issueId := payload.Issue.Id                // issue id int64
	cBody := payload.Comment.Body              //Comment subject
	cuAccount := payload.Comment.User.UserName //gitee domain address
	hdcAssignedCmd := beego.AppConfig.DefaultString("hdcAssignedCmd", "/hdc-assigned")
	hdcCompletedCmd := beego.AppConfig.DefaultString("hdcCompletedCmd", "/hdc-completed")
	hdcUnassignCmd := beego.AppConfig.DefaultString("hdcUnassignCmd", "/hdc-unassign")
	closeIssueCmd := beego.AppConfig.DefaultString("close_issue", "/close")
	if issueNum == "" || cuAccount == "" || cBody == "" {
		logs.Error("Data has null values: issueNum, cuAccount, cBody: ", issueNum, cuAccount, cBody)
		return
	}
	if payload.Issue.State == "closed" || payload.Issue.State == "rejected" ||
		payload.Issue.State == "已完成" || payload.Issue.State == "已拒绝" {
		logs.Error("Cannot edit comment, value: ", payload.Issue)
		return
	}
	issueTitle := common.TrimString(payload.Issue.Title)
	issueType := common.TrimString(payload.Issue.TypeName)
	issueNumber := common.TrimString(payload.Issue.Number)
	repoPath := common.TrimString(payload.Repository.Path)
	owner := common.TrimString(payload.Repository.NameSpace)
	if issueType == CIssueType || issueTitle == CIssueType {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueId, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("QueryEulerOriginIssue, Data does not exist, eiErr: ", eiErr)
			return
		}
		eulerToken := os.Getenv("GITEE_TOKEN")
		if strings.HasPrefix(cBody, hdcAssignedCmd) {
			// first-claimed task
			UserClaimTask(payload, eulerToken, owner, eoi)
		} else if strings.HasPrefix(cBody, hdcCompletedCmd) {
			// User submits task
			UserSubmitsTask(payload, eulerToken, owner, eoi)
		} else if strings.HasPrefix(cBody, hdcUnassignCmd) {
			// Give up the task
			UserGiveUpTask(payload, eulerToken, owner, eoi)
		} else if strings.HasPrefix(cBody, closeIssueCmd) {
			// close cmd
			AssignCloseIssue(payload, eulerToken, owner, eoi)
		}
	}
}

func AssignCloseIssue(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	if eoi.IssueState == "closed" {
		// The issue has been closed and cannot be operated again
		logs.Error("The issue has been closed and cannot be operated again,issuetmp: ", eoi)
		return
	}
	// Non-reviewers, cannot modify the status of the issue
	if eoi.IssueAssignee == payload.Comment.User.Login {
		upErr := UpdateIssueToGit(eulerToken, owner, eoi.RepoPath, IssueCloseState, eoi)
		if upErr != nil {
			logs.Error("UpdateIssueToGit, upErr: ", upErr)
			return
		}
		// Modify data status
		eoi.IssueState = IssueCloseState
		upIssueErr := models.UpdateEulerOriginIssue(&eoi, "IssueState")
		if upIssueErr != nil {
			logs.Error("UpdateEulerOriginIssue, upIssueErr: ", upIssueErr)
		}
		return
	}
}

// Give up the task
func UserGiveUpTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	userId := StoreGitUser(payload)
	if userId > 0 {
		// Determine whether the user denies the task
		eu := models.EulerIssueUser{OrId: eoi.OrId, UserId: userId}
		euErr := models.QueryEulerIssueUser(&eu, "OrId", "UserId")
		if eu.Id == 0 {
			logs.Error("Failed to give up the task, have not received this task,euErr: ", euErr)
			is := fmt.Sprintf(IssueGiveUpTask, payload.Comment.User.Login)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 10}
			EulerIssueUserRecord(et)
		} else {
			if eu.Status == 2 && eoi.IssueState == "closed" {
				is := fmt.Sprintf(IssueGiveUpFailure, payload.Comment.User.Login)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 11}
				EulerIssueUserRecord(et)
			} else if eu.Status == 1 && eoi.IssueState != "closed" {
				// give up task
				delErr := models.DeleteEulerIssueUser(&eu, "UserId", "OrId")
				if delErr == nil {
					// Edit label
					hdcTask := beego.AppConfig.String("hdc_task")
					eoi.IssueLabel = hdcTask
					EditLabel(payload, hdcTask, eulerToken, owner, eoi)
					is := fmt.Sprintf(IssueGiveUpSuccess, payload.Comment.User.Login)
					AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
					eir := models.QueryEulerIssueUserRecordset(userId, eoi.OrId, 2)
					if len(eir) < 1 {
						iss := fmt.Sprintf(IssueGiveUpSuccessSend, eoi.GitUrl)
						SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
					}
					et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
						RepoPath: payload.Repository.Path, Owner: owner, Status: 2}
					EulerIssueUserRecord(et)
				}
			}
		}
	}
}

// UserSubmitsTask User submits task
func UserSubmitsTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	userId := StoreGitUser(payload)
	if userId > 0 {
		// Determine whether the submitted task and the claimed task are the same user
		eu := models.EulerIssueUser{OrId: eoi.OrId, UserId: userId}
		euErr := models.QueryEulerIssueUser(&eu, "OrId", "UserId")
		if eu.Id == 0 {
			logs.Error("No user claim information is queried,euErr: ", euErr)
			is := fmt.Sprintf(IssueClaimSubmit, payload.Comment.User.Login)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 8}
			EulerIssueUserRecord(et)
		} else {
			if eu.Status == 2 {
				// Edit label
				hdcTaskRewiew := beego.AppConfig.String("hdc_task_rewiew")
				EditLabel(payload, hdcTaskRewiew, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueClaimSubmitComplete, payload.Comment.User.Login)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 9}
				EulerIssueUserRecord(et)
			} else if eu.Status == 1 {
				// Edit label
				hdcTaskRewiew := beego.AppConfig.String("hdc_task_rewiew")
				EditLabel(payload, hdcTaskRewiew, eulerToken, owner, eoi)
				is := fmt.Sprintf(IssueClaimReSubmit, payload.Comment.User.Login)
				AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
				eir := models.QueryEulerIssueUserRecordset(userId, eoi.OrId, 3)
				if len(eir) < 1 {
					iss := fmt.Sprintf(IssueClaimReSubmitSend, eoi.GitUrl)
					SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
				}
				icrs := fmt.Sprintf(IssueClaimReReviewerSend, payload.Comment.User.Login, eoi.GitUrl)
				SendPrivateLetters(eulerToken, icrs, eoi.IssueAssignee)
				et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
					RepoPath: payload.Repository.Path, Owner: owner, Status: 3}
				EulerIssueUserRecord(et)
			}
		}
	}
}

func UserClaimTask(payload models.CommentPayload, eulerToken, owner string, eoi models.EulerOriginIssue) {
	// Store user information
	userId := StoreGitUser(payload)
	if userId > 0 {
		VerifyClaimReq(payload, userId, eulerToken, owner, eoi)
	} else {
		logs.Error("The user cannot claim the current task, the information is wrong, payload: ", payload)
		is := fmt.Sprintf(IssueClaimWrong, payload.Comment.User.Login)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaimWrongSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
	}
}

func StoreGitUser(payload models.CommentPayload) int64 {
	ei := models.EulerUser{GitUserId: payload.Comment.User.Login}
	eiErr := models.QueryEulerUser(&ei, "GitUserId")
	if eiErr != nil {
		// insert data
		ei.CreateTime = common.GetCurTime()
		ei.Status = 1
		ei.EmailAddr = payload.Comment.User.Email
		ei.GitUserId = payload.Comment.User.Login
		userId, inErr := models.InsertEulerUser(&ei)
		if inErr != nil {
			logs.Error("InsertEulerUser, inerr: ", inErr)
		}
		return userId
	} else {
		// update data
		ei.Status = 1
		ei.UpdateTime = common.GetCurTime()
		upErr := models.UpdateEulerUser(&ei, "Status", "UpdateTime")
		if upErr != nil {
			logs.Error("UpdateEulerUser, inerr: ", upErr)
		}
		return ei.UserId
	}
}

func VerifyClaimReq(payload models.CommentPayload, userId int64, eulerToken, owner string, eoi models.EulerOriginIssue) {
	issueCount := beego.AppConfig.DefaultInt("claimed::issue_count", 3)
	// Verify whether it is the first-claimed task
	eiu := models.QueryEulerIssueUserset(userId, 1)
	ciaimCount := len(eiu)
	if ciaimCount >= issueCount {
		cc := fmt.Sprintf(IssueClaimFailure, payload.Comment.User.Login)
		AddCommentToIssue(cc, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		SendPrivateLetters(eulerToken, IssueClaimFailureSend, payload.Comment.User.Login)
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 5}
		EulerIssueUserRecord(et)
		return
	} else {
		if ciaimCount > 0 {
			for _, e := range eiu {
				if e.OrId == eoi.OrId {
					ic := fmt.Sprintf(IssueClaimSameTask, payload.Comment.User.Login)
					AddCommentToIssue(ic, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
					et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
						RepoPath: payload.Repository.Path, Owner: owner, Status: 6}
					EulerIssueUserRecord(et)
					return
				}
			}
		}
		// Claim the task successfully
		StartClaimTask(payload, userId, eulerToken, owner, eoi)
	}

}

func StartClaimTask(payload models.CommentPayload, userId int64, eulerToken, owner string, eoi models.EulerOriginIssue) {
	eu := models.EulerIssueUser{OrId: eoi.OrId}
	euErr := models.QueryEulerIssueUser(&eu, "OrId")
	if eu.Id == 0 || euErr != nil {
		eu = models.EulerIssueUser{OrId: eoi.OrId, UserId: userId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, SendEmail: 1, Status: 1, CreateTime: common.GetCurTime()}
		id, inErr := models.InsertEulerIssueUser(&eu)
		if id > 0 && inErr == nil {
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 1}
			EulerIssueUserRecord(et)
			is := fmt.Sprintf(IssueClaimSuccess, payload.Comment.User.Login)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueClaimSuccessSend, eoi.GitUrl)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
			// Edit label
			hdcTaskAssign := beego.AppConfig.String("hdc_task_assign")
			EditLabel(payload, hdcTaskAssign, eulerToken, owner, eoi)
		} else {
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
				RepoPath: payload.Repository.Path, Owner: owner, Status: 7}
			EulerIssueUserRecord(et)
			is := fmt.Sprintf(IssueClaimPree, payload.Comment.User.Login)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
			iss := fmt.Sprintf(IssueClaimPreeSend, eoi.GitUrl)
			SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
		}
	} else {
		et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: payload.Issue.Number,
			RepoPath: payload.Repository.Path, Owner: owner, Status: 7}
		EulerIssueUserRecord(et)
		is := fmt.Sprintf(IssueClaimPree, payload.Comment.User.Login)
		AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, eulerToken)
		iss := fmt.Sprintf(IssueClaimPreeSend, eoi.GitUrl)
		SendPrivateLetters(eulerToken, iss, payload.Comment.User.Login)
	}
}

func EulerIssueUserRecord(et EulerIssueUserRecordTp) {
	eir := models.EulerIssueUserRecord{UserId: et.UserId, OrId: et.OrId, IssueNumber: et.IssueNumber,
		RepoPath: et.RepoPath, Owner: et.Owner, Status: et.Status, CreateTime: common.GetCurTime()}
	models.InsertEulerIssueUserRecord(&eir)
}

// EditLabel Edit label
func EditLabel(payload models.CommentPayload, hdcTaskAssign, eulerToken, owner string, eoi models.EulerOriginIssue) {
	labels := eoi.IssueLabel
	if len(labels) > 1 {
		if !strings.Contains(labels, hdcTaskAssign) {
			labels = labels + "," + hdcTaskAssign
		}
	} else {
		labels = hdcTaskAssign
	}
	ChangeIssueLabel(eulerToken, payload.Repository.Path, payload.Issue.Number, owner, labels)
	eoi.IssueLabel = labels
	eoi.UpdateTime = common.GetCurTime()
	upErr := models.UpdateEulerOriginIssue(&eoi, "IssueLabel", "UpdateTime")
	if upErr != nil {
		logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
	}
}

// Entry function for handling issue status
func HandleGaussIssueStateChange(issueHook *models.IssuePayload) error {
	issueId := issueHook.Issue.Id
	issueNumber := common.TrimString(issueHook.Issue.Number)
	repoPath := common.TrimString(issueHook.Repository.Path)
	owner := common.TrimString(issueHook.Repository.NameSpace)
	goi := models.GaussOriginIssue{Owner: owner, RepoPath: repoPath,
		IssueId: issueId, IssueNumber: issueNumber}
	eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
	if goi.OrId == 0 {
		logs.Error("QueryGaussOriginIssue, Data does not exist, eiErr: ", eiErr)
		return errors.New("no data")
	}
	switch issueHook.State {
	case IssueOpenState:
		// Update the status of issue in db
		if goi.IssueState != IssueOpenState {
			goi.IssueState = IssueOpenState
			upIssueErr := models.UpdateGaussOriginIssue(&goi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateGaussOriginIssue, upIssueErr: ", upIssueErr, goi.IssueState)
			}
		}
	case IssueProgressState:
		// Update the status of issue in db
		if goi.IssueState != IssueProgressState {
			goi.IssueState = IssueProgressState
			upIssueErr := models.UpdateGaussOriginIssue(&goi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateGaussOriginIssue, upIssueErr: ", upIssueErr, goi.IssueState)
			}
		}
	case IssueCloseState:
		// Update the status of issue in db
		if goi.IssueState != IssueCloseState {
			goi.IssueState = IssueCloseState
			upIssueErr := models.UpdateGaussOriginIssue(&goi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateGaussOriginIssue, upIssueErr: ", upIssueErr, goi.IssueState)
			}
		}
	case IssueRejectState:
		// Update the status of issue in db
		if goi.IssueState != IssueRejectState {
			goi.IssueState = IssueRejectState
			upIssueErr := models.UpdateGaussOriginIssue(&goi, "IssueState")
			if upIssueErr != nil {
				logs.Error("UpdateGaussOriginIssue, upIssueErr: ", upIssueErr, goi.IssueState)
			}
		}
	}
	return nil
}

// Parse issue comments
func HandleGaussIssueComment(payload models.CommentPayload) {
	if payload.Issue == nil || payload.Comment == nil {
		return
	}
	if payload.Comment.User == nil {
		return
	}
	// The default timeout for receiving hooks
	logs.Info("payload.Comment: ", payload.Comment,
		", Number: ", payload.Issue.Number, "id: ", payload.Issue.Id)
	issueNum := payload.Issue.Number        //issue number string
	issueId := payload.Issue.Id             // issue id int64
	cBody := payload.Comment.Body           //Comment subject
	cuAccount := payload.Comment.User.Login //gitee domain address
	if issueNum == "" || cuAccount == "" || cBody == "" {
		logs.Error("Data has null values: issueNum, "+
			"cuAccount, cBody: ", issueNum, cuAccount, cBody)
		return
	}
	if payload.Issue.State == "closed" || payload.Issue.State == "rejected" ||
		payload.Issue.State == "已完成" || payload.Issue.State == "已拒绝" {
		logs.Error("Cannot edit comment, value: ", payload.Issue)
		return
	}
	assignFlag := reviewIsvalid(cuAccount)
	if !assignFlag {
		logs.Error("Invalid comment, discard, body: ", cBody, ", cuAccount: ", cuAccount)
		return
	}
	issueNumber := common.TrimString(payload.Issue.Number)
	repoPath := common.TrimString(payload.Repository.Path)
	owner := common.TrimString(payload.Repository.NameSpace)
	goi := models.GaussOriginIssue{Owner: owner, RepoPath: repoPath,
		IssueId: issueId, IssueNumber: issueNumber}
	eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
	if goi.OrId == 0 {
		logs.Error("QueryGaussOriginIssue, Data does not exist, eiErr: ", eiErr)
		return
	}
	gaussToken := os.Getenv("GITEE_GAUSS_TOKEN")
	issuePointStr := beego.AppConfig.String("gauss::issue_point")
	if strings.HasPrefix(cBody, GaussHighCmd) {
		// first-claimed task
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussHighCmd, goi.IssueCreate, goi.GitUrl, goi.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, goi.OrId, pointValue, 1,
				goi.IssueNumber, goi.RepoPath, goi.Owner)
			is := fmt.Sprintf(IssueGaussPointComment, goi.IssueCreate)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, gaussToken)
		}
	} else if strings.HasPrefix(cBody, GaussMediumCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussMediumCmd, goi.IssueCreate, goi.GitUrl, goi.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, goi.OrId, pointValue, 1,
				goi.IssueNumber, goi.RepoPath, goi.Owner)
			is := fmt.Sprintf(IssueGaussPointComment, goi.IssueCreate)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, gaussToken)
		}
	} else if strings.HasPrefix(cBody, GaussLowCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussLowCmd, goi.IssueCreate, goi.GitUrl, goi.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, goi.OrId, pointValue, 1,
				goi.IssueNumber, goi.RepoPath, goi.Owner)
			is := fmt.Sprintf(IssueGaussPointComment, goi.IssueCreate)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, gaussToken)
		}
	} else if strings.HasPrefix(cBody, GaussZeroCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussZeroCmd, goi.IssueCreate, goi.GitUrl, goi.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, goi.OrId, pointValue, 1,
				goi.IssueNumber, goi.RepoPath, goi.Owner)
			is := fmt.Sprintf(IssueGaussPointComment, goi.IssueCreate)
			AddCommentToIssue(is, payload.Issue.Number, owner, payload.Repository.Path, gaussToken)
		}
	}
}

// Determine whether the reviewer is valid
func reviewIsvalid(cuAccount string) bool {
	assignFlag := false
	assignee := beego.AppConfig.String("gauss::assignee")
	if len(assignee) > 1 {
		assigneeSlice := strings.Split(assignee, ",")
		if len(assigneeSlice) > 0 {
			for _, as := range assigneeSlice {
				if as == cuAccount {
					assignFlag = true
					break
				}
			}
		}
	}
	return assignFlag
}

// Increase user points
func AddGaussUserPoints(gaussToken, pointStr, cmd, issueCreate, gitUrl string, orId int64) (int64, int64) {
	pointValue := int64(0)
	if len(pointStr) > 1 {
		pointSlice := strings.Split(pointStr, ",")
		if len(pointSlice) > 0 {
			cmdSlice := strings.Split(cmd, "-")
			for _, ps := range pointSlice {
				psSlice := strings.Split(ps, ":")
				if cmdSlice[1] == psSlice[0] {
					pointValue, _ = strconv.ParseInt(psSlice[1], 10, 64)
				}
			}
		}
	}
	gaussLock.Lock()
	userId := CalculateGaussUserPoints(gaussToken, gitUrl, issueCreate, orId, pointValue)
	gaussLock.Unlock()
	return userId, pointValue

}

func AddGaussPointsComplete(userId, orId, pointValue int64, dataType int8, number, repoPath, owner string) {
	gipc := models.GaussIssuePrComplate{UserId: userId, OrId: orId, Type: dataType}
	gipcErr := models.QueryGaussIssuePrComplate(&gipc, "UserId", "OrId", "Type")
	if gipcErr != nil {
		gipc = models.GaussIssuePrComplate{UserId: userId, OrId: orId,
			Number: number, RepoPath: repoPath, Owner: owner,
			Status: 1, IntegralValue: pointValue,
			CreateTime: common.GetCurTime(), Type: dataType}
		gipcId, gipcErr := models.InsertGaussIssuePrComplate(&gipc)
		if gipcErr != nil {
			logs.Error("InsertGaussIssuePrComplate, gipcErr: ", gipcErr, gipcId)
		}
	} else {
		gipc.IntegralValue = pointValue
		gipc.UpdateTime = common.GetCurTime()
		gipcuErr := models.UpdateGaussIssuePrComplate(&gipc, "IntegralValue", "UpdateTime")
		if gipcuErr != nil {
			logs.Error("UpdateGaussIssuePrComplate, gipcuErr: ", gipcuErr)
		}
	}
}

// Calculate user points
func CalculateGaussUserPoints(gaussToken, gitUrl, issueCreate string, orId, pointValue int64) int64 {
	// Query user information
	gu := models.GaussUser{GitUserId: issueCreate}
	guErr := models.QueryGaussUser(&gu, "GitUserId")
	if guErr != nil {
		logs.Error("QueryGaussUser, euErr: ", guErr)
		gu.GitUserId = issueCreate
		gu.Status = 1
		gu.CreateTime = common.GetCurTime()
		userId, inErr := models.InsertGaussUser(&gu)
		if userId == 0 {
			logs.Error("InsertGaussUser, inErr: ", inErr)
			return 0
		}
		gu.UserId = userId
	}
	gipu := models.GaussIssuePrUser{OrId: orId, UserId: gu.UserId, Type: 1}
	gipuErr := models.QueryGaussIssuePrUser(&gipu, "OrId", "UserId", "Type")
	if gipuErr != nil {
		gipu.OrId = orId
		gipu.UserId = gu.UserId
		gipu.Type = 1
		gipu.Status = 2
		gipu.CreateTime = common.GetCurTime()
		models.InsertGaussIssuePrUser(&gipu)
	} else {
		gipu.Status = 2
		gipu.UpdateTime = common.GetCurTime()
		upeiuErr := models.UpdateGaussIssuePrUser(&gipu, "Status", "UpdateTime")
		if upeiuErr != nil {
			logs.Error("UpdateGaussIssuePrUser, upeiuErr: ", upeiuErr)
		}
	}
	guid := models.GaussUserIntegDetail{UserId: gu.UserId, OrId: orId, Type: 1}
	guidErr := models.QueryGaussUserIntegDetail(&guid, "UserId", "OrId", "Type")
	if guidErr == nil {
		if guid.IntegralValue != pointValue {
			guic := models.GaussUserIntegCount{UserId: gu.UserId}
			eicErr := models.QueryGaussUserIntegCount(&guic, "UserId")
			if eicErr != nil {
				guic = models.GaussUserIntegCount{UserId: gu.UserId, IntegralValue: pointValue, CreateTime: common.GetCurTime()}
				guicId, ineicErr := models.InsertGaussUserIntegCount(&guic)
				if ineicErr != nil {
					logs.Error("InsertGaussUserIntegCount, ineicErr: ", ineicErr, guicId)
				}
			} else {
				if guic.IntegralValue >= guid.IntegralValue {
					guic.IntegralValue -= guid.IntegralValue
				}
				guic.IntegralValue += pointValue
				guicErr := models.UpdateGaussUserIntegCount(&guic, "IntegralValue")
				if guicErr != nil {
					logs.Error("UpdateGaussUserIntegCount, guicErr: ", guicErr)
				}
			}
			guid.IntegralValue = pointValue
			guidErr := models.UpdateGaussUserIntegDetail(&guid, "IntegralValue")
			if guidErr != nil {
				logs.Error("UpdateGaussUserIntegDetail, guidErr: ", guidErr)
			}
		}
	} else {
		guid = models.GaussUserIntegDetail{UserId: gu.UserId, OrId: orId,
			IntegralValue: pointValue, CreateTime: common.GetCurTime(), Type: 1}
		id, indErr := models.InsertGaussUserIntegDetail(&guid)
		if id > 0 {
			guic := models.GaussUserIntegCount{UserId: gu.UserId}
			guicErr := models.QueryGaussUserIntegCount(&guic, "UserId")
			if guicErr != nil {
				guic = models.GaussUserIntegCount{UserId: gu.UserId,
					IntegralValue: pointValue, CreateTime: common.GetCurTime()}
				guicId, ineicErr := models.InsertGaussUserIntegCount(&guic)
				if ineicErr != nil {
					logs.Error("InsertGaussUserIntegCount, ineicErr: ", ineicErr, guicId)
				}
			} else {
				guic.IntegralValue += pointValue
				guicErr := models.UpdateGaussUserIntegCount(&guic, "IntegralValue")
				if guicErr != nil {
					logs.Error("UpdateGaussUserIntegCount, upicErr: ", guicErr)
				}
			}

		} else {
			logs.Error("InsertGaussUserIntegDetail, indErr:", indErr)
		}
	}
	// After earning points, send a private message
	if pointValue > 0 {
		iss := fmt.Sprintf(IssueGaussPointSend, gitUrl, pointValue)
		SendPrivateLetters(gaussToken, iss, issueCreate)
	} else {
		iss := fmt.Sprintf(IssueGaussPointzSend, gitUrl)
		SendPrivateLetters(gaussToken, iss, issueCreate)
	}
	return gu.UserId
}

// Parse pr comments
func HandleGaussPrComment(payload models.CommentPayload) {
	if payload.PullRequest.Id == 0 || payload.Comment == nil {
		logs.Error("Data error, PullRequest: ", payload.PullRequest, ", PrComment: ", payload.Comment)
		return
	}
	if payload.Comment.User == nil {
		logs.Error("Data error, PrcommentUser: ", payload.Comment.User)
		return
	}
	// The default timeout for receiving hooks
	logs.Info("pr payload.Comment: ", payload.Comment,
		", Number: ", payload.PullRequest.Number, "id: ", payload.PullRequest.Id)
	prNumber := payload.PullRequest.Number  //pr number string
	prId := payload.PullRequest.Id          //pr id int64
	cBody := payload.Comment.Body           //Comment subject
	cuAccount := payload.Comment.User.Login //gitee domain address
	if prNumber == 0 || cuAccount == "" || cBody == "" {
		logs.Error("Data has null values: prNumber, "+
			"cuAccount, cBody: ", prNumber, cuAccount, cBody)
		return
	}
	assignFlag := reviewIsvalid(cuAccount)
	if !assignFlag {
		logs.Error("Invalid comment, discard, body: ", cBody, ", cuAccount: ", cuAccount)
		return
	}
	repoPath := common.TrimString(payload.Repository.Path)
	owner := common.TrimString(payload.Repository.NameSpace)
	gop := models.GaussOriginPr{Owner: owner, RepoPath: repoPath,
		PrId: prId, PrNumber: prNumber}
	eiErr := models.QueryGaussOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gop.OrId == 0 {
		logs.Error("QueryGaussOriginPr, Data does not exist, eiErr: ", eiErr)
		return
	}
	gaussToken := os.Getenv("GITEE_GAUSS_TOKEN")
	issuePointStr := beego.AppConfig.String("gauss::pr_point")
	if strings.HasPrefix(cBody, GaussHighCmd) {
		// first-claimed task
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussHighCmd, gop.PrCreate, gop.GitUrl, gop.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, gop.OrId, pointValue, 2,
				strconv.FormatInt(gop.PrNumber, 10), gop.RepoPath, gop.Owner)
			is := fmt.Sprintf(PrGaussPointComment, gop.PrCreate)
			AddCommentToPr(is, owner, payload.Repository.Path, gaussToken, prNumber)
		}
	} else if strings.HasPrefix(cBody, GaussMediumCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussMediumCmd, gop.PrCreate, gop.GitUrl, gop.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, gop.OrId, pointValue, 2,
				strconv.FormatInt(gop.PrNumber, 10), gop.RepoPath, gop.Owner)
			is := fmt.Sprintf(PrGaussPointComment, gop.PrCreate)
			AddCommentToPr(is, owner, payload.Repository.Path, gaussToken, prNumber)
		}
	} else if strings.HasPrefix(cBody, GaussLowCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussLowCmd, gop.PrCreate, gop.GitUrl, gop.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, gop.OrId, pointValue, 2,
				strconv.FormatInt(gop.PrNumber, 10), gop.RepoPath, gop.Owner)
			is := fmt.Sprintf(PrGaussPointComment, gop.PrCreate)
			AddCommentToPr(is, owner, payload.Repository.Path, gaussToken, prNumber)
		}
	} else if strings.HasPrefix(cBody, GaussZeroCmd) {
		userId, pointValue := AddGaussUserPoints(gaussToken, issuePointStr,
			GaussZeroCmd, gop.PrCreate, gop.GitUrl, gop.OrId)
		if userId > 0 {
			AddGaussPointsComplete(userId, gop.OrId, pointValue, 2,
				strconv.FormatInt(gop.PrNumber, 10), gop.RepoPath, gop.Owner)
			is := fmt.Sprintf(PrGaussPointComment, gop.PrCreate)
			AddCommentToPr(is, owner, payload.Repository.Path, gaussToken, prNumber)
		}
	}
}
