package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type EulerUser struct {
	UserId     int64  `orm:"pk;auto;column(user_id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitUserId  string `orm:"size(512);column(git_id);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:正常用户;2:异常用户;3:用户已无法领取任务"`
	CreateTime string `orm:"size(32);column(create_time);"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

//GiteOriginIssue Issues that already exist on Code Cloud
type EulerOriginIssue struct {
	OrId           int64  `orm:"pk;auto;column(or_id)"`
	IssueId        int64  `orm:"column(issue_id);unique" description:"issue id,gitee上唯一"`
	GitUrl         string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	IssueNumber    string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	IssueState     string `orm:"column(issue_state);size(50)" description:"issue 状态"`
	IssueType      string `orm:"column(issue_type);size(64)" description:"issue 类型"`
	Title          string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	IssueBody      string `orm:"column(issue_body);null;type(text)" description:"issue 主体"`
	IssueLabel     string `orm:"size(512);column(issue_label)" description:"issue标签, hdc-task;hdc-task-assign;hdc-task-rewiew"`
	IssueCreate    string `orm:"column(issue_create);size(256)" description:"issue issue创建人"`
	IssueAssignee  string `orm:"column(issue_assignee);size(256)" description:"issue 责任人,必填"`
	RepoPath       string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	RepoUrl        string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner          string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status         int8   `orm:"default(0);column(status)" description:"0:正常;1:已删除"`
	IssueStateName string `orm:"size(50);column(issue_state_name)" description:"issue 中文状态"`
	EmergencyLevel string `orm:"size(32);column(emerg_level)" description:"紧急程度 1：高;2:中;3:低"`
	EmergencyValue int    `orm:"default(3);column(emerg_value)" description:"紧急程度 1：高;2:中;3:低"`
	DifficultLevel string `orm:"size(32);column(diff_level)" description:"难易程度 1：高;2:中;3:低"`
	DifficultValue int    `orm:"default(3);column(diff_value)" description:"难易程度 1：高;2:中;3:低"`
	CreateTime     string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime     string `orm:"size(32);column(update_time);null"`
	FinishedTime   string `orm:"size(32);column(finished_time);null"`
	DeleteTime     string `orm:"size(32);column(delete_time);null"`
	GrabTime       string `orm:"size(32);column(grab_time)" description:"记录当前issue抓取的时间"`
}

// Individual claiming the issue task pool
type EulerIssueUser struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	UserId      int64  `orm:"index;column(user_id)"`
	OrId        int64  `orm:"unique;column(or_id)"`
	IssueNumber string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	RepoPath    string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner       string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	SendEmail   int8   `orm:"default(1);column(send_email)" description:"1:未发送邮件; 2:已发送邮件"`
	Status      int8   `orm:"default(1);column(status)" description:"1:认领中;2:已完成; 3:任务已提交"`
	AssignTime  string `orm:"size(32);column(assign_time);null" description:"赛题认领超时日期"`
	CreateTime  string `orm:"size(32);column(create_time)"`
	UpdateTime  string `orm:"size(32);column(update_time);null"`
	DeleteTime  string `orm:"size(32);column(delete_time);null"`
}

