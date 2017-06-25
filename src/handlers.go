package ppp


import (
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"time"
	"gopkg.in/gomail.v2"
)

type MetricsConfiguration struct {
	Max [5]int
}

var MetricsConfig MetricsConfiguration

type EmailConfiguration struct {
	Host string
	Port int
	Username string
	Password string
}

var EmailConfig EmailConfiguration

// Описывает интерфейс функции оповещения
type AlertFunction func(string, *Device)

// Функции, отправляющие сообщения с оповещениями
var ListOfAlertFunctions []AlertFunction

func init() {
	// База данных
	dbconf, err := config.NewConfig("ini", "conf/postgre.conf")
	if err != nil {
		log.Panic(err)
	}
	// "postgres://login:password@host:port/database?sslmode=disable"
	postgresStrConfig := "postgres://" + dbconf.String("login") + ":" +
			dbconf.String("password") + "@" + dbconf.String("host") + ":" + dbconf.String("port") + "/" +
			dbconf.String("database") + "?sslmode=" + dbconf.String("sslmode")
	err = orm.RegisterDataBase("default", "postgres", postgresStrConfig)
	if err != nil {
		log.Panic(err)
		panic(err)
	}

	// Метрики
	metricsconf, err := config.NewConfig("ini", "conf/metrics.conf")
	if err != nil {
		log.Print(err.Error())
	} else {
		MetricsConfig.Max[0], _ = metricsconf.Int("max1")
		MetricsConfig.Max[1], _ = metricsconf.Int("max2")
		MetricsConfig.Max[2], _ = metricsconf.Int("max3")
		MetricsConfig.Max[3], _ = metricsconf.Int("max4")
		MetricsConfig.Max[4], _ = metricsconf.Int("max5")
	}

	// Оповещения
	ListOfAlertFunctions = append(ListOfAlertFunctions, WriteAlertToDatabase)
	ListOfAlertFunctions = append(ListOfAlertFunctions, SendEmail)

	// Email
	emailconf, err := config.NewConfig("ini", "conf/email.conf")
	if err != nil {
		log.Print("CRITICAL:", "Email client is not configured")
	} else {
		EmailConfig.Host = emailconf.String("host")
		if p, err := emailconf.Int("host"); err != nil {
			EmailConfig.Port = 587
		} else {
			EmailConfig.Port = p
		}
		EmailConfig.Username = emailconf.String("username")
		EmailConfig.Password = emailconf.String("password")
		// https://godoc.org/gopkg.in/gomail.v2
		EmailChannel = make(chan *gomail.Message)
		go EmailSender(EmailChannel, EmailConfig)
	}
}

type Device_metrics_input struct {
	Device_id int `json:"device_id"`
	Metric_1 int `json:"metric_1"`
	Metric_2 int `json:"metric_2"`
	Metric_3 int `json:"metric_3"`
	Metric_4 int `json:"metric_4"`
	Metric_5 int `json:"metric_5"`
	Local_time time.Time `json:"local_time"`
}

func (dm *Device_metrics_input) translate() Device_metrics {
	return Device_metrics{
		Device_id: &Device{
			Id: dm.Device_id,
		},
		Local_time: dm.Local_time,
		Metric_1: dm.Metric_1,
		Metric_2: dm.Metric_2,
		Metric_3: dm.Metric_3,
		Metric_4: dm.Metric_4,
		Metric_5: dm.Metric_5,
	}
}

const ALERTTEMPLATE = "Metric %d too high/low, it is %d, but should be less than %d\n"

func HandleMetrics(metrics *Device_metrics) (int64, error) {
	metric_vals := []int{metrics.Metric_1,metrics.Metric_2,metrics.Metric_3,metrics.Metric_4,metrics.Metric_5}
	alert_msg := ""
	ok := true
	for i, m := range metric_vals {
		if MetricsConfig.Max[i] < m || m < 0 {
			ok = false
			alert_msg += fmt.Sprintf(ALERTTEMPLATE, i, m, MetricsConfig.Max[i])
		}
	}
	if !ok {
		// alert user
		go AlertStarter(alert_msg, metrics.Device_id)
	}
	return orm.NewOrm().Insert(metrics)
}

// Действия при ошибке метрик
// TODO: нужна проверка валидности `Device`
func AlertStarter(message string, device *Device) {
	err := orm.NewOrm().QueryTable("devices").RelatedSel().One(device)
	if err == nil {
		for _, f := range ListOfAlertFunctions {
			go f(message, device)
		}
	} else {
		log.Print("CRITICAL:", err.Error())
	}
}

// Записываем оповещение в бд
func WriteAlertToDatabase(message string, device *Device)  {
	_, err := orm.NewOrm().Insert(&Device_alerts{
		Device_id: device,
		Message: message,
	})
	if err != nil {
		log.Print("CRITICAL:", err.Error())
	}
}

// Формируем и отправляем оповещение на почту
func SendEmail(message string, device *Device) {
	m := gomail.NewMessage()
	m.SetHeader("To", device.User_id.Email)
	m.SetHeader("Subject", "Metric alert")
	m.SetBody("text/html", message)
	EmailChannel <- m
}

/*
Здесь должен быть код с записью сообщения в Redis, но т.к. я на винде, я его пропущу
*/

func MainHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body) // decode to json
	defer r.Body.Close()
	var d Device_metrics_input
	err := decoder.Decode(&d)
	if err != nil {
		log.Print("INFO:", err.Error())
		fmt.Fprint(w, err.Error())
	} else {
		x := d.translate()
		n, err := HandleMetrics(&x)
		if err != nil {
			log.Print("INFO:", err.Error())
			fmt.Fprint(w, err.Error())
		} else {
			fmt.Fprint(w, n)
		}
	}
}

var EmailChannel chan *gomail.Message

// Держит канал связи с SMTP сервером
func EmailSender(ch chan *gomail.Message, conf EmailConfiguration) {
	d := gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)

	var s gomail.SendCloser
	var err error
	open := false
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return
			}
			if !open {
				if s, err = d.Dial(); err != nil {
					log.Print(err.Error())
				} else {
					open = true
				}
			}
			m.SetHeader("From", conf.Username)
			if err := gomail.Send(s, m); err != nil {
				log.Print(err)
			}
			// Close the connection to the SMTP server if no email was sent in
			// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := s.Close(); err != nil {
					log.Print(err.Error())
				} else {
					open = false
				}
			}
		}
	}
}
