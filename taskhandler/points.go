package taskhandler

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"hdc-task-manager/common"
	"hdc-task-manager/models"
	"path/filepath"
	"strings"
	"time"
)

type StatistDate struct {
	LastWeekFirst string
	CurWeekFirst  string
	StartMonth    string
	EndMonth      string
}

// Start the integration task
func GetWeekPointsTask(exportWPoint string) {
	logs.Info("Execute the task of exporting weekly points and start...")
	pointTask := toolbox.NewTask("DealWeekPointStart", exportWPoint, DealWeekPointStart)
	toolbox.AddTask("DealWeekPointStart", pointTask)
	logs.Info("The task of exporting weekly points is over...")
}

func GetSpecWeekPointsTask(specexportwpoint string) {
	logs.Info("Execute the task of exporting sepecly weekly points and start...")
	weeklyPointTask := toolbox.NewTask("DealSpecWeekPointStart", specexportwpoint, DealSpecWeekPointStart)
	toolbox.AddTask("DealSpecWeekPointStart", weeklyPointTask)
	logs.Info("The task of exporting sepecly weekly points is over...")
}

func GetMonthPointsTask(exportMPoint string) {
	logs.Info("Execute the task of exporting month points and start...")
	pointTask := toolbox.NewTask("DealMonthPointStart", exportMPoint, DealMonthPointStart)
	toolbox.AddTask("DealMonthPointStart", pointTask)
	logs.Info("The task of exporting month points is over...")
}

// Deal with integration tasks
func DealWeekPointStart() error {
	// 1. Calculate openEuler points
	CalculateOpenEulerPoint(1)
	time.Sleep(time.Second * 120)
	//2. Calculate openGauss points
	CalculateOpenGaussPoint(1)
	return nil
}

func DealSpecWeekPointStart() error {
	// 1. Calculate openEuler points
	CalculateEulerSpecWeekPoint(1)
	return nil
}

// Deal with integration tasks
func DealMonthPointStart() error {
	// 1. Calculate openEuler points
	CalculateOpenEulerPoint(2)
	time.Sleep(time.Second * 120)
	//2. Calculate openGauss points
	CalculateOpenGaussPoint(2)
	return nil
}

func CalculateEulerSpecWeekPoint(flag int) {
	// weekly date
	weekList := ChangeToWeek("2021-07-01", "2021-08-01")
	if len(weekList) > 0 {
		for _, wl := range weekList {
			statistDate := StatistDate{LastWeekFirst: wl[0], CurWeekFirst: wl[1], StartMonth: "", EndMonth: ""}
			ExportEulerPoints(flag, statistDate)
		}
	}
}

func CalculateOpenEulerPoint(flag int) {
	// weekly date
	lastWeekFirst := common.GetLastWeekFirstDate()
	curWeekFirst := common.GetFirstDateOfWeek()
	// last month
	startMonth, endMonth := common.GetLastMonthDate()
	statistDate := StatistDate{LastWeekFirst: lastWeekFirst, CurWeekFirst: curWeekFirst,
		StartMonth: startMonth, EndMonth: endMonth}
	ExportEulerPoints(flag, statistDate)
}

