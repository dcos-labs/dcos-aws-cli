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

	// cli.HandleDefaultSections(app)

	handlePublicIPs(app)

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

	// nodesJson := out.String()




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

	result := gjson.Get(out.String(), `#[attributes.public_ip="true"].id`)
	result.ForEach(func(key, value gjson.Result) bool {
		sshExec := exec.Command(dcosCommand, "node", "ssh", "--option", "StrictHostKeyChecking=no", 
		 "--option", "LogLevel=quiet", "--master-proxy", "--mesos-id="+value.String(), `"curl -s ifconfig.co"`)
		sshExec.Stdout = os.Stdout
		//sshExec.Stderr = os.Stderr
		sshExec.Stdin = os.Stdin

		err := sshExec.Run()
		if err != nil {
			fmt.Printf("[Error during ssh to nodes.] %s\n\n", err)
			fmt.Printf("Make sure ssh keys are configured correctly.\n")

		}

		return true // keep iterating
})



	// m := f.(map[string]interface{})
	// for k, v := range m {
 //    switch vv := v.(type) {
 //    case string:
 //        fmt.Println(k, "is string", vv)
 //    case int:
 //        fmt.Println(k, "is int", vv)
 //    case []interface{}:
 //        fmt.Println(k, "is an array:")
 //        for i, u := range vv {
 //            fmt.Println(i, u)
 //        }
 //    default:
 //        fmt.Println(k, "is of a type I don't know how to handle")
 //    }
 //  }

	return nil
}

func handlePublicIPs(app *kingpin.Application) {
	app.Command("publicIPs", "Output Public IPs for public agents.").Action(publicAgentIPs)
}
