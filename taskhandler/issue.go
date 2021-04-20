package taskhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var issueLock sync.Mutex

//OrgInfo
type OrgInfo struct {
	ID           int32  `json:"id,omitempty"`
	Login        string `json:"login,omitempty"`
	URL          string `json:"url,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	ReposURL     string `json:"repos_url,omitempty"`
	EventsURL    string `json:"events_url,omitempty"`
	MembersURL   string `json:"members_url,omitempty"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	Enterprise   string `json:"enterprise,omitempty"`
	Members      int64  `json:"members,omitempty"`
	PublicRepos  int64  `json:"public_repos,omitempty"`
	PrivateRepos int64  `json:"private_repos,omitempty"`
}

//Branch Get all branches
type Branch struct {
	Name          string `json:"name,omitempty"`
	Protected     bool   `json:"protected,omitempty"`
	ProtectionURL string `json:"protection_url,omitempty"`
}

//PackageInfo package info model
type PackageInfo struct {
	Code string
	Msg  string
	Data Info
}

//Info cve info
type Info struct {
	Description string
}
type GaussIssueUserRecordTp struct {
	UserId      int64
	OrId        int64
	IssueNumber string
	RepoPath    string
	Owner       string
	Status      int8
	Type        int8
}

func GetOriginIssue(owner, eulerToken, taskLabels string) error {
	logs.Info("Synchronize gitee's issue start......")
	orgInfo, err := GetOrgInfo(eulerToken, owner)
	if err != nil {
		logs.Error("GetOrgInfo, owner: ", owner, ",err: ", err)
		return err
	}
	reposNum := orgInfo.PublicRepos + orgInfo.PrivateRepos
	if reposNum <= 0 {
		logs.Info(fmt.Sprintf("%v contain %v repository,grab issue finish!", owner, reposNum))
		return errors.New(fmt.Sprintf("%v contain %v repository,grab issue finish!", owner, reposNum))
	}
	pageSize := reposNum / int64(perPage)
	if reposNum%int64(perPage) > 0 {
		pageSize = pageSize + 1
	}
	var i int64
	for i = 1; i <= pageSize; i++ {
		go GetOrgRepos(eulerToken, owner, taskLabels, i)
	}
	wg.Wait()
	logs.Info("Synchronize gitee's issue  finish...")
	return nil
}

//GrabIssueByRepo grab issue by repository
func GrabIssueByRepo(accToken, owner, repo, state, taskLabels string) {
	page := 1
	for {
		list, err := GetIssueList(accToken, owner, repo, state, taskLabels, page)
		if err != nil {
			logs.Error("GetIssueList, repo: ", repo, ",err: ", err)
			break
		}
		issueLock.Lock()
		handleIssueList(list)
		issueLock.Unlock()
		if len(list) < perPage {
			break
		}
		page++

	}
}

func handleIssueList(list []models.HookIssue) {
	if len(list) == 0 {
		return
	}
	//var gil []models.GiteOriginIssue
	for _, v := range list {
		issueTitle := common.TrimString(v.Title)
		issueType := common.TrimString(v.IssueType)
		issueNumber := common.TrimString(v.Number)
		repoPath := common.TrimString(v.Repository.Path)
		owner := common.TrimString(v.Repository.NameSpace.Path)
		if issueType == CIssueType || issueTitle == CIssueType {
			eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath, IssueId: v.Id, IssueNumber: issueNumber}
			eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
			if eiErr != nil {
				CreateIssueOrgData(v, &eoi, 1)
				eId, orErr := models.InsertEulerOriginIssue(&eoi)
				if orErr != nil {
					logs.Error("InsertEulerOriginIssue, id: ", eId, ",err: ", orErr)
					continue
				}
			} else {
				updateStr := CreateIssueOrgData(v, &eoi, 2)
				upErr := models.UpdateEulerOriginIssue(&eoi, updateStr...)
				if upErr != nil {
					logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
					continue
				}
			}
		}
	}
}

