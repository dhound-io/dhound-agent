package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"
)

const SystemStateFileName string = ".state/.dhound-state"

type SystemState struct {
	Sources []*SourceState                  `json:"s"`
	Input   chan []*SecurityEventsContainer `json:"-"`
}

func (state *SystemState) Sync() {

	for eventsContainers := range state.Input {
		if len(eventsContainers) > 0 {

			originalState := &SystemState{}
			originalState.ReadOriginalState()

			for _, eventsContainer := range eventsContainers {

				sourceId := (*eventsContainer).SourceId

				sourceState := originalState.Find(sourceId)

				sourceState.Offset = eventsContainer.Offset
				sourceState.Source = eventsContainer.Source

				sourceState.LastUpdatedTimeUtcNumber = DateToCustomLong(time.Now())
			}

			originalState.Save()
		}
	}
}

func (state *SystemState) Save() {

	CreateDirIfNotExist(filepath.Dir(SystemStateFileName), 0664)

	content, _ := json.Marshal(state)
	err := ioutil.WriteFile(SystemStateFileName, content, 0664)
	if err != nil {
		emitLine(logLevel.important, "Failed to save state of the sources (%s).", SystemStateFileName, err.Error())
	}
}

func (state *SystemState) ReadOriginalState() {
	if IsFileExists(SystemStateFileName) {
		content, err := ioutil.ReadFile(SystemStateFileName)
		if err != nil {
			emitLine(logLevel.important, "failed read system state file %s, error: %s", SystemStateFileName, err.Error())
		} else {
			err := json.Unmarshal(content, &state)
			if err != nil {
				emitLine(logLevel.important, "failed converting content to Json from system state file %s. error: %s, Content: %s", SystemStateFileName, err.Error(), content)
			}
		}

		state.RemoveExpiredStates()
	}

	// debugJson(state)
}

func (state *SystemState) RemoveExpiredStates() {

	// max 30 days to keep state source
	expirationTime := time.Now().UTC().Add(time.Hour * 24 * 30 * (-1))
	expirationTimeNumber := DateToCustomLong(expirationTime)

	sourceList := make([]*SourceState, 0)

	for _, sourceState := range state.Sources {
		if sourceState.LastUpdatedTimeUtcNumber > expirationTimeNumber {
			sourceList = append(sourceList, sourceState)
		}
	}

	state.Sources = sourceList
}

func (state *SystemState) Restore() {
	CreateDirIfNotExist(".state", 0765)
	state.ReadOriginalState()
}

func (state *SystemState) Find(sourceId string) *SourceState {
	for _, sourceState := range state.Sources {
		if sourceId == sourceState.SourceId {
			return sourceState
		}
	}

	newSource := &SourceState{
		SourceId: sourceId,
		Offset:   0,
	}

	state.Sources = append(state.Sources, newSource)

	return newSource
}
