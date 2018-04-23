// +build !windows

package main

type WinEventLogCrawler struct {
	Rules       []*RuleConfig
	SystemState *SystemState
	NextChannel chan *SecurityEventsContainer
	Options     *Options
}

func (crawler *WinEventLogCrawler) Init() {}

func (crawler *WinEventLogCrawler) Run() {}
