package testwork1

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type User struct {
	Id    int `orm:"column(id);pk;auto"`
	Name  string `orm:"column(name)"`
	Email string `orm:"column(email)"`
}

func (t *User) TableName() string {
	return "users"
}

type Device struct {
	Id    int `orm:"column(id);pk;auto"`
	Name  string `orm:"column(name)"`
	User_id *User `orm:"column(user_id);rel(fk)"`
}

func (t *Device) TableName() string {
	return "devices"
}

// Значения метрик должны быть всегда положительны,
// иначе нужно их делать положительными при помощи операций + и -
type Device_metrics struct {
	Id    int `orm:"column(id);pk;auto"`
	Device_id *Device `orm:"column(device_id);rel(fk)"`
	Metric_1 int `orm:"column(metric_1);"`
	Metric_2 int `orm:"column(metric_2)"`
	Metric_3 int `orm:"column(metric_3)"`
	Metric_4 int `orm:"column(metric_4)"`
	Metric_5 int `orm:"column(metric_5)"`
	Metric_6 int `orm:"column(metric_6)"`
	Local_time time.Time `orm:"column(local_time);type(datetime)"`
}

func (t *Device_metrics) TableName() string {
	return "device_metrics"
}

type Device_alerts struct {
	Id    int `orm:"column(id);pk;auto"`
	Device_id *Device `orm:"column(device_id);rel(fk)"`
	Message string `orm:"column(message)"`
}

func (t *Device_alerts) TableName() string {
	return "device_alerts"
}

func init() {
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Device))
	orm.RegisterModel(new(Device_metrics))
	orm.RegisterModel(new(Device_alerts))
}