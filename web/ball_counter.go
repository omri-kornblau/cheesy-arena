// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/gorilla/mux"
)

func (web *Web) ballCountHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	color := vars["color"]
	if color != "red" && color != "blue" {
		handleWebErr(w, fmt.Errorf("ball counter: invalid color '%s'", color))
		return
	}

	level := vars["level"]
	if level != "lower" && level != "upper" {
		handleWebErr(w, fmt.Errorf("ball counter: invalid level '%s'", level))
		return
	}

	// Default color is blue
	realtimeScore := &web.arena.BlueRealtimeScore
	if color == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	}

	score := &(*realtimeScore).CurrentScore

	matchTimeSec := time.Since(web.arena.MatchStartTime).Seconds()
	if matchTimeSec >= float64(game.MatchTiming.WarmupDurationSec) &&
		matchTimeSec <= game.GetDurationToAutoCountingEnd().Seconds() {
		// Auto
		if level == "upper" {
			score.AutoCargoUpper[0]++
		} else {
			score.AutoCargoLower[0]++
		}
	} else if matchTimeSec <= game.GetDurationToTeleopCountingEnd().Seconds() {
		// Teleop
		if level == "upper" {
			score.TeleopCargoUpper[0]++
		} else {
			score.TeleopCargoLower[0]++
		}
	}

	web.arena.RealtimeScoreNotifier.Notify()
}

func (web *Web) ballCountErrorHadnler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	index := handleFormValue(r.PostFormValue("index"))
	color := handleFormValue(r.PostFormValue("color"))
	err := handleFormValue(r.PostFormValue("error"))

	handleWebErr(w, fmt.Errorf("ball counter: index '%s', color '%s', error '%s'", index, color, err))
}

func handleFormValue(value string) string {
	if value == "" {
		return "?"
	}

	return value
}
