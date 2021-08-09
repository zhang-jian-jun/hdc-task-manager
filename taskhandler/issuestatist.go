package taskhandler

func EulerIssueStatistics() error {
	//fileName := ""
	//totalName := ""
	//dir := beego.AppConfig.String("path_file")
	//eulerToken := beego.AppConfig.String("repo::git_token")
	//taskStartTime := beego.AppConfig.String("task_start_time")
	//owner := beego.AppConfig.String("repo::owner")
	//curDate := common.GetCurDate()
	//ChangeToWeek(taskStartTime[:10], curDate)
	//// File storage directory
	//CreateDir(dir)
	//totalFileSlice := make([]string, 0)
	//fileSlice := make([]string, 0)
	//// weekly date
	//lastWeekFirst := common.GetLastWeekFirstDate()
	//curWeekFirst := common.GetFirstDateOfWeek()
	//// last month
	//startMonth, endMonth := common.GetLastMonthDate()
	//// Get user information
	//eulerUser := models.QueryOpenEulerUserAll()
	//if len(eulerUser) > 0 {
	//	if flag == 1 {
	//		totalExcelValue := make([]ExcelValue, 0)
	//		fileName = "HDC_openEuler_weekly_points"
	//		totalName = "HDC_openEuler_total_points"
	//		fileName = fileName + "_" + lastWeekFirst[:10] + "_" + curWeekFirst[:10] + ".xlsx"
	//		fileName = filepath.Join(dir, fileName)
	//		totalName = totalName + ".xlsx"
	//		totalName = filepath.Join(dir, totalName)
	//		zipFileName := "HDC_openEuler_weekly_points" + ".zip"
	//		zipFileName = filepath.Join(dir, zipFileName)
	//		fileExcelPath := createExcel(fileName)
	//		if fileExcelPath == "" {
	//			logs.Error("Failed to create file")
	//			return
	//		}
	//		totalExcelPath := createExcel(totalName)
	//		if fileExcelPath == "" {
	//			logs.Error("Failed to create file")
	//			return
	//		}
	//		weekExcelValue := make([]ExcelValue, 0)
	//		for i, eu := range eulerUser {
	//			userFlag := false
	//			for _, userValue := range noProcUserSlice {
	//				if userValue == eu.GitUserId {
	//					userFlag = true
	//					break
	//				}
	//			}
	//			if userFlag {
	//				continue
	//			}
	//			logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
	//			weekExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId, eu.EmailAddr,
	//				lastWeekFirst, curWeekFirst, eu.UserId, weekExcelValue, i)
	//			totalExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId, eu.EmailAddr,
	//				"", "", eu.UserId, totalExcelValue, i)
	//		}
	//		ExcelData(fileExcelPath, weekExcelValue)
	//		ExcelData(totalExcelPath, totalExcelValue)
	//		totalFileSlice = append(totalFileSlice, fileName)
	//		totalFileSlice = append(totalFileSlice, totalName)
	//		zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
	//		if zipErr != nil {
	//			logs.Error("File compression failed: err: ", zipErr)
	//		}
	//		SendEulerExcel(zipFileName, lastWeekFirst, curWeekFirst, 1)
	//		fileSlice = append(fileSlice, fileName)
	//		fileSlice = append(fileSlice, totalName)
	//		fileSlice = append(fileSlice, zipFileName)
	//		DelFile(fileSlice)
	//	} else {
	//		totalExcelValue := make([]ExcelValue, 0)
	//		fileName = "HDC_openEuler_monthly_points"
	//		totalName = "HDC_openEuler_total_points"
	//		fileName = fileName + "_" + startMonth[:10] + "_" + endMonth[:10] + ".xlsx"
	//		fileName = filepath.Join(dir, fileName)
	//		totalName = totalName + ".xlsx"
	//		totalName = filepath.Join(dir, totalName)
	//		zipFileName := "HDC_openEuler_monthly_points" + ".zip"
	//		zipFileName = filepath.Join(dir, zipFileName)
	//		fileExcelPath := createExcel(fileName)
	//		if fileExcelPath == "" {
	//			logs.Error("Failed to create file")
	//			return
	//		}
	//		totalExcelPath := createExcel(totalName)
	//		if fileExcelPath == "" {
	//			logs.Error("Failed to create file")
	//			return
	//		}
	//		monthExcelValue := make([]ExcelValue, 0)
	//		for i, eu := range eulerUser {
	//			userFlag := false
	//			for _, userValue := range noProcUserSlice {
	//				if userValue == eu.GitUserId {
	//					userFlag = true
	//					break
	//				}
	//			}
	//			if userFlag {
	//				continue
	//			}
	//			logs.Info(fmt.Sprintf("Calculate the integral value of the first: %d user: %s", i, eu.GitUserId))
	//			monthExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId,
	//				eu.EmailAddr, startMonth, endMonth, eu.UserId, monthExcelValue, i)
	//			totalExcelValue = CalculateEulerPoint(eulerToken, taskStartTime, owner, eu.GitUserId,
	//				eu.EmailAddr, "", "", eu.UserId, totalExcelValue, i)
	//		}
	//		ExcelData(fileExcelPath, monthExcelValue)
	//		ExcelData(totalExcelPath, totalExcelValue)
	//		totalFileSlice = append(totalFileSlice, fileName)
	//		totalFileSlice = append(totalFileSlice, totalName)
	//		zipErr := ZipFiles(zipFileName, totalFileSlice, dir, dir)
	//		if zipErr != nil {
	//			logs.Error("File compression failed: err: ", zipErr)
	//		}
	//		SendEulerExcel(zipFileName, startMonth, endMonth, 2)
	//		fileSlice = append(fileSlice, fileName)
	//		fileSlice = append(fileSlice, totalName)
	//		fileSlice = append(fileSlice, zipFileName)
	//		DelFile(fileSlice)
	//	}
	//}
	return nil
}