func ExportEulerPoints(flag int, statistDate StatistDate) {
	fileName := ""
	totalName := ""
	dir := beego.AppConfig.String("path_file")
	eulerToken := beego.AppConfig.String("repo::git_token")
	taskStartTime := beego.AppConfig.String("task_start_time")
	owner := beego.AppConfig.String("repo::owner")
	noProcUser := beego.AppConfig.String("repo::user_login")
	noProcUserSlice := strings.Split(noProcUser, ",")
	// File storage directory
	CreateDir(dir)
	totalFileSlice := make([]string, 0)
	fileSlice := make([]string, 0)
	// weekly date
	lastWeekFirst := statistDate.LastWeekFirst
	curWeekFirst := statistDate.CurWeekFirst
	// last month
	startMonth := statistDate.StartMonth
	endMonth := statistDate.EndMonth
	// Get user information
	eulerUser := models.QueryOpenEulerUserAll()
	if len(eulerUser) > 0 {
		if flag == 1 {
			totalExcelValue := make([]ExcelValue, 0)
			fileName = "HDC_openEuler_weekly_points"
			totalName = "HDC_openEuler_total_points"
			fileName = fileName + "_" + lastWeekFirst[:10] + "_" + curWeekFirst[:10] + ".xlsx"
			fileName = filepath.Join(dir, fileName)
			totalName = totalName + ".xlsx"
			totalName = filepath.Join(dir, totalName)
			zipFileName := "HDC_openEuler_weekly_points" + ".zip"
			zipFileName = filepath.Join(dir, zipFileName)
			fileExcelPath := createExcel(fileName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			totalExcelPath := createExcel(totalName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			weekExcelValue := make([]ExcelValue, 0)
			for i, eu := range eulerUser {
				userFlag := false
				for _, userValue := range noProcUserSlice {
					if userValue == eu.GitUserId {
						userFlag = true
						break
					}
				}
				if userFlag {
					continue
				}
				logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
				weekExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId, eu.EmailAddr,
					lastWeekFirst, curWeekFirst, eu.UserId, weekExcelValue, i)
				totalExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId, eu.EmailAddr,
					"", "", eu.UserId, totalExcelValue, i)
			}
			ExcelData(fileExcelPath, weekExcelValue)
			ExcelData(totalExcelPath, totalExcelValue)
			totalFileSlice = append(totalFileSlice, fileName)
			totalFileSlice = append(totalFileSlice, totalName)
			zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
			if zipErr != nil {
				logs.Error("File compression failed: err: ", zipErr)
			}
			SendEulerExcel(zipFileName, lastWeekFirst, curWeekFirst, 1)
			fileSlice = append(fileSlice, fileName)
			fileSlice = append(fileSlice, totalName)
			fileSlice = append(fileSlice, zipFileName)
			DelFile(fileSlice)
		} else {
			totalExcelValue := make([]ExcelValue, 0)
			fileName = "HDC_openEuler_monthly_points"
			totalName = "HDC_openEuler_total_points"
			fileName = fileName + "_" + startMonth[:10] + "_" + endMonth[:10] + ".xlsx"
			fileName = filepath.Join(dir, fileName)
			totalName = totalName + ".xlsx"
			totalName = filepath.Join(dir, totalName)
			zipFileName := "HDC_openEuler_monthly_points" + ".zip"
			zipFileName = filepath.Join(dir, zipFileName)
			fileExcelPath := createExcel(fileName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			totalExcelPath := createExcel(totalName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			monthExcelValue := make([]ExcelValue, 0)
			for i, eu := range eulerUser {
				userFlag := false
				for _, userValue := range noProcUserSlice {
					if userValue == eu.GitUserId {
						userFlag = true
						break
					}
				}
				if userFlag {
					continue
				}
				logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
				monthExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId,
					eu.EmailAddr, startMonth, endMonth, eu.UserId, monthExcelValue, i)
				totalExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId,
					eu.EmailAddr, "", "", eu.UserId, totalExcelValue, i)
			}
			ExcelData(fileExcelPath, monthExcelValue)
			ExcelData(totalExcelPath, totalExcelValue)
			totalFileSlice = append(totalFileSlice, fileName)
			totalFileSlice = append(totalFileSlice, totalName)
			zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
			if zipErr != nil {
				logs.Error("File compression failed: err: ", zipErr)
			}
			SendEulerExcel(zipFileName, startMonth, endMonth, 2)
			fileSlice = append(fileSlice, fileName)
			fileSlice = append(fileSlice, totalName)
			fileSlice = append(fileSlice, zipFileName)
			DelFile(fileSlice)
		}
	}
}

