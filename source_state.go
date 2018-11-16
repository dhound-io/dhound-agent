package main

type SourceState struct {
	SourceId                 string `json:"srcid"`
	Source                   string `json:"src"`
	Offset                   int64  `json:"offset"`
	Line                     int64  `json:"line,omitempty"`
	LastUpdatedTimeUtcNumber int64  `json:"t,omitempty"`
}
