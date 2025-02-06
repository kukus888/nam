package main

import (
	"fmt"

	"resty.dev/v3"
)

type RundeckClient struct {
	Client *resty.Client
	url    string
	token  string
}

func NewRundeckClient(url string, token string) *RundeckClient {
	cli := resty.New()
	cli.SetHeader("X-Rundeck-Auth-Token", token)
	cli.SetBaseURL(url)
	cli.SetTrace(true) /// TODO: Remove this line for production
	return &RundeckClient{url: url, token: token, Client: cli}
}

func (rundeckClient *RundeckClient) GetSystemInfo() (*RundeckSystemInfo, error) {
	res, err := rundeckClient.Client.R().
		SetExpectResponseContentType("application/json").
		SetResult(&RundeckSystemInfo{}).
		Get("/api/21/system/info")
	if err != nil {
		fmt.Printf("Rundeck System Info: %+v\n", res)
	}
	dto := res.Result().(*RundeckSystemInfo)
	return dto, err
}

type RundeckSystemInfo struct {
	System struct {
		Executions struct {
			Active        string `json:"active"`
			ExecutionMode string `json:"executionMode"`
		} `json:"executions"`
		Extended    interface{} `json:"extended"`
		Healthcheck struct {
			ContentType string `json:"contentType"`
			Href        string `json:"href"`
		} `json:"healthcheck"`
		JVM struct {
			ImplementationVersion string `json:"implementationVersion"`
			Name                  string `json:"name"`
			Vendor                string `json:"vendor"`
			Version               string `json:"version"`
		} `json:"jvm"`
		Metrics struct {
			ContentType string `json:"contentType"`
			Href        string `json:"href"`
		} `json:"metrics"`
		OS struct {
			Arch    string `json:"arch"`
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"os"`
		Ping struct {
			ContentType string `json:"contentType"`
			Href        string `json:"href"`
		} `json:"ping"`
		Rundeck struct {
			APIVersion string `json:"apiversion"`
			Base       string `json:"base"`
			Build      string `json:"build"`
			BuildGit   string `json:"buildGit"`
			Node       string `json:"node"`
			ServerUUID string `json:"serverUUID"`
			Version    string `json:"version"`
		} `json:"rundeck"`
		Stats struct {
			CPU struct {
				LoadAverage struct {
					Average int    `json:"average"`
					Unit    string `json:"unit"`
				} `json:"loadAverage"`
				Processors int `json:"processors"`
			} `json:"cpu"`
			Memory struct {
				Free  int    `json:"free"`
				Max   int    `json:"max"`
				Total int    `json:"total"`
				Unit  string `json:"unit"`
			} `json:"memory"`
			Scheduler struct {
				Running        int `json:"running"`
				ThreadPoolSize int `json:"threadPoolSize"`
			} `json:"scheduler"`
			Threads struct {
				Active int `json:"active"`
			} `json:"threads"`
			Uptime struct {
				Duration int `json:"duration"`
				Since    struct {
					Datetime string `json:"datetime"`
					Epoch    int64  `json:"epoch"`
					Unit     string `json:"unit"`
				} `json:"since"`
				Unit string `json:"unit"`
			} `json:"uptime"`
		} `json:"stats"`
		ThreadDump struct {
			ContentType string `json:"contentType"`
			Href        string `json:"href"`
		} `json:"threadDump"`
		Timestamp struct {
			Datetime string `json:"datetime"`
			Epoch    int64  `json:"epoch"`
			Unit     string `json:"unit"`
		} `json:"timestamp"`
	} `json:"system"`
}