func CalculateOpenGaussPoint(flag int) {
	fileName := ""
	totalName := ""
	dir := beego.AppConfig.String("path_file")
	// File storage directory
	CreateDir(dir)
	totalFileSlice := make([]string, 0)
	fileSlice := make([]string, 0)
	// weekly date
	lastWeekFirst := common.GetLastWeekFirstDate()
	curWeekFirst := common.GetFirstDateOfWeek()
	// last month
	startMonth, endMonth := common.GetLastMonthDate()
	// Get user information
	gaussUser := models.QueryOpenGaussUserAll()
	if len(gaussUser) > 0 {
		if flag == 1 {
			totalExcelValue := make([]ExcelValue, 0)
			fileName = "HDC_openGauss_weekly_points"
			totalName = "HDC_openGauss_total_points"
			fileName = fileName + "_" + lastWeekFirst[:10] + "_" + curWeekFirst[:10] + ".xlsx"
			fileName = filepath.Join(dir, fileName)
			totalName = totalName + ".xlsx"
			totalName = filepath.Join(dir, totalName)
			zipFileName := "HDC_openGauss_weekly_points" + ".zip"
			zipFileName = filepath.Join(dir, zipFileName)
			fileExcelPath := createExcel(fileName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			totalExcelPath := createExcel(totalName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			weekExcelValue := make([]ExcelValue, 0)
			for i, eu := range gaussUser {
				logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
				weekExcelValue = CalculateGaussPoint(eu.GitUserId, eu.EmailAddr, lastWeekFirst, curWeekFirst, eu.UserId, weekExcelValue, i)
				totalExcelValue = CalculateGaussPoint(eu.GitUserId, eu.EmailAddr, "", "", eu.UserId, totalExcelValue, i)
			}
			ExcelData(fileExcelPath, weekExcelValue)
			ExcelData(totalExcelPath, totalExcelValue)
			totalFileSlice = append(totalFileSlice, fileName)
			totalFileSlice = append(totalFileSlice, totalName)
			zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
			if zipErr != nil {
				logs.Error("File compression failed: err: ", zipErr)
			}
			SendGaussExcel(zipFileName, lastWeekFirst, curWeekFirst, 1)
			fileSlice = append(fileSlice, fileName)
			fileSlice = append(fileSlice, totalName)
			fileSlice = append(fileSlice, zipFileName)
			DelFile(fileSlice)
		} else {
			totalExcelValue := make([]ExcelValue, 0)
			fileName = "HDC_openGauss_monthly_points"
			totalName = "HDC_openGauss_total_points"
			fileName = fileName + "_" + startMonth[:10] + "_" + endMonth[:10] + ".xlsx"
			fileName = filepath.Join(dir, fileName)
			totalName = totalName + ".xlsx"
			totalName = filepath.Join(dir, totalName)
			zipFileName := "HDC_openGauss_monthly_points" + ".zip"
			zipFileName = filepath.Join(dir, zipFileName)
			fileExcelPath := createExcel(fileName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			totalExcelPath := createExcel(totalName)
			if fileExcelPath == "" {
				logs.Error("Failed to create file")
				return
			}
			monthExcelValue := make([]ExcelValue, 0)
			for i, eu := range gaussUser {
				logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
				monthExcelValue = CalculateGaussPoint(eu.GitUserId, eu.EmailAddr, startMonth, endMonth, eu.UserId, monthExcelValue, i)
				totalExcelValue = CalculateGaussPoint(eu.GitUserId, eu.EmailAddr, "", "", eu.UserId, totalExcelValue, i)
			}
			ExcelData(fileExcelPath, monthExcelValue)
			ExcelData(totalExcelPath, totalExcelValue)
			totalFileSlice = append(totalFileSlice, fileName)
			totalFileSlice = append(totalFileSlice, totalName)
			zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
			if zipErr != nil {
				logs.Error("File compression failed: err: ", zipErr)
			}
			SendGaussExcel(zipFileName, startMonth, endMonth, 2)
			fileSlice = append(fileSlice, fileName)
			fileSlice = append(fileSlice, totalName)
			fileSlice = append(fileSlice, zipFileName)
			DelFile(fileSlice)
		}
	}
}

// Calculate points data on a weekly basis
func CalculateEulerPoint(eulerToken, taskStartTime, owner,
	gitLogin, gitEmail, startTime, endTime string,
	userId int64, evch []ExcelValue, index int) []ExcelValue {
	var evu ExcelValue
	var pv models.PointValue
	evu.GitLogin = gitLogin
	evu.GitEmail = gitEmail
	evu.CurIndex = index + 1
	models.QueryEulerUserIntegDetailValue(&pv, startTime, endTime, userId)
	if pv.Integration > 0 {
		evu.Points = pv.Integration
	}
	var plv models.PointValue
	models.QueryEulerUserIntegDetailCount(&plv, startTime, endTime, userId)
	if plv.Integration > 0 {
		evu.IssueCount = plv.Integration
	}
	sTime := StaticIssueTime{WeekIssueStartTime: startTime, WeekIssueEndTime: endTime,
		MonthIssueStartTime: startTime, MonthIssueEndTime: endTime, TotalIssueTime: taskStartTime}
	staticCount := GetUserPublicUpEvents(gitLogin, eulerToken, owner, 0, 20, sTime)
	if startTime == "" {
		evu.CommitIssueCount = staticCount.TotalIssueCount
		evu.CommentIssueCount = staticCount.TotalIssueCommentCount
		evu.CommitPrCount = staticCount.TotalPullRequestCount
	} else {
		evu.CommitIssueCount = staticCount.monthIssueCount
		evu.CommentIssueCount = staticCount.monthIssueCommentCount
		evu.CommitPrCount = staticCount.monthPullRequestCount
	}
	logs.Info("====================>evu: ", evu)
	evch = append(evch, evu)
	return evch
}

// Calculate points data on a weekly basis
func CalculateGaussPoint(gitLogin, gitEmail, lastWeekFirst, curWeekFirst string,
	userId int64, evch []ExcelValue, index int) []ExcelValue {
	var evu ExcelValue
	var pv models.PointValue
	evu.GitLogin = gitLogin
	evu.GitEmail = gitEmail
	evu.CurIndex = index + 1
	models.QueryGaussUserIntegDetailValue(&pv, lastWeekFirst, curWeekFirst, userId)
	if pv.Integration > 0 {
		evu.Points = pv.Integration
	}
	var plv models.PointValue
	models.QueryGaussUserIntegDetailCount(&plv, lastWeekFirst, curWeekFirst, userId)
	if plv.Integration > 0 {
		evu.IssueCount = plv.Integration
	}
	evch = append(evch, evu)
	return evch
}
