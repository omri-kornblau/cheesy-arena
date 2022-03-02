// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

const (
	hubAutoGracePeriodSec   = 5
	HubTeleopGracePeriodSec = 10
)

var MatchTiming = struct {
	WarmupDurationSec           int
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TimeoutDurationSec          int
}{0, 15, 2, 135, 30, 0}

var AutoCountingOffset = 5
var TeleopCountingOffset = 10

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec) *
		time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec) * time.Second
}

func GetDurationToAutoCountingEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+AutoCountingOffset) *
		time.Second
}

func GetDurationToTeleopCountingEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
		MatchTiming.TeleopDurationSec+TeleopCountingOffset) * time.Second
}
