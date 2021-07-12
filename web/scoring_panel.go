// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

type WheelSensorMessage struct {
	Color string `json:"color"`
	Rotation int `json:"rotation"`
	WaitedTowSeconds bool `json:"waited2"`
	WaitedFiveSeconds bool `json:"waited5"`
  }
// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	var realtimeScore **field.RealtimeScore
	var controlPanel *game.ControlPanel
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
		controlPanel = &web.arena.RedRealtimeScore.ControlPanel
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
		controlPanel = &web.arena.BlueRealtimeScore.ControlPanel
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, commandData, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		score := &(*realtimeScore).CurrentScore
		scoreChanged := false
		scoreChanged = true

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if number, err := strconv.Atoi(command); err == nil && number >= 1 && number <= 6 {
			// Handle per-robot scoring fields.
			if number <= 3 {
				index := number - 1
				score.ExitedInitiationLine[index] = !score.ExitedInitiationLine[index]
				scoreChanged = true
			} else {
				index := number - 4
				score.EndgameStatuses[index]++
				if score.EndgameStatuses[index] == 3 {
					score.EndgameStatuses[index] = 0
				}
				scoreChanged = true
			}
		} else {
			switch strings.ToUpper(command) {
			case "CI":
				// Don't read score from counter if not in match
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						if incrementGoal(score.AutoCellsInner[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
					if web.arena.MatchState == field.TeleopPeriod {
						if incrementGoal(score.TeleopCellsInner[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
				}

			case "CO":
				// Don't read score from counter if not in match
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						if incrementGoal(score.AutoCellsOuter[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
					if web.arena.MatchState == field.TeleopPeriod {
						if incrementGoal(score.TeleopCellsOuter[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
				}

			case "CL":
				// Don't read score from counter if not in match
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						if incrementGoal(score.AutoCellsBottom[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
					if web.arena.MatchState == field.TeleopPeriod {
						if incrementGoal(score.TeleopCellsBottom[:],
							score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
							scoreChanged = true
						}
					}
				}
			case "WL":
				if web.arena.MatchState == field.TeleopPeriod {
					var wheelSensorMessage WheelSensorMessage
					json.Unmarshal([]byte(commandData.(string)), &wheelSensorMessage)

					if wheelSensorMessage.Color == "red" {
						controlPanel.CurrentColor = game.ColorRed
					} else if wheelSensorMessage.Color == "green" {
						controlPanel.CurrentColor = game.ColorGreen
					} else if wheelSensorMessage.Color == "yellow" {
						controlPanel.CurrentColor = game.ColorYellow
					} else if wheelSensorMessage.Color == "blue" {
						controlPanel.CurrentColor = game.ColorBlue
					}

					if score.StageAtCapacity(game.Stage3, true) {
						fmt.Println(controlPanel.CurrentColor, controlPanel.GetPositionControlTargetColor())
						if controlPanel.CurrentColor == controlPanel.GetPositionControlTargetColor() && wheelSensorMessage.WaitedFiveSeconds {
							score.ControlPanelStatus = game.ControlPanelPosition
							scoreChanged = true
						}
					} else if score.StageAtCapacity(game.Stage2, true) {
						controlPanel.OrbitRotation += wheelSensorMessage.Rotation
						amountOfRotations := Abs(controlPanel.OrbitRotation)
						if amountOfRotations >= controlPanel.MaxOrbitRotation {
							controlPanel.MaxOrbitRotation = amountOfRotations
						} else if amountOfRotations + 4 <= controlPanel.MaxOrbitRotation {
							rotationsDiff := controlPanel.MaxOrbitRotation - amountOfRotations
							if controlPanel.OrbitRotation > 0 {
								controlPanel.OrbitRotation = 0 - rotationsDiff
							} else if controlPanel.OrbitRotation < 0 {
								controlPanel.OrbitRotation = rotationsDiff
							} else {
								controlPanel.OrbitRotation = 0
							}

							amountOfRotations = Abs(controlPanel.OrbitRotation)
							controlPanel.MaxOrbitRotation = amountOfRotations
						}

						if amountOfRotations >= 3 * 8 && amountOfRotations <= 5 * 8 && wheelSensorMessage.WaitedTowSeconds {
							score.ControlPanelStatus = game.ControlPanelRotation
							controlPanel.OrbitRotation = 0
							scoreChanged = true
						} else if amountOfRotations >= 5 * 8 {
							controlPanel.OrbitRotation = 0
						}

						fmt.Println(controlPanel.OrbitRotation, controlPanel.MaxOrbitRotation)
					}
				}

			case "Q":
				if decrementGoal(score.AutoCellsInner[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "A":
				if decrementGoal(score.AutoCellsOuter[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "Z":
				if decrementGoal(score.AutoCellsBottom[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "W":
				if incrementGoal(score.AutoCellsInner[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "S":
				if incrementGoal(score.AutoCellsOuter[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "X":
				if incrementGoal(score.AutoCellsBottom[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "E":
				if decrementGoal(score.TeleopCellsInner[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "D":
				if decrementGoal(score.TeleopCellsOuter[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "C":
				if decrementGoal(score.TeleopCellsBottom[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "R":
				if incrementGoal(score.TeleopCellsInner[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "F":
				if incrementGoal(score.TeleopCellsOuter[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "V":
				if incrementGoal(score.TeleopCellsBottom[:],
					score.CellCountingStage(web.arena.MatchState >= field.TeleopPeriod)) {
					scoreChanged = true
				}
			case "O":
				if score.ControlPanelStatus >= game.ControlPanelRotation {
					score.ControlPanelStatus = game.ControlPanelNone
				} else if score.StageAtCapacity(game.Stage2, true) {
					score.ControlPanelStatus = game.ControlPanelRotation
				}
				scoreChanged = true
			case "K":
				if score.ControlPanelStatus == game.ControlPanelRotation {
					controlPanel := &(*realtimeScore).ControlPanel
					controlPanel.CurrentColor++
					if controlPanel.CurrentColor == 5 {
						controlPanel.CurrentColor = 1
					}
					scoreChanged = true
				}
			case "P":
				if score.ControlPanelStatus == game.ControlPanelPosition {
					score.ControlPanelStatus = game.ControlPanelRotation
				} else if score.StageAtCapacity(game.Stage3, true) {
					score.ControlPanelStatus = game.ControlPanelPosition
				}
				scoreChanged = true
			case "L":
				score.RungIsLevel = !score.RungIsLevel
				scoreChanged = true
			}
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Increments the power cell count for the given goal, if the preconditions are met.
func incrementGoal(goal []int, currentStage game.Stage) bool {
	if int(currentStage) < len(goal) {
		goal[currentStage]++
		return true
	}
	return false
}

// Decrements the power cell count for the given goal, if the preconditions are met.
func decrementGoal(goal []int, currentStage game.Stage) bool {
	if int(currentStage) < len(goal) && goal[currentStage] > 0 {
		goal[currentStage]--
		return true
	}
	return false
}
