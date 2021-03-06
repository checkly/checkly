package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/checkly/checkly-go-sdk"
)

var alertSettings = checkly.AlertSettings{
	EscalationType: checkly.RunBased,
	RunBasedEscalation: checkly.RunBasedEscalation{
		FailedRunThreshold: 1,
	},
	TimeBasedEscalation: checkly.TimeBasedEscalation{
		MinutesFailingThreshold: 5,
	},
	Reminders: checkly.Reminders{
		Interval: 5,
	},
	SSLCertificates: checkly.SSLCertificates{
		Enabled:        false,
		AlertThreshold: 3,
	},
}

var apiCheck = checkly.Check{
	Name:                 "My API Check",
	Type:                 checkly.TypeAPI,
	Frequency:            5,
	DegradedResponseTime: 5000,
	MaxResponseTime:      15000,
	Activated:            true,
	Muted:                false,
	ShouldFail:           false,
	DoubleCheck:          false,
	SSLCheck:             true,
	LocalSetupScript:     "",
	LocalTearDownScript:  "",
	Locations: []string{
		"eu-west-1",
		"ap-northeast-2",
	},
	Tags: []string{
		"foo",
		"bar",
	},
	AlertSettings:          alertSettings,
	UseGlobalAlertSettings: false,
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		Body:     "",
		BodyType: "NONE",
	},
}

var browserCheck = checkly.Check{
	Name:          "My Browser Check",
	Type:          checkly.TypeBrowser,
	Frequency:     5,
	Activated:     true,
	Muted:         false,
	ShouldFail:    false,
	DoubleCheck:   false,
	SSLCheck:      true,
	Locations:     []string{"eu-west-1"},
	AlertSettings: alertSettings,
	Script: `const assert = require("chai").assert;
	const puppeteer = require("puppeteer");

	const browser = await puppeteer.launch();
	const page = await browser.newPage();
	await page.goto("https://example.com");
	const title = await page.title();

	assert.equal(title, "Example Site");
	await browser.close();`,
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "HELLO",
			Value: "Hello world",
		},
	},
	Request: checkly.Request{
		Method: http.MethodGet,
		URL:    "http://example.com",
	},
}

var group = checkly.Group{
	Name:        "test",
	Activated:   true,
	Muted:       false,
	Tags:        []string{"auto"},
	Locations:   []string{"eu-west-1"},
	Concurrency: 3,
	APICheckDefaults: checkly.APICheckDefaults{
		BaseURL: "example.com/api/test",
		Headers: []checkly.KeyValue{
			{
				Key:   "X-Test",
				Value: "foo",
			},
		},
		QueryParameters: []checkly.KeyValue{
			{
				Key:   "query",
				Value: "foo",
			},
		},
		Assertions: []checkly.Assertion{
			{
				Source:     checkly.StatusCode,
				Comparison: checkly.Equals,
				Target:     "200",
			},
		},
		BasicAuth: checkly.BasicAuth{
			Username: "user",
			Password: "pass",
		},
	},
	EnvironmentVariables: []checkly.EnvironmentVariable{
		{
			Key:   "ENVTEST",
			Value: "Hello world",
		},
	},
	DoubleCheck:            true,
	UseGlobalAlertSettings: false,
	AlertSettings: checkly.AlertSettings{
		EscalationType: checkly.RunBased,
		RunBasedEscalation: checkly.RunBasedEscalation{
			FailedRunThreshold: 1,
		},
		TimeBasedEscalation: checkly.TimeBasedEscalation{
			MinutesFailingThreshold: 5,
		},
		Reminders: checkly.Reminders{
			Amount:   0,
			Interval: 5,
		},
		SSLCertificates: checkly.SSLCertificates{
			Enabled:        true,
			AlertThreshold: 30,
		},
	},
	AlertChannelSubscriptions: []checkly.Subscription{
		{
			Activated: true,
		},
	},
	LocalSetupScript:    "setup-test",
	LocalTearDownScript: "teardown-test",
}

func main() {
	apiKey := os.Getenv("CHECKLY_API_KEY")
	if apiKey == "" {
		log.Fatal("no CHECKLY_API_KEY set")
	}
	client := checkly.NewClient(apiKey)
	// uncomment this to enable dumping of API requests and responses
	// client.Debug = os.Stdout
	group, err := client.CreateGroup(group)
	if err != nil {
		log.Fatalf("creating group: %v", err)
	}
	fmt.Printf("New check group created with ID %d\n", group.ID)
	for _, check := range []checkly.Check{apiCheck, browserCheck} {
		gotCheck, err := client.Create(check)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("New check created with ID %s\n", gotCheck.ID)
	}
}
