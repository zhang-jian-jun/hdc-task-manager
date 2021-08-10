package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type PointValue struct {
	Integration int64
}

func QueryOpenEulerIssueAll() (eoi []EulerOriginIssue) {
	o := orm.NewOrm()
	var num int64
	num, err := o.Raw("select * from hdc_euler_origin_issue").QueryRows(&eoi)
	if num > 0 {
		logs.Info("QueryOpenEulerIssueAll, num: ", num)
	} else {
		logs.Error("QueryOpenEulerIssueAll, err: ", err)
	}
	return
}

func QueryOpenEulerUserAll() (eu []EulerUser) {
	o := orm.NewOrm()
	var num int64
	num, err := o.Raw("select * from hdc_euler_user").QueryRows(&eu)
	if num > 0 {
		logs.Info("QueryOpenEulerUserAll, num: ", num)
	} else {
		logs.Error("QueryOpenEulerUserAll, err: ", err)
	}
	return
}

func QueryOpenGaussUserAll() (gu []GaussUser) {
	o := orm.NewOrm()
	var num int64
	num, err := o.Raw("select * from hdc_gauss_user").QueryRows(&gu)
	if num > 0 {
		logs.Info("QueryOpenGaussUserAll, num: ", num)
	} else {
		logs.Error("QueryOpenGaussUserAll, err: ", err)
	}
	return
}

func QueryEulerUserIntegDetailValue(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select sum(integral_value) as integration FROM hdc_euler_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ?", userId, startTime, endTime).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailValue, err: ", err)
		}
	} else {
		err := o.Raw("select sum(integral_value) as integration "+
			"FROM hdc_euler_user_integ_detail where user_id = ?", userId).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailValue, err: ", err)
		}
	}
	return
}

func QueryEulerUserIntegDetailCount(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select count(or_id) as integration FROM hdc_euler_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ? and integral_value > ?",
			userId, startTime, endTime, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailCount, err: ", err)
		}
	} else {
		err := o.Raw("select count(or_id) as integration FROM hdc_euler_user_integ_detail "+
			"where user_id = ? and integral_value > ?", userId, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryEulerUserIntegDetailCount, err: ", err)
		}
	}
	return
}

func (elt *EmailList) Read(field ...string) ([]EmailList, error) {
	o := orm.NewOrm()
	var el []EmailList
	var num int64
	num, err := o.Raw("select *"+
		" from hdc_email_list where email_type = ?", elt.EmailType).QueryRows(&el)
	if err == nil && num > 0 {
		return el, nil
	}
	logs.Error("hdc_email_list ,err: ", err)
	return el, err
}

func QueryGaussUserIntegDetailValue(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select sum(integral_value) as integration from hdc_gauss_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ?", userId, startTime, endTime).QueryRow(pv)
		if err != nil {
			logs.Error("QueryGaussUserIntegDetailValue, err: ", err)
		}
	} else {
		err := o.Raw("select sum(integral_value) as integration "+
			"FROM hdc_gauss_user_integ_detail where user_id = ?", userId).QueryRow(pv)
		if err != nil {
			logs.Error("QueryGaussUserIntegDetailValue, err: ", err)
		}
	}
	return
}

func QueryGaussUserIntegDetailCount(pv *PointValue, startTime, endTime string, userId int64) {
	o := orm.NewOrm()
	if len(startTime) > 1 {
		err := o.Raw("select count(or_id) as integration FROM hdc_gauss_user_integ_detail "+
			"where user_id = ? and create_time >= ? and create_time < ? and integral_value > ?",
			userId, startTime, endTime, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryGaussUserIntegDetailCount, err: ", err)
		}
	} else {
		err := o.Raw("select count(or_id) as integration FROM hdc_gauss_user_integ_detail "+
			"where user_id = ? and integral_value > ?", userId, 0).QueryRow(pv)
		if err != nil {
			logs.Error("QueryGaussUserIntegDetailCount, err: ", err)
		}
	}
	return
}
