package taskhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/util"
	"net/http"
	"regexp"
	"strings"
)

// Get the constant of the original issue
const (
	//GiteOrgInfoURL get gitee org info url
	GiteOrgInfoURL = `https://gitee.com/api/v5/orgs/%v?access_token=%v`
	//GiteOrgReposURL get all repository url
	GiteOrgReposURL = `https://gitee.com/api/v5/orgs/%v/repos?access_token=%v&type=all&page=%v&per_page=%v`
	//GiteRepoIssuesURL get issue list url
	GiteRepoIssuesURL = `https://gitee.com/api/v5/repos/%v/%v/issues?access_token=%v&state=%v&labels=%v&sort=created&direction=desc&page=%v&per_page=%v`
	//GiteRepoBranch get repo branch url
	GiteRepoBranch = `https://gitee.com/api/v5/repos/%v/%v/branches?access_token=%v`
	//RepoInfoURL get repo info url
	RepoInfoURL = "https://api.openeuler.org/pkgmanage/packages/packageInfo?table_name=openEuler_LTS_20.03&pkg_name=%s"
	perPage     = 50
	//IssueType Types of issues crawled
	CIssueType = "HDC-任务打榜赛"
)

const (
	// Notification of claim failure
	IssueClaimFailure     = `@%v, 您认领的任务已经达上限,无法再次领取新的任务.`
	IssueClaimFailureSend = `您认领的任务已经达上限,无法再次领取新的任务, 请先提交已认领的任务或者取消已认领的任务.`
	// Claim the same task notification multiple times
	IssueClaimSameTask = `@%v, 您已认领过当前任务.`
	// Notification of successful claim task
	IssueClaimSuccess     = `@%v, 您已成功认领当前任务, 认领任务>处理任务>提交任务>审核通过>获得积分.`
	IssueClaimSuccessSend = `%v, 您已成功认领当前任务, 认领任务>处理任务>提交任务>审核通过>获得积分.`
	// The task has been preemptively claimed by others
	IssueClaimPree     = `@%v, 您无法认领此任务, 已被他人认领.`
	IssueClaimPreeSend = `%v, 您无法认领此任务, 已被他人认领.`
	// The user cannot claim the current task, the information is wrong
	IssueClaimWrong     = `@%v, 当前无法认领此任务, 稍后重试.`
	IssueClaimWrongSend = `%v, 当前无法认领此任务, 稍后重试.`
	// Submit task
	IssueClaimSubmit         = `@%v, 任务认领者才能提交此任务.`
	IssueClaimReSubmit       = `@%v, 任务已提交,此任务审核者会尽快审核并在私信或者openEuler官网反馈结果.`
	IssueClaimSubmitComplete = `@%v, 任务已审核完成, 请查看私信或者官网获取结果.`
	IssueClaimReSubmitSend   = `%v, 此任务已提交,此任务审核者会尽快审核并在私信或者openEuler官网反馈结果.`
	IssueClaimReReviewerSend = `%v, 认领的任务: %v 已提交, 如果审核通过请关闭issue, 如果审核未通过, 请在issue评论反馈原因, 并删除标签: hdc-task-rewiew`
	// Give up the task
	IssueGiveUpTask        = `@%v, 认领此任务的开发者才能取消此任务.`
	IssueGiveUpFailure     = `@%v, 任务已审核完成, 无法取消此任务.`
	IssueGiveUpSuccess     = `@%v, 您已取消此任务.`
	IssueGiveUpSuccessSend = `%v, 此任务您已取消, 可以继续认领其他任务.`
	IssueStateProc         = `@%v, 此issue状态, 只能由issue责任人改变, 如需改变, 请先变更issue责任人.`
	// After earning points, send a private message
	IssuePointSend            = `您提交的任务: %v 已审核通过, 当前任务获得的积分为: %v分.`
	IssueBlackSend            = `您认领的任务: %v 已取消, 请知晓.`
	IssueBlackClaimFailure    = `@%v, 您无法认领此任务, 请知晓.`
	IssueUnassignClaimCount   = `@%v, 由于您取消任务次数已达到上线, 账号已被冻结, 待账号解冻后才能继续操作.`
	IssueForciGiveUpSuccess   = `@%v, 由于您长期未完成此任务, 系统已为您取消此任务.`
	IssueUncompleteClaimCount = `由于参赛者: @%v 取消任务次数已达到上线, 账号已被冻结, 待账号解冻后才能继续操作.`
)

