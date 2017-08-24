package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/mesosphere/dcos-commons/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := cli.New()

	handlePublicIPs(app)
	handleExposedApps(app)

	kingpin.MustParse(app.Parse(cli.GetArguments()))
}

func runDcosCommand(arg ...string) {
	var out bytes.Buffer
	cmd := exec.Command("dcos", arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[Error] %s\n\n", err)
		fmt.Printf("Unable to run DC/OS command: %s\n", strings.Join(arg, " "))
		fmt.Printf("Make sure your PATH includes the 'dcos' executable.\n")
	}
}

func publicAgentIPs(c *kingpin.ParseContext) error {
	var out bytes.Buffer
	dcosCommand := "dcos"
	nodeListExec := exec.Command(dcosCommand, "node", "--json")
	nodeListExec.Stdin = os.Stdin
	nodeListExec.Stdout = &out
	nodeListExec.Stderr = os.Stderr

	err := nodeListExec.Run()
	if err != nil {
		fmt.Printf("[Error] %s\n\n", err)
		fmt.Printf("Unable to run DC/OS command")
		fmt.Printf("Make sure your PATH includes the 'dcos' executable.\n")
	}

	var f interface{}
	err = json.Unmarshal(out.Bytes(), &f)
	if err != nil {
		fmt.Printf("[Error] %s\n\n", err)
	}

	result := gjson.Get(out.String(), `#[attributes.public_ip="true"]#.id`)
	result.ForEach(func(key, value gjson.Result) bool {
		sshExec := exec.Command(dcosCommand, "node", "ssh", "--option", "StrictHostKeyChecking=no", 
		 "--option", "LogLevel=quiet", "--master-proxy", "--mesos-id="+value.String(), `"curl -s ifconfig.co"`)
		sshExec.Stdout = os.Stdout
		sshExec.Stdin = os.Stdin

		err := sshExec.Run()
		if err != nil {
			fmt.Printf("[Error during ssh to nodes.] %s\n\n", err)
			fmt.Printf("Make sure ssh keys are configured correctly.\n")
		}
		return true // keep iterating
	})
	
	return nil
}

func handlePublicIPs(app *kingpin.Application) {
	app.Command("publicIPs", "Output Public IPs for public agents.").Action(publicAgentIPs)
}

func exposedApps(c *kingpin.ParseContext) error {
	var out bytes.Buffer
	dcosCommand := "dcos"
	nodeListExec := exec.Command(dcosCommand, "node", "--json")
	nodeListExec.Stdin = os.Stdin
	nodeListExec.Stdout = &out
	nodeListExec.Stderr = os.Stderr

	err := nodeListExec.Run()
	if err != nil {
		fmt.Printf("[Error] %s\n\n", err)
		fmt.Printf("Unable to run DC/OS command")
		fmt.Printf("Make sure your PATH includes the 'dcos' executable.\n")
	}

	var f interface{}
	err = json.Unmarshal(out.Bytes(), &f)
	if err != nil {
		fmt.Printf("[Error] %s\n\n", err)
	}

	var publicAgentMesosIDs []string
	result := gjson.Get(out.String(), `#[attributes.public_ip="true"]#.id`)
	result.ForEach(func(key, value gjson.Result) bool {
	// For each public agent in the cluster
		publicAgentMesosIDs = append(publicAgentMesosIDs, value.String())
		return true
	})

	for _, mesosID := range publicAgentMesosIDs {
		// get public IP
		sshExec := exec.Command(dcosCommand, "node", "ssh", "--option", "StrictHostKeyChecking=no", 
		 "--option", "LogLevel=quiet", "--master-proxy", "--mesos-id="+mesosID, `"curl -s ifconfig.co"`)
		var publicIpByte bytes.Buffer
		sshExec.Stdout = &publicIpByte
		sshExec.Stdin = os.Stdin

		err := sshExec.Run()
		if err != nil {
			fmt.Printf("[Error during ssh to nodes.] %s\n\n", err)
			fmt.Printf("Make sure ssh keys are configured correctly.\n")
		}


		// get apps running on that public agent 
		sshExec = exec.Command(dcosCommand, "task", "--json")
		var tasksJson bytes.Buffer
		sshExec.Stdout = &tasksJson
		sshExec.Stdin = os.Stdin

		err = sshExec.Run()
		if err != nil {
			fmt.Printf("[Error running dcos task command.] %s\n\n", err)
			fmt.Printf("Make sure ssh keys are configured correctly.\n")
		}

		tasks := gjson.Get(tasksJson.String(), `#[slave_id="`+ mesosID +`"]#`)
		tasks.ForEach(func(key1, innerValue gjson.Result) bool {
			name := gjson.Get(innerValue.String(), "name")
			ports := gjson.Get(innerValue.String(), "resources.ports")

			if (ports.Exists()) {
				fmt.Printf("AppName: %s, PublicIP: %s, Ports: %s \n", name.String(), strings.TrimSpace(publicIpByte.String()), ports.String() )
			}
			
			return true
		})
	}
	
	return nil
}

func handleExposedApps(app *kingpin.Application) {
	app.Command("exposedApps", "Display apps and their ports running on public agents.").Action(exposedApps)
}
