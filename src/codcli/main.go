package main

import (
	"codcli/launcher"
	"components/ambari"
	"flag"
	"fmt"
	"kube"
	"log"
	"os"
	"reflect"
	"strconv"
	"util"

	"github.com/spf13/viper"
)

func main() {

	util.SetDefaultConfig()

	if len(os.Args) == 1 {
		fmt.Println("usage: cod <command> [<args>]")
		fmt.Println("Commands: ")
		fmt.Println("\tstart   starts the kubernetes cluster")
		fmt.Println("\tstop    stops the kubernetes cluster")
		fmt.Println("\trestart stops the current cluster and restarts a new one")
		fmt.Println("\treset   removes all deployed components")
		fmt.Println("\tinfo    displays the cluster info")
		fmt.Println("\tpods    lists all pods running on the cluster")
		fmt.Println("\trun     runs an application on spark")
		fmt.Println("\tdeploy  deploys bdp components on a running cluster")
		return
	}

	runCommand := flag.NewFlagSet("run", flag.ExitOnError)
	deployCommand := flag.NewFlagSet("deploy", flag.ExitOnError)
	scaleCommand := flag.NewFlagSet("scale", flag.ExitOnError)

	jarFlag := runCommand.String("jar", "", "")
	gitFlag := runCommand.String("git", "", "")
	urlFlag := runCommand.String("url", "", "")

	allFlag := deployCommand.Bool("all", false, "")
	confFlag := deployCommand.String("conf", "", "")
	clusterFlag := deployCommand.String("cluster", "", "")
	forceFlag := deployCommand.Bool("f", false, "force the deployment by removing any running instances of the components")

	switch os.Args[1] {
	case "start":
		kube.StartCluster()
	case "stop":
		kube.StopCluster()
	case "restart":
		kube.StopCluster()
		kube.StartCluster()
	case "deploy":
		deployCommand.Parse(os.Args[2:])
	case "reset":
		kube.ResetCluster()
	case "run":
		runCommand.Parse(os.Args[2:])
	case "scale":
		scaleCommand.Parse(os.Args[2:])
	case "pods":
		fmt.Println(kube.GetPods())
	case "info":
		fmt.Println(kube.ClusterInfo())
	case "kube":
		fmt.Println(kube.ExecOnKube(os.Args[2], os.Args[3:]...))
	case "test":
		test()
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if scaleCommand.Parsed() {
		if len(os.Args[2:]) == 0 {
			fmt.Println("usage: cod scale controller-name size")
			os.Exit(2)
		}
		size, err := strconv.Atoi(scaleCommand.Arg(1))
		if err != nil {
			fmt.Println("invalide size argument")
			fmt.Println("usage: cod scale controller-name size")
			os.Exit(2)
		}
		kube.ScaleController(scaleCommand.Arg(0), size)
	}
	if runCommand.Parsed() {
		if len(os.Args[2:]) == 0 {
			fmt.Println("usage: cod run [<args>]")
			fmt.Println("args: ")
			fmt.Println("  -jar	   path to the jar file of the application")
			fmt.Println("  -git    link to the git repository of the application")
			fmt.Println("  -url    link for direct download of the application")
			os.Exit(2)
		}
		if *jarFlag != "" && (*gitFlag != "" || *urlFlag != "") {
			launcher.LaunchApplication(*jarFlag, *gitFlag, os.Args[6:]...)
		}
	}

	if deployCommand.Parsed() {
		if len(os.Args[2:]) == 0 {
			fmt.Println("usage: cod deploy [<args>]")
			fmt.Println("args: ")
			fmt.Println("  -conf    path to a toml config file")
			fmt.Println("  -f       force deployment (removes running components first)")
			fmt.Println("  -cluster the kubernetes context to use for deployment")
			fmt.Println("  -all     to deploy all components with default configuration")
			os.Exit(2)
		}
		if *confFlag != "" {
			log.Printf("Loading configuration file %s \n", *confFlag)
			viper.SetConfigFile(*confFlag)

			err := viper.ReadInConfig()
			if err != nil {
				log.Fatalf("Error loading config file: %s \n", err)
			}
			//fmt.Println(viper.AllSettings())
		}

		if *clusterFlag != "" {
			kube.SetContext(*clusterFlag)
		}
		config := util.ConfigStruct()
		//fmt.Println(config)
		statuses := launcher.LaunchComponents(*allFlag, *forceFlag, config)
		v := reflect.ValueOf(statuses)
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).FieldByName("State").Bool() {
				fmt.Print(v.Field(i).FieldByName("Message"))
				fmt.Print(v.Field(i).FieldByName("URL"))
				fmt.Println("")
			}
		}
	}
}

func test() {
	kube.Expose("pod", ambari.GetNamenode(), "--port=8020", "--target-port=8020", "--name=namenode", "--type=NodePort")
}
