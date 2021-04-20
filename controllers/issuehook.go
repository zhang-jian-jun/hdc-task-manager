package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"hdc-task-manager/taskhandler"
	"strings"
)

var (
	//GiteeUserAgent gitee hook request flag
	GiteeUserAgent = "git-oschina-hook"
	//XGiteeToken password or sign
	XGiteeToken = "X-Gitee-Token"
	//XGIteeEventType webhook event type
	XGIteeEventType = "X-Gitee-Event"
	//NoteHookType type of comment
	NoteHookType = "Note Hook"
	//PullReqHookType type of pull request
	PullReqHookType = "Merge Request Hook"
	//PushTagHookType type of push or tag
	PushTagHookType = "Tag Push Hook"
	//IssueHookType type of issue
	IssueHookType = "Issue Hook"
)

// Create data into db according to hook callback
//HookEventControllers gitee hook callback
type HookEventControllers struct {
	beego.Controller
}

//Post handle gitee webhook
// @router / [post]
func (c *HookEventControllers) Post() {
	if ok := c.isLegitimateHookEvent(); !ok {
		c.Ctx.ResponseWriter.WriteHeader(406)
		c.Ctx.WriteString("Illegal incident, discarded")
		return
	}
	eventType := c.Ctx.Request.Header.Get(XGIteeEventType)
	c.Ctx.ResponseWriter.WriteHeader(200)
	c.Ctx.WriteString("Event received: " + eventType)
	switch eventType {
	case NoteHookType: //handle comment hook data
		c.handleNoteDate()
	case PullReqHookType:
		c.handlePullReq()
	case IssueHookType:
		c.handleIssue()
	case PushTagHookType:
		c.handlePushTag()
	default:
		logs.Info(eventType)
	}
}

//isLegitimateHookEvent according to gitee doc judge
func (c *HookEventControllers) isLegitimateHookEvent() (ok bool) {
	ok = true
	//judge user agent
	uAgent := c.Ctx.Request.Header.Get("User-Agent")
	if uAgent != GiteeUserAgent {
		ok = false
	}
	ctType := c.Ctx.Request.Header.Get("Content-Type")
	if "application/json" != ctType {
		ok = false
	}
	//judge hook password
	xToken := c.Ctx.Request.Header.Get(XGiteeToken)
	//logs.Info(xToken)
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	if xToken != hookPwd {
		logs.Error("hookPwd Err, xToken: ", xToken)
	}
	return
}