func CreateIssueOrgData(hi models.HookIssue, eoi *models.EulerOriginIssue, flag int) []string {
	updateStr := make([]string, 0)
	issueState := common.TrimString(hi.State)
	issueZhState := common.TrimString(hi.IssueState)
	eoi.IssueState = issueState
	updateStr = append(updateStr, "IssueState")
	eoi.IssueStateName = issueZhState
	updateStr = append(updateStr, "IssueStateName")
	eoi.GitUrl = hi.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	eoi.IssueCreate = hi.User.Login
	updateStr = append(updateStr, "IssueCreate")
	eoi.RepoUrl = hi.Repository.Url
	//updateStr = append(updateStr, "RepoUrl")
	eoi.IssueNumber = common.TrimString(hi.Number)
	eoi.IssueId = hi.Id
	eoi.RepoPath = hi.Repository.Path
	eoi.Owner = hi.Repository.NameSpace.Path
	eoi.Status = 1
	updateStr = append(updateStr, "Status")
	if len(hi.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.CreateTime = hi.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.UpdateTime = hi.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.FinishedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		eoi.FinishedTime = hi.FinishedAt.String()
		updateStr = append(updateStr, "FinishedTime")
	}
	labelStr := ""
	if hi.Labels != nil && len(hi.Labels) > 0 {
		for _, la := range hi.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	eoi.IssueLabel = labelStr
	updateStr = append(updateStr, "IssueLabel")
	eoi.IssueType = common.TrimString(hi.IssueType)
	updateStr = append(updateStr, "IssueType")
	eoi.Title = hi.Title
	updateStr = append(updateStr, "Title")
	vb := strings.ReplaceAll(hi.Body, "：", "：")
	eoi.IssueBody = vb
	updateStr = append(updateStr, "IssueBody")
	el := RegexpEmergencyLevel.FindAllStringSubmatch(hi.Body, -1)
	if len(el) > 0 && len(el[0]) > 0 {
		eoi.EmergencyLevel = common.TrimString(el[0][1])
		updateStr = append(updateStr, "EmergencyLevel")
		value := RegexpDigit.FindAllStringSubmatch(eoi.EmergencyLevel, -1)
		eoi.EmergencyValue, _ = strconv.Atoi(value[0][1])
		updateStr = append(updateStr, "EmergencyValue")
	}
	dd := RegexpDegreeDiff.FindAllStringSubmatch(hi.Body, -1)
	if len(dd) > 0 && len(dd[0]) > 0 {
		eoi.DifficultLevel = common.TrimString(dd[0][1])
		updateStr = append(updateStr, "DifficultLevel")
		value := RegexpDigit.FindAllStringSubmatch(eoi.DifficultLevel, -1)
		eoi.DifficultValue, _ = strconv.Atoi(value[0][1])
		updateStr = append(updateStr, "DifficultValue")
	}
	if flag == 1 {
		eoi.GrabTime = common.GetCurTime()
	}
	eoi.IssueAssignee = hi.Assignee.Login
	updateStr = append(updateStr, "IssueAssignee")
	return updateStr
}

//GetOrgInfo get  organization information
func GetOrgInfo(accToken, org string) (OrgInfo, error) {
	oi := OrgInfo{}
	resp, err := http.Get(fmt.Sprintf(GiteOrgInfoURL, org, accToken))
	if err != nil {
		return oi, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return oi, err
	}
	err = json.Unmarshal(body, &oi)
	return oi, err
}

//GetOrgRepos get organization repository
func GetOrgRepos(accToken, org, taskLabels string, page int64) {
	wg.Add(1)
	defer wg.Done()
	resp, err := http.Get(fmt.Sprintf(GiteOrgReposURL, org, accToken, page, perPage))
	if err != nil {
		logs.Error("Get, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	defer resp.Body.Close()
	var reps []models.Repository
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("ReadAll, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	//logs.Info("GetOrgRepos, body: ", string(body))
	err = json.Unmarshal(body, &reps)
	if err != nil {
		logs.Error("Unmarshal, GiteOrgReposURL: ", GiteOrgReposURL, ", org: ", GiteOrgReposURL, ",err: ", err)
		return
	}
	for _, v := range reps {
		GrabIssueByRepo(accToken, org, v.Name, "all", taskLabels)
	}
}

//GetIssueList get the repository issue list
func GetIssueList(accToken, owner, repo, state, taskLabel string, page int) (issueList []models.HookIssue, err error) {
	giteUrl := fmt.Sprintf(GiteRepoIssuesURL, owner, repo, accToken, state, taskLabel, page, perPage)
	resp, err := http.Get(giteUrl)
	if err != nil {
		logs.Error("Get, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
		return issueList, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("ReadAll, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
		return issueList, err
	}
	logs.Info("-----------issue list: ", string(body))
	err = json.Unmarshal(body, &issueList)
	if err != nil {
		logs.Error("Unmarshal, GiteRepoIssuesURL: ", giteUrl, ", repo: ", repo, ", err: ", err)
	}
	//logs.Info("++++++++++issueList: ", issueList)
	return
}

func AddHookIssue(issueData *models.IssuePayload) {
	issueTitle := common.TrimString(issueData.Issue.Title)
	issueType := common.TrimString(issueData.Issue.TypeName)
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	if issueType == CIssueType || issueTitle == CIssueType {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eiErr != nil {
			CreateHookIssueOrgData(issueData, &eoi, 1)
			eId, orErr := models.InsertEulerOriginIssue(&eoi)
			if orErr != nil {
				logs.Error("InsertEulerOriginIssue, id: ", eId, ",err: ", orErr)
				return
			}
		} else {
			updateStr := CreateHookIssueOrgData(issueData, &eoi, 2)
			upErr := models.UpdateEulerOriginIssue(&eoi, updateStr...)
			if upErr != nil {
				logs.Error("UpdateEulerOriginIssue, upErr: ", upErr)
				return
			}
		}
	}
}

func CreateHookIssueOrgData(hi *models.IssuePayload, eoi *models.EulerOriginIssue, flag int) []string {
	updateStr := make([]string, 0)
	issueState := common.TrimString(hi.State)
	issueZhState := common.TrimString(hi.Issue.StateName)
	eoi.IssueState = issueState
	updateStr = append(updateStr, "IssueState")
	eoi.IssueStateName = issueZhState
	updateStr = append(updateStr, "IssueStateName")
	eoi.GitUrl = hi.Issue.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	eoi.IssueCreate = hi.User.Login
	updateStr = append(updateStr, "IssueCreate")
	eoi.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	eoi.IssueNumber = common.TrimString(hi.Issue.Number)
	eoi.IssueId = hi.Issue.Id
	eoi.RepoPath = hi.Repository.Path
	eoi.Owner = hi.Repository.NameSpace
	eoi.Status = 1
	updateStr = append(updateStr, "Status")
	if len(hi.Issue.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.CreateTime = hi.Issue.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.Issue.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.UpdateTime = hi.Issue.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.Issue.FinishedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		eoi.FinishedTime = hi.Issue.FinishedAt.String()
		updateStr = append(updateStr, "FinishedTime")
	}
	labelStr := ""
	if hi.Issue.Labels != nil && len(hi.Issue.Labels) > 0 {
		for _, la := range hi.Issue.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
	}
	eoi.IssueLabel = labelStr
	updateStr = append(updateStr, "IssueLabel")
	eoi.IssueType = common.TrimString(hi.Issue.TypeName)
	updateStr = append(updateStr, "IssueType")
	eoi.Title = hi.Issue.Title
	updateStr = append(updateStr, "Title")
	vb := strings.ReplaceAll(hi.Issue.Body, "：", "：")
	eoi.IssueBody = vb
	updateStr = append(updateStr, "IssueBody")
	el := RegexpEmergencyLevel.FindAllStringSubmatch(hi.Issue.Body, -1)
	if len(el) > 0 && len(el[0]) > 0 {
		eoi.EmergencyLevel = common.TrimString(el[0][1])
		updateStr = append(updateStr, "EmergencyLevel")
		value := RegexpDigit.FindAllStringSubmatch(eoi.EmergencyLevel, -1)
		eoi.EmergencyValue, _ = strconv.Atoi(value[0][1])
		updateStr = append(updateStr, "EmergencyValue")
	}
	dd := RegexpDegreeDiff.FindAllStringSubmatch(hi.Issue.Body, -1)
	if len(dd) > 0 && len(dd[0]) > 0 {
		eoi.DifficultLevel = common.TrimString(dd[0][1])
		updateStr = append(updateStr, "DifficultLevel")
		value := RegexpDigit.FindAllStringSubmatch(eoi.DifficultLevel, -1)
		eoi.DifficultValue, _ = strconv.Atoi(value[0][1])
		updateStr = append(updateStr, "DifficultValue")
	}
	if flag == 1 {
		eoi.GrabTime = common.GetCurTime()
	}
	eoi.IssueAssignee = hi.Assignee.Login
	updateStr = append(updateStr, "IssueAssignee")
	logs.Info("eoi===>", eoi)
	return updateStr
}

func DelHookIssue(issueData *models.IssuePayload) {
	issueTitle := common.TrimString(issueData.Issue.Title)
	issueType := common.TrimString(issueData.Issue.TypeName)
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	if issueType == CIssueType || issueTitle == CIssueType {
		eoi := models.EulerOriginIssue{Owner: owner, RepoPath: repoPath,
			IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
		eiErr := models.QueryEulerOriginIssue(&eoi, "Owner", "RepoPath", "IssueId", "IssueNumber")
		if eoi.OrId == 0 {
			logs.Error("DelHookIssue, Data does not exist, eiErr: ", eiErr)
			return
		} else {
			userId, delErr := models.DeleteEulerOriginIssueAll(&eoi)
			if delErr != nil {
				logs.Error("DeleteEulerOriginIssueAll, Data deletion failed, delErr: ", delErr)
				return
			}
			et := EulerIssueUserRecordTp{UserId: userId, OrId: eoi.OrId, IssueNumber: eoi.IssueNumber,
				RepoPath: eoi.RepoPath, Owner: owner, Status: 12}
			EulerIssueUserRecord(et)
		}
	}
}

func CreateHookGaussIssueOrgData(hi *models.IssuePayload, eoi *models.GaussOriginIssue, flag int) []string {
	updateStr := make([]string, 0)
	issueState := common.TrimString(hi.State)
	issueZhState := common.TrimString(hi.Issue.StateName)
	eoi.IssueState = issueState
	updateStr = append(updateStr, "IssueState")
	eoi.IssueStateName = issueZhState
	updateStr = append(updateStr, "IssueStateName")
	eoi.GitUrl = hi.Issue.HtmlUrl
	updateStr = append(updateStr, "GitUrl")
	eoi.IssueCreate = hi.User.Login
	updateStr = append(updateStr, "IssueCreate")
	eoi.RepoUrl = hi.Repository.Url
	updateStr = append(updateStr, "RepoUrl")
	eoi.IssueNumber = common.TrimString(hi.Issue.Number)
	eoi.IssueId = hi.Issue.Id
	eoi.RepoPath = hi.Repository.Path
	eoi.Owner = hi.Repository.NameSpace
	eoi.Status = 1
	updateStr = append(updateStr, "Status")
	if len(hi.Issue.CreateAt.String()) > 1 {
		//eoi.CreateTime = common.TimeToLocal(hi.CreateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.CreateTime = hi.Issue.CreateAt.String()
		updateStr = append(updateStr, "CreateTime")
	}
	if len(hi.Issue.UpdateAt.String()) > 1 {
		//eoi.UpdateTime = common.TimeToLocal(hi.UpdateAt.String()[:19], "2006-01-02T15:04:05")
		eoi.UpdateTime = hi.Issue.UpdateAt.String()
		updateStr = append(updateStr, "UpdateTime")
	}
	if len(hi.Issue.FinishedAt.String()) > 1 {
		//eoi.FinishedTime = common.TimeToLocal(hi.FinishedAt.String()[:19], "2006-01-02T15:04:05")
		eoi.FinishedTime = hi.Issue.FinishedAt.String()
		updateStr = append(updateStr, "FinishedTime")
	}
	labelStr := ""
	if hi.Issue.Labels != nil && len(hi.Issue.Labels) > 0 {
		for _, la := range hi.Issue.Labels {
			labelStr = labelStr + la.Name + ","
		}
		labelStr = labelStr[:len(labelStr)-1]
		eoi.IssueLabel = labelStr
		updateStr = append(updateStr, "IssueLabel")
	}
	eoi.IssueType = common.TrimString(hi.Issue.TypeName)
	updateStr = append(updateStr, "IssueType")
	eoi.Title = hi.Issue.Title
	updateStr = append(updateStr, "Title")
	vb := strings.ReplaceAll(hi.Issue.Body, "：", "：")
	eoi.IssueBody = vb
	updateStr = append(updateStr, "IssueBody")
	if flag == 1 {
		eoi.GrabTime = common.GetCurTime()
	}
	eoi.IssueAssignee = hi.Assignee.Login
	updateStr = append(updateStr, "IssueAssignee")
	logs.Info("eoi===>", eoi)
	return updateStr
}

// AddHookGaussIssue Increase issue
func AddHookGaussIssue(issueData *models.IssuePayload) {
	gaussToken := os.Getenv("GITEE_GAUSS_TOKEN")
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	goi := models.GaussOriginIssue{Owner: owner, RepoPath: repoPath,
		IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
	eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
	if eiErr != nil {
		CreateHookGaussIssueOrgData(issueData, &goi, 1)
		eId, orErr := models.InsertGaussOriginIssue(&goi)
		if orErr != nil {
			logs.Error("InsertGaussOriginIssue, id: ", eId, ",err: ", orErr)
			return
		}
		goi.OrId = eId
	} else {
		updateStr := CreateHookGaussIssueOrgData(issueData, &goi, 2)
		upErr := models.UpdateGaussOriginIssue(&goi, updateStr...)
		if upErr != nil {
			logs.Error("UpdateGaussOriginIssue, upErr: ", upErr)
			return
		}
	}
	// Create user information
	userId := StoreGitGaussUser(issueData.User.Login, issueData.User.Email)
	if userId > 0 {
		et := GaussIssueUserRecordTp{UserId: userId, OrId: goi.OrId, IssueNumber: issueData.Issue.Number,
			RepoPath: issueData.Repository.Path, Owner: owner, Status: 1, Type: 1}
		GaussIssueUserRecord(et)
		// Create the correspondence between users and issues, as well as user points information
		GaussIssueUser(userId, goi.OrId, goi.IssueNumber, goi.RepoPath, goi.Owner, 1)
		// Calculate the points earned by users
		CreateUserPoints(userId, goi.OrId, 0, 1)
		hdcGaussLabel := beego.AppConfig.String("hdc_gauss_label")
		if len(goi.IssueLabel) > 1 && strings.Contains(strings.ToLower(goi.IssueLabel), hdcGaussLabel) {
			// Will write issue comments
			igc := fmt.Sprintf(IssueGaussComment, issueData.User.Login)
			AddCommentToIssue(igc, issueData.Issue.Number, owner, issueData.Repository.Path, gaussToken)
			// edit label
			//hdcGuassLabel := beego.AppConfig.String("hdc_gauss_label")
			//EditGaussLabel(issueData.Issue.Number, hdcGuassLabel, gaussToken, owner, goi)
			// Send private message
			igcs := fmt.Sprintf(IssueGaussCommentSend, goi.GitUrl)
			SendPrivateLetters(gaussToken, igcs, issueData.User.Login)
			assigneeStr := beego.AppConfig.String("gauss::assignee")
			if len(assigneeStr) > 1 {
				assigneeSlice := strings.Split(assigneeStr, ",")
				if len(assigneeSlice) > 0 {
					for _, as := range assigneeSlice {
						igcs := fmt.Sprintf(IssueGaussRewiewSend, issueData.User.Login, goi.GitUrl)
						SendPrivateLetters(gaussToken, igcs, as)
					}
				}
			}
		}
	}
}

// EditGaussLabel Edit label
func EditGaussLabel(number, hdcGuassLabel, gaussToken, owner string, goi models.GaussOriginIssue) {
	labels := goi.IssueLabel
	lalelSlice := []string{beego.AppConfig.String("hdc_gauss_label")}
	lalelSlice = append(lalelSlice, hdcGuassLabel)
	for _, lab := range lalelSlice {
		if len(labels) > 1 {
			labSlice := strings.Split(labels, ",")
			tmpSlice := []string{}
			for _, lb := range labSlice {
				if !strings.HasPrefix(lb, "challenge-") {
					tmpSlice = append(tmpSlice, lb)
				}
			}
			if len(tmpSlice) > 0 {
				labels = strings.Join(tmpSlice, ",")
			}
			if !strings.Contains(labels, lab) {
				labels = labels + "," + lab
			}
		} else {
			labels = lab
		}
	}
	ChangeIssueLabel(gaussToken, goi.RepoPath, number, owner, labels)
	goi.IssueLabel = labels
	goi.UpdateTime = common.GetCurTime()
	upErr := models.UpdateGaussOriginIssue(&goi, "IssueLabel", "UpdateTime")
	if upErr != nil {
		logs.Error("UpdateGaussOriginIssue, upErr: ", upErr)
	}
}

// EditGaussPrLabel Edit label
func EditGaussPrLabel(hdcGuassLabel, gaussToken, owner string, gop models.GaussOriginPr, number int64) {
	labels := gop.PrLabel
	lalelSlice := []string{beego.AppConfig.String("hdc_gauss_label")}
	lalelSlice = append(lalelSlice, hdcGuassLabel)
	for _, lab := range lalelSlice {
		if len(labels) > 1 {
			labSlice := strings.Split(labels, ",")
			tmpSlice := []string{}
			for _, lb := range labSlice {
				if !strings.HasPrefix(lb, "challenge-") {
					tmpSlice = append(tmpSlice, lb)
				}
			}
			if len(tmpSlice) > 0 {
				labels = strings.Join(tmpSlice, ",")
			}
			if !strings.Contains(labels, lab) {
				labels = labels + "," + lab
			}
		} else {
			labels = lab
		}
	}
	ChangePrLabel(gaussToken, gop.RepoPath, owner, labels, number)
	gop.PrLabel = labels
	gop.UpdateTime = common.GetCurTime()
	upErr := models.UpdateGaussOriginPr(&gop, "PrLabel", "UpdateTime")
	if upErr != nil {
		logs.Error("UpdateGaussOriginPr, upErr: ", upErr)
	}
}

// Calculate the points earned by users
func CreateUserPoints(userId, orId, points int64, dataType int8) {
	// Query points information
	gid := models.GaussUserIntegDetail{UserId: userId, OrId: orId, Type: dataType}
	gidErr := models.QueryGaussUserIntegDetail(&gid, "UserId", "OrId", "Type")
	if gid.Id > 0 {
		gid.IntegralValue += points
		gidErr = models.UpdateGaussUserIntegDetail(&gid, "IntegralValue")
		if gidErr != nil {
			logs.Error("UpdateGaussUserIntegDetail, gidErr: ", gidErr)
			return
		}
	} else {
		gid = models.GaussUserIntegDetail{UserId: userId, OrId: orId, Type: dataType,
			IntegralValue: points, CreateTime: common.GetCurTime()}
		id, indErr := models.InsertGaussUserIntegDetail(&gid)
		if id == 0 {
			logs.Error("InsertGaussUserIntegDetail, indErr:", indErr, id)
			return
		}
	}
	gic := models.GaussUserIntegCount{UserId: userId}
	gicErr := models.QueryGaussUserIntegCount(&gic, "UserId")
	if gic.Id == 0 {
		gic = models.GaussUserIntegCount{UserId: userId, IntegralValue: points, CreateTime: common.GetCurTime()}
		gicId, ingicErr := models.InsertGaussUserIntegCount(&gic)
		if gicId == 0 {
			logs.Error("InsertGaussUserIntegCount, ingicErr: ", ingicErr, gicId)
		}
	} else {
		gic.IntegralValue += points
		gicErr = models.UpdateGaussUserIntegCount(&gic, "IntegralValue")
		if gicErr != nil {
			logs.Error("UpdateEulerUserIntegCount, gicErr: ", gicErr)
		}
	}
}

// Create the correspondence between users and issues, as well as user points information
func GaussIssueUser(userId, orId int64, num, repo, owner string, dataType int8) {
	gip := models.GaussIssuePrUser{OrId: orId, UserId: userId, Type: dataType}
	euErr := models.QueryGaussIssuePrUser(&gip, "OrId", "UserId", "Type")
	if gip.Id == 0 {
		gip = models.GaussIssuePrUser{OrId: orId, UserId: userId, Number: num,
			RepoPath: repo, Owner: owner, SendEmail: 1, Status: 1,
			CreateTime: common.GetCurTime(), Type: dataType}
		id, inErr := models.InsertGaussIssuePrUser(&gip)
		if id == 0 {
			logs.Error("InsertGaussIssuePrUser, euErr: ", euErr, ",inErr: ", inErr)
		}
	} else {
		gip.Status = 1
		gip.UpdateTime = common.GetCurTime()
		upErr := models.UpdateGaussIssuePrUser(&gip)
		if upErr != nil {
			logs.Error("UpdateGaussIssuePrUser, upErr: ", upErr)
		}
	}
}

// StoreGitGaussUser store user
func StoreGitGaussUser(gieeId, email string) int64 {
	gu := models.GaussUser{GitUserId: gieeId}
	eiErr := models.QueryGaussUser(&gu, "GitUserId")
	if eiErr != nil {
		// insert data
		gu.CreateTime = common.GetCurTime()
		gu.Status = 1
		gu.EmailAddr = email
		gu.GitUserId = gieeId
		userId, inErr := models.InsertGaussUser(&gu)
		if inErr != nil {
			logs.Error("InsertGaussUser, inerr: ", inErr)
		}
		return userId
	} else {
		// update data
		gu.Status = 1
		gu.UpdateTime = common.GetCurTime()
		upErr := models.UpdateGaussUser(&gu, "Status", "UpdateTime")
		if upErr != nil {
			logs.Error("UpdateGaussUser, inerr: ", upErr)
		}
		return gu.UserId
	}
}

func GaussIssueUserRecord(et GaussIssueUserRecordTp) {
	eir := models.GaussIssueUserRecord{UserId: et.UserId, OrId: et.OrId, Number: et.IssueNumber,
		RepoPath: et.RepoPath, Owner: et.Owner, Status: et.Status,
		CreateTime: common.GetCurTime(), Type: et.Type}
	models.InsertGaussIssueUserRecord(&eir)
}

// DelHookGaussIssue delete issue
func DelHookGaussIssue(issueData *models.IssuePayload) {
	issueNumber := common.TrimString(issueData.Issue.Number)
	repoPath := common.TrimString(issueData.Repository.Path)
	owner := common.TrimString(issueData.Repository.NameSpace)
	goi := models.GaussOriginIssue{Owner: owner, RepoPath: repoPath,
		IssueId: issueData.Issue.Id, IssueNumber: issueNumber}
	eiErr := models.QueryGaussOriginIssue(&goi, "Owner", "RepoPath", "IssueId", "IssueNumber")
	if goi.OrId == 0 {
		logs.Error("DelHookGaussIssue, Data does not exist, eiErr: ", eiErr)
		return
	} else {
		userId, delErr := models.DeleteGaussOriginIssueAll(&goi, 1)
		if delErr != nil {
			logs.Error("DeleteGaussOriginIssueAll, Data deletion failed, delErr: ", delErr)
			return
		}
		et := GaussIssueUserRecordTp{UserId: userId, OrId: goi.OrId, IssueNumber: goi.IssueNumber,
			RepoPath: goi.RepoPath, Owner: owner, Status: 3}
		GaussIssueUserRecord(et)
	}
}
