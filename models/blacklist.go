package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func QueryEulerBlackUserAll(status int8) (ebu []EulerBlackUser) {
	o := orm.NewOrm()
	if status > 0 {
		num, _ := o.Raw("select * from hdc_euler_black_user where status = ?", status).QueryRows(&ebu)
		if num > 0 {
			logs.Info("QueryEulerBlackUserAll, num: ", num)
		}
	} else {
		num, _ := o.Raw("select * from hdc_euler_black_user").QueryRows(&ebu)
		if num > 0 {
			logs.Info("QueryEulerBlackUserAll, num: ", num)
		}
	}
	return
}

func DelEulerUnassignBlack(id int64) {
	o := orm.NewOrm()
	err := o.Raw("delete from hdc_euler_unassign_user where id = ?", id).QueryRow()
	logs.Info("DelEulerUnassignBlack", err)
}

func QueryEulerUnassignUserAll(afterDate string) (euu []EulerUnassignUser) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from hdc_euler_unassign_user where " +
		"unassign_time < ? and unassign_time != ''", afterDate).QueryRows(&euu)
	if num > 0 {
		logs.Info("QueryEulerUnassignUserAll, num: ", num)
	}
	return
}

func QueryEulerUncompleteUserAll(afterDate string) (eiu []EulerIssueUser) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from hdc_euler_issue_user where " +
		"assign_time < ? and status = ? and assign_time != ''", afterDate, 1).QueryRows(&eiu)
	if num > 0 {
		logs.Info("QueryEulerUncompleteUserAll, num: ", num)
	}
	return
}

func QueryEulerUncompleteUserHistory() (eiu []EulerIssueUser) {
	o := orm.NewOrm()
	num, _ := o.Raw("select * from hdc_euler_issue_user where " +
		"status = ? and assign_time IS NULL", 1).QueryRows(&eiu)
	if num > 0 {
		logs.Info("QueryEulerUncompleteUserAll, num: ", num)
	}
	return
}