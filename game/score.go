// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	ExitedTarmac     [3]bool
	AutoCargoLower   int
	AutoCargoUpper   int
	TeleopCargoLower int
	TeleopCargoUpper int
	EndgameStatuses  [3]EndgameStatus
	Fouls            []Foul
	ElimDq           bool
}

type ScoreSummary struct {
	ExitedTarmacPoints            int
	AutoCargoPoints               int
	TeleopCargoPoints             int
	CargoPoints                   int
	CargoLeftForCargoRankingPoint int
	LowerCargo                    int
	UpperCargo                    int
	AutoPoints                    int
	TeleopPoints                  int
	EndgamePoints                 int
	FoulPoints                    int
	Score                         int
	CargoRankingPoint             bool
	EndgameRankingPoint           bool
}

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameLowRung
	EndgameMidRung
	EndgameHighRung
	EndgameTraversalRung
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul, teleopStarted bool) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, exited := range score.ExitedTarmac {
		if exited {
			summary.ExitedTarmacPoints += 2
		}
	}

	// Calculate cargo points.
	summary.AutoCargoPoints += score.AutoCargoLower * 2

	summary.AutoCargoPoints = score.AutoCargoUpper * 4

	summary.TeleopCargoPoints = score.TeleopCargoUpper * 2

	summary.TeleopCargoPoints += score.TeleopCargoLower * 1

	summary.CargoPoints = summary.AutoCargoPoints + summary.TeleopCargoPoints

	// Calculate auto and teleop points.
	summary.AutoPoints = summary.ExitedTarmacPoints +
		summary.AutoCargoPoints

	summary.TeleopPoints = summary.ExitedTarmacPoints +
		summary.TeleopCargoPoints

	// Calculate cargo bonus RP.
	cargoRPThreshold := 20

	if score.AutoCargoLower+score.AutoCargoUpper >= 5 {
		cargoRPThreshold = 18
	}

	summary.LowerCargo = score.AutoCargoLower + score.TeleopCargoLower
	summary.UpperCargo = score.AutoCargoUpper + score.TeleopCargoUpper
	summary.CargoRankingPoint = summary.UpperCargo+summary.LowerCargo >= cargoRPThreshold

	summary.CargoLeftForCargoRankingPoint = cargoRPThreshold - cargoRPThreshold
	if summary.CargoLeftForCargoRankingPoint < 0 {
		summary.CargoLeftForCargoRankingPoint = 0
	}

	// Calculate endgame points.
	for _, endgameStatus := range score.EndgameStatuses {
		switch endgameStatus {
		case EndgameNone:
			continue
		case EndgameLowRung:
			summary.EndgamePoints += 4
		case EndgameMidRung:
			summary.EndgamePoints += 6
		case EndgameHighRung:
			summary.EndgamePoints += 10
		case EndgameTraversalRung:
			summary.EndgamePoints += 15
		}
	}

	// Calculate endgame bonus RP.
	const EndgamePointsForRP = 16
	summary.EndgameRankingPoint = summary.EndgamePoints >= EndgamePointsForRP

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	summary.Score = summary.AutoPoints +
		summary.TeleopCargoPoints +
		summary.EndgamePoints +
		summary.FoulPoints

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.ExitedTarmac != other.ExitedTarmac ||
		score.AutoCargoLower != other.AutoCargoLower ||
		score.AutoCargoUpper != other.AutoCargoUpper ||
		score.TeleopCargoLower != other.TeleopCargoLower ||
		score.TeleopCargoUpper != other.TeleopCargoUpper ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.ElimDq != other.ElimDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
