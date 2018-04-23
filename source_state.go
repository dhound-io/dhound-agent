package main

type SourceState struct {
	SourceId                 string `json:"srcid"`
	Source                   string `json:"src"`
	Offset                   int64  `json:"offset"`
	LastUpdatedTimeUtcNumber int64  `json:"t,omitempty"`
}
