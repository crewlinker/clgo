package clecsmeta

import (
	"encoding/json"
	"time"
)

const (
	labelEcsCluster               = "com.amazonaws.ecs.cluster"
	labelEcsContainerName         = "com.amazonaws.ecs.container-name"
	labelEcsTaskArn               = "com.amazonaws.ecs.task-arn"
	labelEcsTaskDefinitionFamily  = "com.amazonaws.ecs.task-definition-family"
	labelEcsTaskDefinitionVersion = "com.amazonaws.ecs.task-definition-version"
)

type ContainerMetadataV4 struct {
	DockerID      string    `json:"DockerId"`
	Name          string    `json:"Name"`
	DockerName    string    `json:"DockerName"`
	Image         string    `json:"Image"`
	ImageID       string    `json:"ImageID"`
	Labels        LabelsV4  `json:"Labels"`
	DesiredStatus string    `json:"DesiredStatus"`
	KnownStatus   string    `json:"KnownStatus"`
	Limits        Limits    `json:"Limits"`
	CreatedAt     time.Time `json:"CreatedAt"`
	StartedAt     time.Time `json:"StartedAt"`
	Type          string    `json:"Type"`
	ContainerARN  string    `json:"ContainerARN"`
	LogDriver     string    `json:"LogDriver"`
	LogOptions    struct {
		AwsLogsCreateGroup string `json:"awslogs-create-group"`
		AwsLogsGroup       string `json:"awslogs-group"`
		AwsLogsStream      string `json:"awslogs-stream"`
		AwsRegion          string `json:"awslogs-region"`
	} `json:"LogOptions"`
	Networks []struct {
		NetworkMode              string   `json:"NetworkMode"`
		IPv4Addresses            []string `json:"IPv4Addresses"`
		AttachmentIndex          int      `json:"AttachmentIndex"`
		IPv4SubnetCIDRBlock      string   `json:"IPv4SubnetCIDRBlock"`
		MACAddress               string   `json:"MACAddress"`
		DomainNameServers        []string `json:"DomainNameServers"`
		DomainNameSearchList     []string `json:"DomainNameSearchList"`
		PrivateDNSName           string   `json:"PrivateDNSName"`
		SubnetGatewayIpv4Address string   `json:"SubnetGatewayIpv4Address"`
	} `json:"Networks"`
}

type Limits struct {
	CPU    float64 `json:"CPU"`
	Memory int     `json:"Memory"`
}

type LabelsV4 struct {
	EcsCluster               string `json:"-"`
	EcsContainerName         string `json:"-"`
	EcsTaskArn               string `json:"-"`
	EcsTaskDefinitionFamily  string `json:"-"`
	EcsTaskDefinitionVersion string `json:"-"`

	rest map[string]string
}

func (l LabelsV4) Get(name string) string {
	return l.rest[name]
}

func (l *LabelsV4) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &l.rest); err != nil {
		return err //nolint:wrapcheck
	}

	if cluster, ok := l.rest[labelEcsCluster]; ok {
		l.EcsCluster = cluster
		delete(l.rest, labelEcsCluster)
	}

	if containerName, ok := l.rest[labelEcsContainerName]; ok {
		l.EcsContainerName = containerName
		delete(l.rest, labelEcsContainerName)
	}

	if taskArn, ok := l.rest[labelEcsTaskArn]; ok {
		l.EcsTaskArn = taskArn
		delete(l.rest, labelEcsTaskArn)
	}

	if family, ok := l.rest[labelEcsTaskDefinitionFamily]; ok {
		l.EcsTaskDefinitionFamily = family
		delete(l.rest, labelEcsTaskDefinitionFamily)
	}

	if version, ok := l.rest[labelEcsTaskDefinitionVersion]; ok {
		l.EcsTaskDefinitionVersion = version
		delete(l.rest, labelEcsTaskDefinitionVersion)
	}

	return nil
}

type TaskMetadataV4 struct {
	Cluster          string                `json:"Cluster"`
	TaskARN          string                `json:"TaskARN"`
	Family           string                `json:"Family"`
	Revision         string                `json:"Revision"`
	DesiredStatus    string                `json:"DesiredStatus"`
	KnownStatus      string                `json:"KnownStatus"`
	Limits           Limits                `json:"Limits"`
	PullStartedAt    time.Time             `json:"PullStartedAt"`
	PullStoppedAt    time.Time             `json:"PullStoppedAt"`
	AvailabilityZone string                `json:"AvailabilityZone"`
	LaunchType       string                `json:"LaunchType"`
	Containers       []ContainerMetadataV4 `json:"Containers"`
}
