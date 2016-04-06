package rabbitmq

import (
	"fmt"
	"kube"
	"log"
	"time"
	"util"

	"github.com/spf13/viper"
)

func CleanUp() {
	log.Println("Rabbitmq: cleaning up cluster...")
	kube.DeleteResource("rc", "rabbitmq-controller")
	kube.DeleteResource("svc", "rabbitmq")

	for {
		remaining := kube.RemainingPods("rabbitmq")
		if remaining == 0 {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}
	util.ReleasePID("rabbitmq")
}

func Start(config util.Config, forceDeploy bool) {
	if !forceDeploy {
		if util.IsRunning("rabbitmq") {
			log.Println("Rabbitmq: already running, skipping start ...")
			return
		}
	}
	CleanUp()

	log.Println("Rabbitmq: Launching rabbitmq")
	util.GenerateConfig("rabbitmq-controller.json", "rabbitmq", config)
	kube.CreateResource(viper.GetString("BDP_CONFIG_DIR") + "/tmp/rabbitmq-controller.json")

	log.Println("Rabbitmq: Waiting for Rabbitmq pods to start...")
	for {
		pending := kube.PendingPods()
		if pending == 0 {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	kube.CreateResource(viper.GetString("BDP_CONFIG_DIR") + "/rabbitmq/rabbitmq-service.json")
	util.SetPID("rabbitmq")
	log.Println("Rabbitmq: done")

}

func Status() util.Status {
	status := util.Status{State: false, Message: "Not Running", URL: ""}
	if util.IsRunning("rabbitmq") {
		status.State = true
		status.Message = fmt.Sprintf("RabbitMQ UI accessible through ")
		status.URL = fmt.Sprintf("http://%s:31316", kube.PodPublicIP("rabbitmq"))
	}
	return status
}
