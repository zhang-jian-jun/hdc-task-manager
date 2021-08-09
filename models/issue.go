package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//HookUser gitee user model
type HookUser struct {
	Id        int64
	Login     string //Same as username
	Name      string //User's nickname
	Email     string //User's mailbox
	UserName  string `json:"username"`   //User's code cloud personal space address
	AvatarUrl string `json:"avatar_url"` //profile picture
	SiteAdmin bool   `json:"site_admin"` //Is it an administrator

}

//HookIssue gitee issue model
type HookIssue struct {
	Id           int64
	Number       string
	Title        string
	Labels       []IssueLabel `json:"labels"`
	State        string       `json:"state"`
	HtmlUrl      string       `json:"html_url"` //Comment on the url on the code cloud
	Body         string       `json:"body"`
	User         HookUser
	Assignee     HookUser
	Repository   Repository
	CreateAt     time.Time `json:"created_at"`
	UpdateAt     time.Time `json:"updated_at"`
	FinishedAt   time.Time `json:"finished_at"`
	IssueType    string    `json:"issue_type"`
	SecurityHole bool      `json:"security_hole"`
	IssueState   string    `json:"issue_state"`
	StateName    string    `json:"state_name"`
	TypeName     string    `json:"type_name"`
}

//IssueLabel issue label
type IssueLabel struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

//Repository gitee repository model
type Repository struct {
	Id        int64
	FullName  string `json:"full_name"`
	Url       string `json:"url"`
	Path      string `json:"path"`
	Name      string
	NameSpace NameSpace `json:"namespace"`
	Owner     HookUser
	Private   bool
	public    bool
}

//HookRepository gitee repository model
type HookRepository struct {
	Id        int64
	FullName  string `json:"full_name"`
	Url       string `json:"url"`
	Path      string `json:"path"`
	NameSpace string `json:"namespace"`
	Name      string
	Owner     HookUser
	Private   bool
	public    bool
}

type NameSpace struct {
	Id   int64
	Path string `json:"path"`
	Name string `json:"name"`
}

type IssuePayload struct {
	HookId      int64     `json:"hook_id"`   //  Hook id.
	HookUrl     string    `json:"hook_url"`  // route
	HookName    string    `json:"hook_name"` // issue_hooks。
	Password    string    `json:"password"`  // Hook code
	Action      string    //issue status
	Issue       HookIssue //issue information
	Sender      HookUser  //The user information that triggered the hook.
	TargetUser  HookUser  `json:"target_user"` //User information entrusted to handle the issue
	User        HookUser  //issue creator
	Assignee    HookUser  //issue principal
	Iid         string    //issue Logo
	Title       string    //issue title
	Description string    //issue description
	State       string    //issue status
	Url         string    //issue URL on code cloud
	Repository  HookRepository
}

func QueryEulerOriginIssue(eoi *EulerOriginIssue, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

// insert data
func InsertEulerOriginIssue(eoi *EulerOriginIssue) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateEulerOriginIssue(eoi *EulerOriginIssue, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryGaussOriginIssue(eoi *GaussOriginIssue, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eoi, field...)
	return err
}

func QueryEulerUserUnassigned(euu *EulerUnassignUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(euu, field...)
	return err
}

// insert data
func InsertEulerUserUnassigned(euu *EulerUnassignUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(euu)
	return id, err
}

func UpdateEulerUserUnassigned(euu *EulerUnassignUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(euu, fields...)
	return err
}

func DelEulerUserUnassigned(euu *EulerUnassignUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Delete(euu, fields...)
	return err
}

// insert data
func InsertGaussOriginIssue(eoi *GaussOriginIssue) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eoi)
	return id, err
}

func UpdateGaussOriginIssue(eoi *GaussOriginIssue, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eoi, fields...)
	return err
}

func QueryGaussUser(eu *GaussUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eu, field...)
	return err
}

func InsertGaussUser(eu *GaussUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eu)
	return id, err
}

func UpdateGaussUser(eu *GaussUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eu, fields...)
	return err
}

func InsertGaussIssueUserRecord(er *GaussIssueUserRecord) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(er)
	return id, err
}

func QueryGaussIssuePrUser(eiu *GaussIssuePrUser, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertGaussIssuePrUser(eiu *GaussIssuePrUser) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateGaussIssuePrUser(eiu *GaussIssuePrUser, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func QueryGaussUserIntegDetail(eiu *GaussUserIntegDetail, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertGaussUserIntegDetail(eiu *GaussUserIntegDetail) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateGaussUserIntegDetail(eiu *GaussUserIntegDetail, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func QueryGaussUserIntegCount(eiu *GaussUserIntegCount, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertGaussUserIntegCount(eiu *GaussUserIntegCount) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateGaussUserIntegCount(eiu *GaussUserIntegCount, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

func QueryGaussIssuePrComplate(eiu *GaussIssuePrComplate, field ...string) error {
	o := orm.NewOrm()
	err := o.Read(eiu, field...)
	return err
}

// insert data
func InsertGaussIssuePrComplate(eiu *GaussIssuePrComplate) (int64, error) {
	o := orm.NewOrm()
	id, err := o.Insert(eiu)
	return id, err
}

func UpdateGaussIssuePrComplate(eiu *GaussIssuePrComplate, fields ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(eiu, fields...)
	return err
}

