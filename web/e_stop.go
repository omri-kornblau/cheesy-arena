// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the team list.

package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Shows the team list.
func (web *Web) eStopHadnler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("e-stop: invalid alliance '%s'", alliance))
		return
	}

	// Default alliance is Red
	allianceLetter := "R"
	if alliance == "blue" {
		allianceLetter = "B"
	}

	driverStationIndex, err := strconv.Atoi(vars["driverStationIndex"])
	if err != nil {
		handleWebErr(w, err)
		return
	}

	if driverStationIndex < 1 || driverStationIndex > 3 {
		handleWebErr(w, fmt.Errorf("e-stop: invalid driverStationIndex '%d'", driverStationIndex))
		return
	}

	driverStationCode := fmt.Sprintf("%s%d", allianceLetter, driverStationIndex)

	log.Printf("Got e-stop from station %s", driverStationCode)
	web.arena.HandleEstop(driverStationCode, true)
}