const (
	// issue
	IssueGaussComment      = `@%v, 感谢您提交此issue, 我们会尽快评估此issue, 经过评审后给出对应的评分, 评分信息请关注openGauss官网或向您发送私信, 谢谢!`
	IssueGaussCommentSend  = `感谢您提交此issue: %v, 我们会尽快评估此issue, 经过评审后给出对应的评分, 评分信息请关注openGauss官网或向您发送私信, 谢谢!`
	IssueGaussRewiewSend   = `参赛者: @%v, 已提交issue: %v, 请尽快在此issue评论区给出对应评分,评分命令: /challenge-high, /challenge-medium, /challenge-low, /challenge-zero`
	IssueGaussPointSend    = `您提交的issue: %v, 已获得积分: %v分, 感谢您的参与, 谢谢!`
	IssueGaussPointzSend   = `您提交的issue: %v, 经过评估未获得积分, 感谢您的参与, 您可以继续提交其他issue或pr, 谢谢!`
	IssueGaussPointComment = `@%v, 感谢您为openGauss社区贡献的issue, 对此issue的评审结果已向您发送私信, 请查收, 如有疑问可以在issue评论区留言, 谢谢!`
	// pr
	PrGaussComment      = `@%v, 感谢您提交此pr, 我们会尽快评估此pr, 经过评审后给出对应的评分, 评分信息请关注openGauss官网或向您发送私信, 谢谢!`
	PrGaussCommentSend  = `感谢您提交此pr: %v, 我们会尽快评估此pr, 经过评审后给出对应的评分, 评分信息请关注openGauss官网或向您发送私信, 谢谢!`
	PrGaussRewiewSend   = `参赛者: @%v, 已提交pr: %v, 请尽快在此pr评论区给出对应评分,评分命令: /challenge-high, /challenge-medium, /challenge-low, /challenge-zero`
	PrGaussPointComment = `@%v, 感谢您为openGauss社区贡献的pr, 对此pr的评审结果已向您发送私信, 请查收, 如有疑问可以在pr评论区留言, 谢谢!`
)

const (
	//IssueRejectState issue state rejected
	IssueRejectState = "rejected"
	//IssueCloseState issue state closed
	IssueCloseState = "closed"
	//IssueProgressState issue  state progressing
	IssueProgressState = "progressing"
	//IssueOpenState issue state open
	IssueOpenState = "open"
)

const (
	GaussHighCmd   = "/challenge-high"
	GaussMediumCmd = "/challenge-medium"
	GaussLowCmd    = "/challenge-low"
	GaussZeroCmd   = "/challenge-zero"
	GaussLabelCmd  = "/hdc-p-challenge"
)

const (
	IssueComment = "IssueCommentEvent"
	PullRequest  = "PullRequestEvent"
	IssueRequest = "IssueEvent"
)

var (
	RegexpEmergencyLevel = regexp.MustCompile(`紧急程度[:：](?s:(.*?))难易程度[:：]`)
	RegexpDegreeDiff     = regexp.MustCompile(`难易程度[:：](?s:(.*?))$`)
	RegexpDigit          = regexp.MustCompile(`([0-9]+)`)
)

type StaticIssueInfo struct {
	WeekIssueCount         int64
	WeekIssueCommentCount  int64
	WeekPullRequestCount   int64
	monthIssueCount        int64
	monthIssueCommentCount int64
	monthPullRequestCount  int64
	TotalIssueCount        int64
	TotalIssueCommentCount int64
	TotalPullRequestCount  int64
}

type StaticIssueTime struct {
	WeekIssueStartTime  string
	WeekIssueEndTime    string
	MonthIssueStartTime string
	MonthIssueEndTime   string
	TotalIssueTime      string
}

//AddCommentToIssue Add a comment to the issue
func AddCommentToIssue(msg, issueNum, owner, repo, access string) {
	url := fmt.Sprintf(`https://gitee.com/api/v5/repos/%v/%v/issues/%v/comments`, owner, repo, issueNum)
	param := fmt.Sprintf(`{"access_token": "%s","body":"%s"}`, access, msg)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Add issue comment back:", res)
}

