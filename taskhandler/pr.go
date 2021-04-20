package taskhandler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"os"
	"strconv"
	"strings"
)

// UpdatePrAssignee Update pr responsible person
func UpdatePrAssignee(prHook models.PrPayload) {
	prNumber := prHook.PullRequest.Number
	repoPath := common.TrimString(prHook.Repository.Path)
	//Update the person in charge of the issue template
	gop := models.GaussOriginPr{Owner: prHook.Repository.NameSpace, RepoPath: repoPath,
		PrId: prHook.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryGaussOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		logs.Error(gopErr)
		return
	}
	assigneeSlice := []string{}
	if len(prHook.PullRequest.Assignees) > 0 {
		for _, as := range prHook.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.Login)
		}

	} else {
		if len(prHook.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, prHook.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
	}
	upErr := models.UpdateGaussOriginPr(&gop, "PrAssignee")
	if upErr != nil {
		logs.Error(upErr)
	}
}

// AddPr Add pr
func AddHookGaussPr(prData *models.PrPayload, openFlag int) {
	gaussToken := os.Getenv("GITEE_GAUSS_TOKEN")
	prNumber := prData.PullRequest.Number
	repoPath := common.TrimString(prData.Repository.Path)
	owner := common.TrimString(prData.Repository.NameSpace)
	gop := models.GaussOriginPr{Owner: prData.Repository.NameSpace, RepoPath: repoPath,
		PrId: prData.PullRequest.Id, PrNumber: prNumber}
	gopErr := models.QueryGaussOriginPr(&gop, "Owner", "RepoPath", "PrId", "PrNumber")
	if gopErr != nil {
		CreateHookGaussPrOrgData(prData, &gop, 1)
		prId, iprErr := models.InsertGaussOriginPr(&gop)
		if iprErr != nil {
			logs.Error("InsertGaussOriginPr, prId: ", prId, ",iprErr: ", iprErr)
			return
		}
		gop.OrId = prId
	} else {
		updateStr := CreateHookGaussPrOrgData(prData, &gop, 2)
		uprErr := models.UpdateGaussOriginPr(&gop, updateStr...)
		if uprErr != nil {
			logs.Error("UpdateGaussOriginPr, uprErr: ", uprErr)
			return
		}
	}
	if openFlag == 1 {
		// Create user information
		userId := StoreGitGaussUser(prData.PullRequest.User.Login, prData.PullRequest.User.Email)
		if userId > 0 {
			et := GaussIssueUserRecordTp{UserId: userId, OrId: gop.OrId,
				IssueNumber: strconv.FormatInt(gop.PrNumber, 10),
				RepoPath:    gop.RepoPath, Owner: owner, Status: 2, Type: 2}
			GaussIssueUserRecord(et)
			// Create the correspondence between users and pr, as well as user points information
			GaussIssueUser(userId, gop.OrId, strconv.FormatInt(gop.PrNumber, 10), gop.RepoPath, gop.Owner, 2)
			// Calculate the points earned by users
			CreateUserPoints(userId, gop.OrId, 0, 2)
			hdcGaussLabel := beego.AppConfig.String("hdc_gauss_label")
			if len(gop.PrLabel) > 1 && strings.Contains(strings.ToLower(gop.PrLabel), hdcGaussLabel) {
				// Will write issue comments
				igc := fmt.Sprintf(PrGaussComment, prData.PullRequest.User.Login)
				AddCommentToPr(igc, owner, prData.Repository.Path, gaussToken, prData.PullRequest.Number)
				// edit label
				//hdcGuassLabel := beego.AppConfig.String("hdc_gauss_label")
				//EditGaussPrLabel(hdcGuassLabel, gaussToken, owner, gop, prData.PullRequest.Number)
				// Send private message
				igcs := fmt.Sprintf(PrGaussCommentSend, gop.GitUrl)
				SendPrivateLetters(gaussToken, igcs, prData.PullRequest.User.Login)
				assigneeStr := beego.AppConfig.String("gauss::assignee")
				if len(assigneeStr) > 1 {
					assigneeSlice := strings.Split(assigneeStr, ",")
					if len(assigneeSlice) > 0 {
						for _, as := range assigneeSlice {
							igcs := fmt.Sprintf(PrGaussRewiewSend, prData.PullRequest.User.Login, gop.GitUrl)
							SendPrivateLetters(gaussToken, igcs, as)
						}
					}
				}
			}
		}
	}
}

func CreateHookGaussPrOrgData(hi *models.PrPayload, gop *models.GaussOriginPr, flag int) []string {
	updateStr := make([]string, 0)
	gop.PrNumber = hi.PullRequest.Number
	gop.PrId = hi.PullRequest.Id
	prState := common.TrimString(hi.State)
	gop.PrState = prState
	updateStr = append(updateStr, "PrState")
	gop.GitUrl = hi.PullRequest.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	gop.Title = hi.PullRequest.Title
	updateStr = append(updateStr, "Title")
	gop.PrBody = hi.PullRequest.Body
	updateStr = append(updateStr, "PrBody")
	if flag == 1 {
		gop.GrabTime = common.GetCurTime()
	}
	labelStr := ""
	if hi.PullRequest.Labels != nil && len(hi.PullRequest.Labels) > 0 {
		for _, la := range hi.PullRequest.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
		gop.PrLabel = labelStr
		updateStr = append(updateStr, "PrLabel")
	}
	gop.PrCreate = hi.PullRequest.User.Login
	updateStr = append(updateStr, "PrCreate")
	gop.PrUpdate = hi.PullRequest.UpdatedBy.Login
	updateStr = append(updateStr, "PrUpdate")
	assigneeSlice := []string{}
	if len(hi.PullRequest.Assignees) > 0 {
		for _, as := range hi.PullRequest.Assignees {
			assigneeSlice = append(assigneeSlice, as.Login)
		}

	} else {
		if len(hi.PullRequest.Assignee) > 0 {
			assigneeSlice = append(assigneeSlice, hi.PullRequest.Assignee)
		}
	}
	if len(assigneeSlice) > 0 {
		gop.PrAssignee = strings.Join(assigneeSlice, ",")
		updateStr = append(updateStr, "PrAssignee")
	}
	gop.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	gop.RepoPath = hi.Repository.Path
	gop.Owner = hi.Repository.NameSpace
	gop.Status = 1
	updateStr = append(updateStr, "Status")
	gop.TargetBranch = hi.TargetBranch
	updateStr = append(updateStr, "TargetBranch")
	if len(hi.PullRequest.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		gop.CreateTime = hi.PullRequest.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.PullRequest.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		gop.UpdateTime = hi.PullRequest.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.PullRequest.ClosedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.ClosedTime = hi.PullRequest.ClosedAt.String()
		updateStr = append(updateStr, "ClosedTime")
	}
	if len(hi.PullRequest.MergedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		gop.MergedTime = hi.PullRequest.MergedAt.String()
		updateStr = append(updateStr, "MergedTime")
	}
	logs.Info("gop===>", gop)
	return updateStr
}
