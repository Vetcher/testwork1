package testwork1

import (
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"time"
)

type MetricsConf struct {
	Max1 int
	Max2 int
	Max3 int
	Max4 int
	Max5 int
	Max6 int
}

var metricsconfig MetricsConf

func init() {
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
	metricsconf, err := config.NewConfig("ini", "conf/metrics.conf")
	if err != nil {
		log.Print(err.Error())
	} else {
		metricsconfig.Max1, _ = metricsconf.Int("max1")
		metricsconfig.Max2, _ = metricsconf.Int("max2")
		metricsconfig.Max3, _ = metricsconf.Int("max3")
		metricsconfig.Max4, _ = metricsconf.Int("max4")
		metricsconfig.Max5, _ = metricsconf.Int("max5")
		metricsconfig.Max6, _ = metricsconf.Int("max6")
	}
}

type Device_metrics_input struct {
	Device_id int `json:"device_id"`
	Metric_1 int `json:"metric_1"`
	Metric_2 int `json:"metric_2"`
	Metric_3 int `json:"metric_3"`
	Metric_4 int `json:"metric_4"`
	Metric_5 int `json:"metric_5"`
	Metric_6 int `json:"metric_6"`
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
		Metric_6: dm.Metric_6,
	}
}

func WriteToDb(metrics *Device_metrics) (int64, error) {

	return orm.NewOrm().Insert(metrics)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	var pure []byte
	_, err := r.Body.Read(pure)
	if err != nil {
		log.Print("INFO:", err.Error())
		fmt.Fprint(w, err.Error())
	} else {
		var d Device_metrics_input
		err = json.Unmarshal(pure, &d)
		if err != nil {
			log.Print("INFO:", err.Error())
			fmt.Fprint(w, err.Error())
		} else {
			n, err := WriteToDb(&d.translate())
			if err != nil {
				log.Print("INFO:", err.Error())
				fmt.Fprint(w, err.Error())
			} else {
				fmt.Fprint(w, n)
			}
		}
	}
}

func main() {
	http.HandleFunc("/", MainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