type EulerIssueUserRecord struct {
	Id          int64  `orm:"pk;auto;column(id)"`
	UserId      int64  `orm:"index;column(user_id)"`
	OrId        int64  `orm:"index;column(or_id)"`
	IssueNumber string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	RepoPath    string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner       string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status int8 `orm:"default(1);column(status)" description:"1:已认领;2:取消认领;3:完成提交;4:审核通过;
						5:认领超额; 6:重复认领; 7: 已被他人认领;8:提交他人任务; 9:重复完成提交; 10: 他人取消任务, 11: 任务完成,取消任务失败,
						12: 删除issue, 13: 黑名单用户认领失败, 14: 取消任务次数已达上线,不能再次认领任务"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type EulerIssueUserComplate struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"index;column(user_id)"`
	OrId          int64  `orm:"index;column(or_id)"`
	IssueNumber   string `orm:"column(issue_num);size(50);" description:"issue 编号"`
	RepoPath      string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner         string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status        int8   `orm:"default(1);column(status)" description:"1:已完成"`
	IntegralValue int64  `orm:"column(integral_value)" description:"已获得多少积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

type EulerUserIntegCount struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"unique;column(user_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"用户获得总的积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

type EulerUserIntegDetail struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id)"`
	OrId          int64  `orm:"column(or_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"获得的积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

type GaussUser struct {
	UserId     int64  `orm:"pk;auto;column(user_id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitUserId  string `orm:"size(512);column(git_id);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:正常用户;2:异常用户;3:用户已无法领取任务"`
	CreateTime string `orm:"size(32);column(create_time);"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

//GiteOriginIssue Issues that already exist on Code Cloud
type GaussOriginIssue struct {
	OrId           int64  `orm:"pk;auto;column(or_id)"`
	IssueId        int64  `orm:"column(issue_id);unique" description:"issue id,gitee上唯一"`
	GitUrl         string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	IssueNumber    string `orm:"column(issue_num);size(50)" description:"issue 编号"`
	IssueState     string `orm:"column(issue_state);size(50)" description:"issue 状态"`
	IssueType      string `orm:"column(issue_type);size(64)" description:"issue 类型"`
	Title          string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	IssueBody      string `orm:"column(issue_body);null;type(text)" description:"issue 主体"`
	IssueLabel     string `orm:"size(512);column(issue_label)" description:"issue标签, hdc-p-challenge"`
	IssueCreate    string `orm:"column(issue_create);size(256)" description:"issue issue创建人"`
	IssueAssignee  string `orm:"column(issue_assignee);size(256)" description:"issue 责任人,必填"`
	RepoPath       string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	RepoUrl        string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner          string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status         int8   `orm:"default(0);column(status)" description:"0:正常;1:已删除"`
	IssueStateName string `orm:"size(50);column(issue_state_name)" description:"issue 中文状态"`
	CreateTime     string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime     string `orm:"size(32);column(update_time);null"`
	FinishedTime   string `orm:"size(32);column(finished_time);null"`
	DeleteTime     string `orm:"size(32);column(delete_time);null"`
	GrabTime       string `orm:"size(32);column(grab_time)" description:"记录当前issue抓取的时间"`
}

// Individual claiming the issue task pool
type GaussIssuePrUser struct {
	Id         int64  `orm:"pk;auto;column(id)"`
	UserId     int64  `orm:"index;column(user_id)"`
	OrId       int64  `orm:"column(or_id)"`
	Number     string `orm:"column(num);size(50);" description:"issue 编号或者pr编号"`
	RepoPath   string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner      string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	SendEmail  int8   `orm:"default(1);column(send_email)" description:"1:未发送邮件; 2:已发送邮件"`
	Status     int8   `orm:"default(1);column(status)" description:"1:创建中;2:已完成"`
	Type       int8   `orm:"default(1);column(data_type)" description:"1:issue;2:pr"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type GaussIssueUserRecord struct {
	Id         int64  `orm:"pk;auto;column(id)"`
	UserId     int64  `orm:"index;column(user_id)"`
	OrId       int64  `orm:"index;column(or_id)"`
	Number     string `orm:"column(num);size(50);" description:"issue 编号或者pr编号"`
	RepoPath   string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner      string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status     int8   `orm:"default(1);column(status)" description:"1:创建issue; 2:创建pr;3:删除issue;4:关闭pr"`
	Type       int8   `orm:"default(1);column(data_type)" description:"1:issue;2:pr"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type GaussIssuePrComplate struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"index;column(user_id)"`
	OrId          int64  `orm:"index;column(or_id)"`
	Number        string `orm:"column(num);size(50);" description:"issue 编号或者pr编号"`
	RepoPath      string `orm:"column(issue_repo);size(512)" description:"仓库空间地址"`
	Owner         string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	Status        int8   `orm:"default(1);column(status)" description:"1:已完成"`
	IntegralValue int64  `orm:"column(integral_value)" description:"已获得多少积分"`
	Type          int8   `orm:"default(1);column(data_type)" description:"1:issue;2:pr"`
	CreateTime    string `orm:"size(32);column(create_time)"`
	UpdateTime    string `orm:"size(32);column(update_time);null"`
	DeleteTime    string `orm:"size(32);column(delete_time);null"`
}

type GaussUserIntegCount struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"unique;column(user_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"用户获得总的积分"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

type GaussUserIntegDetail struct {
	Id            int64  `orm:"pk;auto;column(id)"`
	UserId        int64  `orm:"column(user_id)"`
	OrId          int64  `orm:"column(or_id)"`
	IntegralValue int64  `orm:"column(integral_value)" description:"获得的积分"`
	Type          int8   `orm:"default(1);column(data_type)" description:"1:issue;2:pr"`
	CreateTime    string `orm:"size(32);column(create_time)"`
}

//GaussOriginPr pr that already exist on Code Cloud
type GaussOriginPr struct {
	OrId         int64  `orm:"pk;auto;column(or_id)"`
	PrId         int64  `orm:"column(Pr_id);unique" description:"pr id,gitee上唯一"`
	GitUrl       string `orm:"column(git_url);size(512)" description:"issue gitee 链接"`
	PrNumber     int64  `orm:"column(pr_num)" description:"pr 编号"`
	PrState      string `orm:"column(pr_state);size(50)" description:"pr 状态"`
	Title        string `orm:"column(issue_title);type(text);null" description:"issue 标题"`
	PrBody       string `orm:"column(pr_body);null;type(text)" description:"pr 主体"`
	PrLabel      string `orm:"size(512);column(pr_label)" description:"pr标签, hdc-p-challenge"`
	PrCreate     string `orm:"column(pr_create);size(256)" description:"pr 创建人"`
	PrUpdate     string `orm:"column(pr_update);size(256)" description:"pr 更新人"`
	PrAssignee   string `orm:"column(pr_assignee);size(256)" description:"pr 责任人,必填"`
	RepoPath     string `orm:"column(pr_repo);size(512)" description:"仓库空间地址"`
	RepoUrl      string `orm:"column(repo_url);type(text)" description:"仓库码云地址链接"`
	Owner        string `orm:"column(owner_repo);size(64)" description:"仓库所在组织"`
	TargetBranch string `orm:"column(target_branch);size(64)" description:"pr提交的分支"`
	Status       int8   `orm:"default(0);column(status)" description:"0:正常;1:已删除"`
	CreateTime   string `orm:"size(32);column(create_time);null" description:"issue创建时间"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	ClosedTime   string `orm:"size(32);column(closed_time);null"`
	MergedTime   string `orm:"size(32);column(merged_time);null"`
	GrabTime     string `orm:"size(32);column(grab_time)" description:"记录当前pr的时间"`
}

type EmailList struct {
	Id        int64  `orm:"pk;auto"`
	EmailName string `orm:"size(512);column(email_name);null" description:"收件人名称"`
	EmailType int8   `orm:"column(email_type);default(1)" description:"1:openEuler;2:openGauss;3:都发"`
	SendType  int8   `orm:"column(send_type);default(1)" description:"1:收件人;2:抄送人"`
}

// Individual claiming the issue task pool
type EulerBlackUser struct {
	Id         int64  `orm:"pk;auto;column(id)"`
	UserName   string `orm:"size(256);column(user_name);null"`
	GitUserId  string `orm:"size(512);column(git_id);unique"`
	EmailAddr  string `orm:"size(256);column(email_address)"`
	Status     int8   `orm:"default(1);column(status)" description:"1:待释放任务；2:已释放任务"`
	CreateTime string `orm:"size(32);column(create_time)"`
	UpdateTime string `orm:"size(32);column(update_time);null"`
	DeleteTime string `orm:"size(32);column(delete_time);null"`
}

type EulerUnassignUser struct {
	Id           int64  `orm:"pk;auto;column(id)"`
	UserId       int64  `orm:"index;column(user_id)"`
	GitUserId    string `orm:"size(512);column(git_id);unique"`
	CountValue   int8   `orm:"default(0);column(count_val)" description:"用户每个月领取任务取消次数"`
	UnassignTime string `orm:"size(32);column(unassign_time);null" description:"冻结结束时间"`
	CreateTime   string `orm:"size(32);column(create_time);"`
	UpdateTime   string `orm:"size(32);column(update_time);null"`
	DeleteTime   string `orm:"size(32);column(delete_time);null"`
}

func CreateDb() bool {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("config init error:", err)
		return false
	}
	prefix := BConfig.String("mysql::dbprefix")
	InitdbType, _ := beego.AppConfig.Int("initdb")
	if InitdbType == 1 {
		orm.RegisterModelWithPrefix(prefix, new(EulerUser),
			new(EulerOriginIssue), new(EulerIssueUser),
			new(EulerIssueUserRecord), new(EulerIssueUserComplate),
			new(EulerUserIntegCount), new(EulerUserIntegDetail),
			new(GaussUser), new(GaussOriginIssue),
			new(GaussIssuePrUser), new(GaussIssueUserRecord),
			new(GaussIssuePrComplate), new(GaussUserIntegCount),
			new(GaussUserIntegDetail), new(GaussOriginPr),
			new(EmailList), new(EulerBlackUser),
			new(EulerUnassignUser),
		)
		logs.Info("table create success!")
		errosyn := orm.RunSyncdb("default", false, true)
		if errosyn != nil {
			logs.Error(errosyn)
		}
	}
	return true
}