//AddCommentToPr Add a comment to the pr
func AddCommentToPr(msg, owner, repo, access string, prNumber int64) {
	url := fmt.Sprintf(`https://gitee.com/api/v5/repos/%v/%v/pulls/%v/comments`, owner, repo, prNumber)
	param := fmt.Sprintf(`{"access_token": "%s","body":"%s"}`, access, msg)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Add pr comment back:", res)
}

//SendPrivateLetters Send a private message to a gitee user
func SendPrivateLetters(access, content, useName string) {
	url := "https://gitee.com/api/v5/notifications/messages"
	param := fmt.Sprintf(`{"access_token":"%s","username":"%s","content":"%s"}`, access, useName, content)
	res, err := util.HTTPPost(url, param)
	if err != nil {
		logs.Error(err)
	}
	logs.Info("Send private message:", res)
}

func UpdateIssueLabels(token, repo, issueNum, owner, label string) bool {
	labelStr := label
	labelSlice := strings.Split(label, ",")
	if len(labelSlice) > 0 {
		laSlice := []string{}
		for _, la := range labelSlice {
			laSlice = append(laSlice, fmt.Sprintf("\"%v\"", la))
		}
		if len(laSlice) > 0 {
			labelStr = strings.Join(laSlice, ",")
		}
	}
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%v/%v/issues/%v/labels?access_token=%v", owner, repo, issueNum, token)
	reqBody := fmt.Sprintf("[%v]", labelStr)
	logs.Info("UpdateIssueLabels, reqBody: ", reqBody)
	resp, err := util.HTTPPut(url, reqBody)
	if err != nil {
		logs.Error("UpdateIssueLabels, Failed to update label, url: ", url, ", err: ", err)
		return false
	}
	if len(resp) > 0 {
		if _, ok := resp[0]["id"]; !ok {
			logs.Error("UpdateIssueLabels, Failed to update label, err: ", ok, ", url: ", url)
			return false
		}
	}
	return true
}

//ChangePrLabel update  pr label
func ChangePrLabel(token, repo, owner, label string, prNumber int64) bool {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/pulls/%v", owner, repo, prNumber)
	param := struct {
		AccessToken string `json:"access_token"`
		Label       string `json:"labels"`
	}{token, label}
	pj, err := json.Marshal(&param)
	if err != nil {
		logs.Error(err)
		return false
	}
	return UpdateGiteIssue(url, pj)
}

