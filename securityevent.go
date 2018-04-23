package main

type SecurityEvent struct {
	SecurityId         uint              `json:"sid"`
	SecurityGroupId    uint              `json:"gid,omitempty"`
	EventTimeUtcNumber int64             `json:"t"`
	Message            string            `json:"m,omitempty"`
	IpAddress          string            `json:"ip"`
	AdditionalFields   map[string]string `json:"a,omitempty"`
	Critical           bool              `json:"-"`
	Source             *string           `json:"src,omitempty"`
}
