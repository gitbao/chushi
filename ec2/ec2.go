package ec2

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gitbao/gitbao/model"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
)

var client *ec2.EC2

func init() {

	accessKey := os.Getenv("GITBAO_AWS_KEY")
	secretKey := os.Getenv("GITBAO_AWS_SECRET")
	auth, err := aws.GetAuth(accessKey, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	client = ec2.New(auth, aws.USEast)
	// resp, err := client.DescribeAvailabilityZones(nil)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Printf("%#v", resp)
}

func CreateInstance() (server model.Server) {
	securityGroup := ec2.SecurityGroup{
		Id: "sg-77a2721a",
	}
	runInstance := ec2.RunInstances{
		ImageId:        "ami-9a562df2", //ubuntu 14.04
		InstanceType:   "m3.medium",
		KeyName:        "dev",
		MaxCount:       1,
		MinCount:       1,
		SecurityGroups: []ec2.SecurityGroup{securityGroup},
	}
	resp, err := client.RunInstances(&runInstance)
	instanceId := resp.Instances[0].InstanceId
	fmt.Println(resp.Instances[0].InstanceId)
	log.Printf("%#v\n", resp)
	log.Printf("%#v\n", err)
	for {
		resp, err := GetInstanceInfo(instanceId)
		_ = err
		if resp.InstanceStatus[0].InstanceState.Name == "running" {
			break
		}
		time.Sleep(time.Second * 5)
	}
	publicIp, err := GetIp(instanceId)
	fmt.Println(publicIp)
	server = model.Server{
		Ip:         publicIp,
		InstanceId: instanceId,
	}
	// DestroyInstance(instanceId)

	return
}

func GetInstanceInfo(instanceId string) (resp *ec2.DescribeInstanceStatusResp, err error) {

	describeInstanceStatus := ec2.DescribeInstanceStatus{
		InstanceIds:         []string{instanceId},
		IncludeAllInstances: true,
	}
	resp, err = client.DescribeInstanceStatus(&describeInstanceStatus, nil)
	fmt.Printf("%#v\n", resp.InstanceStatus[0].InstanceState.Name)
	return
}

func DestroyInstance(instanceId string) (err error) {
	resp, err := client.TerminateInstances([]string{instanceId})
	log.Printf("%#v\n", resp)
	log.Printf("%#v\n", err)
	return err
}

func GetIp(instanceId string) (publicIp string, err error) {
	resp, err := client.Instances([]string{instanceId}, nil)
	if err != nil {
		return "", err
	}
	return resp.Reservations[0].Instances[0].PublicIpAddress, err
}