//UpdateGiteIssue update gitee issue
func UpdateGiteIssue(url string, param []byte) bool {
	read := bytes.NewReader(param)
	req, err := http.NewRequest(http.MethodPatch, url, read)
	if err != nil {
		logs.Error(err)
		return false
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error(err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

// issue status transition
func IssueStateRev(issueState string) (statusName string) {
	if issueState != "" && len(issueState) > 1 {
		if issueState == "待办的" || issueState == "开启的" ||
			strings.ToLower(issueState) == "open" {
			statusName = "open"
		} else if issueState == "进行中" || strings.ToLower(issueState) == "started" ||
			strings.ToLower(issueState) == "progressing" {
			statusName = "progressing"
		} else if issueState == "已完成" || strings.ToLower(issueState) == "closed" {
			statusName = "closed"
		} else if issueState == "已拒绝" || strings.ToLower(issueState) == "rejected" {
			statusName = "rejected"
		} else if issueState == "已挂起" || strings.ToLower(issueState) == "suspended" {
			statusName = "suspended"
		} else {
			statusName = issueState
		}
	}
	return
}

//Get public updates from users
func GetUserPublicUpEvents(userName, accessToken, ownerList string, prevId, limit int64, sTime StaticIssueTime) StaticIssueInfo {
	localPrevId := prevId
	localLimit := limit
	sii := StaticIssueInfo{}
	for {
		url := ""
		if prevId > 0 {
			url = fmt.Sprintf("https://gitee.com/api/v5/users/%v/events/public?access_token=%v&prev_id=%v&limit=%v",
				userName, accessToken, localPrevId, localLimit)
		} else {
			url = fmt.Sprintf("https://gitee.com/api/v5/users/%v/events/public?access_token=%v&limit=%v",
				userName, accessToken, localLimit)
		}
		publicData, err := util.HTTPGet(url)
		if err == nil && publicData != nil {
			for _, value := range publicData {
				if _, ok := value["id"]; !ok {
					logs.Error("publicData, err: ", ok, "url: ", url)
					continue
				}
				staticIssueComment(value, ownerList, userName, sTime, &sii)
			}
		} else {
			break
		}
	}
	return sii
}

func staticIssueComment(value map[string]interface{},
	ownerList, userName string, sTime StaticIssueTime, sii *StaticIssueInfo) {
	createdAt := value["created_at"].(string)
	ct := int64(0)
	wst := int64(0)
	wet := int64(0)
	mst := int64(0)
	met := int64(0)
	tst := int64(0)
	if len(createdAt) > 0 {
		if len(createdAt) > 19 {
			ct = util.TimeStrToInt(createdAt[:19], "2006-01-02T15:04:05")
		} else {
			ct = util.TimeStrToInt(createdAt, "2006-01-02T15:04:05")
		}
	}
	if len(sTime.WeekIssueStartTime) > 0 {
		if len(sTime.WeekIssueStartTime) <= 10 {
			sTime.WeekIssueStartTime = sTime.WeekIssueStartTime + " 00:00:00"
		}
		wst = util.TimeStrToInt(sTime.WeekIssueStartTime, "2006-01-02 15:04:05")

	}
	if len(sTime.WeekIssueEndTime) > 0 {
		if len(sTime.WeekIssueEndTime) <= 10 {
			sTime.WeekIssueEndTime = sTime.WeekIssueEndTime + " 00:00:00"
		}
		wet = util.TimeStrToInt(sTime.WeekIssueEndTime, "2006-01-02 15:04:05")
	}
	if len(sTime.MonthIssueStartTime) > 0 {
		if len(sTime.MonthIssueStartTime) <= 10 {
			sTime.MonthIssueStartTime = sTime.MonthIssueStartTime + " 00:00:00"
		}
		mst = util.TimeStrToInt(sTime.MonthIssueStartTime, "2006-01-02 15:04:05")

	}
	if len(sTime.MonthIssueEndTime) > 0 {
		if len(sTime.MonthIssueEndTime) <= 10 {
			sTime.MonthIssueEndTime = sTime.MonthIssueEndTime + " 00:00:00"
		}
		met = util.TimeStrToInt(sTime.MonthIssueEndTime, "2006-01-02 15:04:05")
	}
	if len(sTime.TotalIssueTime) > 0 {
		if len(sTime.TotalIssueTime) <= 10 {
			sTime.TotalIssueTime = sTime.TotalIssueTime + " 00:00:00"
		}
		tst = util.TimeStrToInt(sTime.TotalIssueTime, "2006-01-02 15:04:05")

	}
	repoMap := value["repo"].(map[string]interface{})
	namespaceMap := repoMap["namespace"].(map[string]interface{})
	path := namespaceMap["path"].(string)
	pathFlag := false
	ownSlice := strings.Split(ownerList, ",")
	for _, os := range ownSlice {
		if path == os {
			pathFlag = true
		}
	}
	if !pathFlag {
		logs.Error("path: ", path, ",Not in the current organization and not participating in statistics")
		return
	}
	switch value["type"].(string) {
	case IssueComment:
		payloadMap := value["payload"].(map[string]interface{})
		commentMap := payloadMap["comment"].(map[string]interface{})
		userMap := commentMap["user"].(map[string]interface{})
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekIssueCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthIssueCount += 1
			}
			if tst <= ct {
				sii.TotalIssueCount += 1
			}
		}
	case PullRequest:
		payloadMap := value["payload"].(map[string]interface{})
		headMap := payloadMap["head"].(map[string]interface{})
		userMap := headMap["user"].(map[string]interface{})
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekIssueCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthIssueCount += 1
			}
			if tst <= ct {
				sii.TotalIssueCount += 1
			}
		}
	case IssueRequest:
		payloadMap := value["payload"].(map[string]interface{})
		userMap := payloadMap["user"].(map[string]interface{})
		login := userMap["login"].(string)
		if userName == login {
			if wst <= ct && ct <= wet {
				sii.WeekIssueCount += 1
			}
			if mst <= ct && ct <= met {
				sii.monthIssueCount += 1
			}
			if tst <= ct {
				sii.TotalIssueCount += 1
			}
		}
	}
	return
}