func (c *HookEventControllers) handleNoteDate() {
	var hookNote models.CommentPayload
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &hookNote)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := hookNote.Comment.User.Login
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
		if cuAccount == "opengauss-bot" {
			logs.Error("opengauss-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	hookNote.Password = common.TrimString(hookNote.Password)
	hookPwd = common.TrimString(hookPwd)
	if hookNote.Action == "comment" && hookNote.NoteableType == "Issue" && hookNote.Password == hookPwd {
		logs.Info(string(c.Ctx.Input.RequestBody))
		//handle issue comment
		go taskhandler.HandleIssueComment(hookNote)
	}
}

func (c *HookEventControllers) handlePullReq() {

}

func (c *HookEventControllers) handlePushTag() {

}

func (c *HookEventControllers) handleIssue() {
	logs.Info(string(c.Ctx.Input.RequestBody))
	issueHook := models.IssuePayload{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &issueHook)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := issueHook.Sender.Login
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
		if cuAccount == "opengauss-bot" {
			logs.Error("opengauss-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	issueHook.Password = common.TrimString(issueHook.Password)
	if issueHook.Password != hookPwd {
		logs.Error("Hook callback pwd verification error, hook: ", issueHook)
		return
	}
	if issueHook.Action == "assign" {
		//Update the person in charge of the issue template
		eoi := models.EulerOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eiErr != nil {
			logs.Error(err)
			return
		}
		eoi.IssueAssignee = issueHook.Assignee.Login
		upErr := models.UpdateEulerOriginIssue(&eoi, "IssueAssignee")
		if upErr != nil {
			logs.Error(upErr)
		}
	}
	if issueHook.Action == "state_change" {
		//handle issue state change
		err = taskhandler.HandleIssueStateChange(&issueHook)
		if err != nil {
			logs.Error(err)
			return
		}
	}
	if issueHook.Action == "open" {
		eoi := models.EulerOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId > 0 {
			logs.Error(eiErr, ",eoi: ", eoi)
			return
		}
		taskhandler.AddHookIssue(&issueHook)
	}
	if issueHook.Action == "delete" {
		taskhandler.DelHookIssue(&issueHook)
	}
}

// Create data into db according to hook callback
//HookEventControllers gitee hook callback
type GaussHookEventControllers struct {
	beego.Controller
}

//Post handle gitee webhook
// @router / [post]
func (c *GaussHookEventControllers) Post() {
	if ok := c.isLegitimateHookEvent(); !ok {
		c.Ctx.ResponseWriter.WriteHeader(406)
		c.Ctx.WriteString("Illegal incident, discarded")
		return
	}
	eventType := c.Ctx.Request.Header.Get(XGIteeEventType)
	logs.Info(string(c.Ctx.Input.RequestBody), "\n", "eventTyped: ", eventType)
	c.Ctx.ResponseWriter.WriteHeader(200)
	c.Ctx.WriteString("Event received: " + eventType)
	switch eventType {
	case NoteHookType: //handle comment hook data
		c.handleNoteDate()
	case PullReqHookType:
		c.handlePullReq()
	case IssueHookType:
		c.handleIssue()
	case PushTagHookType:
		c.handlePushTag()
	default:
		logs.Info(eventType)
	}
}

//isLegitimateHookEvent according to gitee doc judge
func (c *GaussHookEventControllers) isLegitimateHookEvent() (ok bool) {
	ok = true
	//judge user agent
	uAgent := c.Ctx.Request.Header.Get("User-Agent")
	if uAgent != GiteeUserAgent {
		ok = false
	}
	ctType := c.Ctx.Request.Header.Get("Content-Type")
	if "application/json" != ctType {
		ok = false
	}
	//judge hook password
	xToken := c.Ctx.Request.Header.Get(XGiteeToken)
	//logs.Info(xToken)
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	if xToken != hookPwd {
		logs.Error("hookPwd Err, xToken: ", xToken)
	}
	return
}

func (c *GaussHookEventControllers) handleIssue() {
	issueHook := models.IssuePayload{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &issueHook)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := issueHook.Sender.Login
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
		if cuAccount == "opengauss-bot" {
			logs.Error("opengauss-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	issueHook.Password = common.TrimString(issueHook.Password)
	if issueHook.Password != hookPwd {
		logs.Error("Hook callback pwd verification error, hook: ", issueHook)
		return
	}
	gaussOwner := beego.AppConfig.String("repo::gauss_owner")
	issueHook.Repository.NameSpace = common.TrimString(issueHook.Repository.NameSpace)
	if gaussOwner != issueHook.Repository.NameSpace {
		logs.Error("This hook does not belong to the current organization:, "+
			"owner: ", issueHook.Repository.NameSpace, gaussOwner)
		return
	}
	labelStr := ""
	if issueHook.Issue.Labels != nil && len(issueHook.Issue.Labels) > 0 {
		for _, la := range issueHook.Issue.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	if len(labelStr) > 1 {
		labelStr = strings.ToLower(labelStr)
	}
	hdcLabels := beego.AppConfig.String("hdc_gauss_label")
	if labelStr == "" || !strings.Contains(labelStr, hdcLabels) {
		logs.Error("Label error, labelStr: ", labelStr, ", hdcLabels: ", hdcLabels)
	}
	if issueHook.Action == "assign" {
		//Update the person in charge of the issue template
		goi := models.GaussOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eiErr != nil {
			logs.Error(eiErr)
			return
		}
		goi.IssueAssignee = issueHook.Assignee.Login
		upErr := models.UpdateGaussOriginIssue(&goi, "IssueAssignee")
		if upErr != nil {
			logs.Error(upErr)
		}
	}
	if issueHook.Action == "state_change" {
		//handle issue state change
		err = taskhandler.HandleGaussIssueStateChange(&issueHook)
		if err != nil {
			logs.Error(err)
			return
		}
	}
	if issueHook.Action == "open" {
		goi := models.GaussOriginIssue{Owner: issueHook.Repository.NameSpace, RepoPath: issueHook.Repository.Path,
			IssueId: issueHook.Issue.Id, IssueNumber: issueHook.Iid}
		eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if goi.OrId > 0 {
			logs.Error("eiErr: ", eiErr, ",Data does not need to be regenerated, goi: ", goi)
			return
		}
		taskhandler.AddHookGaussIssue(&issueHook)
	}
	if issueHook.Action == "delete" {
		taskhandler.DelHookGaussIssue(&issueHook)
	}
}

func (c *GaussHookEventControllers) handleNoteDate() {
	var hookNote models.CommentPayload
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &hookNote)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := hookNote.Comment.User.Login
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
		if cuAccount == "opengauss-bot" {
			logs.Error("opengauss-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	gaussOwner := beego.AppConfig.String("repo::gauss_owner")
	hookNote.Password = common.TrimString(hookNote.Password)
	hookPwd = common.TrimString(hookPwd)
	if hookNote.Action == "comment" && hookNote.NoteableType == "Issue" && hookNote.Password == hookPwd &&
		gaussOwner == common.TrimString(hookNote.Repository.NameSpace) {
		//handle issue comment
		go taskhandler.HandleGaussIssueComment(hookNote)
	}
	if hookNote.Action == "comment" && hookNote.NoteableType == "PullRequest" && hookNote.Password == hookPwd &&
		gaussOwner == common.TrimString(hookNote.Repository.NameSpace) {
		//handle pr comment
		go taskhandler.HandleGaussPrComment(hookNote)
	}
}

func (c *GaussHookEventControllers) handlePullReq() {
	prHook := models.PrPayload{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &prHook)
	if err != nil {
		logs.Error(err)
		return
	}
	cuAccount := prHook.Sender.Login
	if cuAccount != "" && len(cuAccount) > 1 {
		if cuAccount == "openeuler-ci-bot" {
			logs.Error("openeuler-ci-bot, Ignore this comment")
			return
		}
		if cuAccount == "opengauss-bot" {
			logs.Error("opengauss-bot, Ignore this comment")
			return
		}
	}
	hookPwd := beego.AppConfig.String("hook::hookpwd")
	prHook.Password = common.TrimString(prHook.Password)
	if prHook.Password != hookPwd {
		logs.Error("Hook callback pwd verification error, hook: ", prHook)
		return
	}
	gaussOwner := beego.AppConfig.String("repo::gauss_owner")
	prHook.Repository.NameSpace = common.TrimString(prHook.Repository.NameSpace)
	if gaussOwner != prHook.Repository.NameSpace {
		logs.Error("This hook does not belong to the current organization:, "+
			"owner: ", prHook.Repository.NameSpace, gaussOwner)
		return
	}
	labelStr := ""
	if prHook.PullRequest.Labels != nil && len(prHook.PullRequest.Labels) > 0 {
		for _, la := range prHook.PullRequest.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	if len(labelStr) > 1 {
		labelStr = strings.ToLower(labelStr)
	}
	hdcLabels := beego.AppConfig.String("hdc_gauss_label")
	if labelStr == "" || !strings.Contains(labelStr, hdcLabels) {
		logs.Error("Label error, labelStr: ", labelStr, ", hdcLabels: ", hdcLabels)
	}
	if prHook.Action == "assign" {
		taskhandler.UpdatePrAssignee(prHook)
	}
	//if issueHook.Action == "state_change" {
	//	//handle issue state change
	//	err = taskhandler.HandleGaussIssueStateChange(&issueHook)
	//	if err != nil {
	//		logs.Error(err)
	//		return
	//	}
	//}
	if prHook.Action == "open" {
		taskhandler.AddHookGaussPr(&prHook, 1)
	}
	if prHook.Action == "update" {
		taskhandler.AddHookGaussPr(&prHook, 2)
	}
	if prHook.Action == "closed" {
		//taskhandler.DelHookGaussIssue(&issueHook)
	}
}

func (c *GaussHookEventControllers) handlePushTag() {

}
