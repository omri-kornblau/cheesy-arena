// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package web

import (
	"fmt"
	"net/http"

	"github.com/Team254/cheesy-arena/field"
	"github.com/gorilla/mux"
)

// Shows the team list.
func (web *Web) ballCountHadnler(w http.ResponseWriter, r *http.Request) {
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
	if level != "high" && level != "low" {
		handleWebErr(w, fmt.Errorf("ball counter: invalid level '%s'", level))
		return
	}

	// Don't read score from counter if not in match
	if web.arena.MatchState == field.PostMatch || web.arena.MatchState == field.PreMatch {
		return
	}

	// Default color is blue
	realtimeScore := &web.arena.BlueRealtimeScore
	if color == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	}

	score := &(*realtimeScore).CurrentScore

	if web.arena.MatchState == field.AutoPeriod {
		if level == "high" {
			score.AutoCargoUpper[0]++
		} else {
			score.AutoCargoLower[0]++
		}
	} else {
		if level == "high" {
			score.TeleopCargoUpper[0]++
		} else {
			score.TeleopCargoLower[0]++
		}
	}

	web.arena.RealtimeScoreNotifier.Notify()
}
