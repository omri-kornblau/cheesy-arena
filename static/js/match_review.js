// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for editing a match in the match review page.

var scoreTemplate = Handlebars.compile($("#scoreTemplate").html());
var allianceResults = {};
var matchResult;

// Hijack the form submission to inject the data in JSON form so that it's easier for the server to parse.
$("form").submit(function() {
  updateResults("red");
  updateResults("blue");

  matchResult.RedScore = allianceResults["red"].score;
  matchResult.BlueScore = allianceResults["blue"].score;
  matchResult.RedCards = allianceResults["red"].cards;
  matchResult.BlueCards = allianceResults["blue"].cards;
  var matchResultJson = JSON.stringify(matchResult);

  // Inject the JSON data into the form as hidden inputs.
  $("<input />").attr("type", "hidden").attr("name", "matchResultJson").attr("value", matchResultJson).appendTo("form");

  return true;
});

// Draws the match-editing form for one alliance based on the cached result data.
var renderResults = function(alliance) {
  var result = allianceResults[alliance];
  var scoreContent = scoreTemplate(result);
  $("#" + alliance + "Score").html(scoreContent);

  // Set the values of the form fields from the JSON results data.
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;

    if (i < 1) {
      getInputElement(alliance, "AutoCargoUpper").val(result.score.AutoCargoUpper);
      getInputElement(alliance, "AutoCargoLower").val(result.score.AutoCargoUpper);

      getInputElement(alliance, "TeleopCargoUpper").val(result.score.TeleopCargoUpper);
      getInputElement(alliance, "TeleopCargoLower").val(result.score.TeleopCargoLower);
    }

    getInputElement(alliance, "ExitedTarmac" + i1).prop("checked", result.score.ExitedTarmac[i]);
    getInputElement(alliance, "EndgameStatuses" + i1, result.score.EndgameStatuses[i]).prop("checked", true);
  }

  if (result.score.Fouls != null) {
    $.each(result.score.Fouls, function(k, v) {
      getInputElement(alliance, "Foul" + k + "Team", v.TeamId).prop("checked", true);
      getSelectElement(alliance, "Foul" + k + "RuleId").val(v.RuleId);
      getInputElement(alliance, "Foul" + k + "Time").val(v.TimeInMatchSec);
    });
  }

  if (result.cards != null) {
    $.each(result.cards, function(k, v) {
      getInputElement(alliance, "Team" + k + "Card", v).prop("checked", true);
    });
  }
};

// Converts the current form values back into JSON structures and caches them.
var updateResults = function(alliance) {
  var result = allianceResults[alliance];
  var formData = {};
  $.each($("form").serializeArray(), function(k, v) {
    formData[v.name] = v.value;
  });

  result.score.ExitedTarmac = [];
  result.score.AutoCargoUpper = 0;
  result.score.AutoCargoLower = 0;
  result.score.TeleopCargoUpper = 0;
  result.score.TeleopCargoLower = 0;
  result.score.EndgameStatuses = [];
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;

    if (i < 1) {
      result.score.AutoCargoUpper = parseInt(formData[alliance + "AutoCargoUpper"]);
      result.score.AutoCargoLower = parseInt(formData[alliance + "AutoCargoLower"]);
    }

    result.score.ExitedTarmac[i] = formData[alliance + "ExitedTarmac" + i1] === "on";
    result.score.EndgameStatuses[i] = parseInt(formData[alliance + "EndgameStatuses" + i1]);

      result.score.TeleopCargoUpper = parseInt(formData[alliance + "TeleopCargoUpper"]);
      result.score.TeleopCargoLower = parseInt(formData[alliance + "TeleopCargoLower"]);
  }

  result.score.Fouls = [];
  for (var i = 0; formData[alliance + "Foul" + i + "Time"]; i++) {
    var prefix = alliance + "Foul" + i;
    var foul = {TeamId: parseInt(formData[prefix + "Team"]), RuleId: parseInt(formData[prefix + "RuleId"]),
                TimeInMatchSec: parseFloat(formData[prefix + "Time"])};
    result.score.Fouls.push(foul);
  }

  result.cards = {};
  $.each([result.team1, result.team2, result.team3], function(i, team) {
    result.cards[team] = formData[alliance + "Team" + team + "Card"];
  });
};

// Appends a blank foul to the end of the list.
var addFoul = function(alliance) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.push({TeamId: 0, Rule: "", TimeInMatchSec: 0});
  renderResults(alliance);
};

// Removes the given foul from the list.
var deleteFoul = function(alliance, index) {
  updateResults(alliance);
  var result = allianceResults[alliance];
  result.score.Fouls.splice(index, 1);
  renderResults(alliance);
};

// Returns the form input element having the given parameters.
var getInputElement = function(alliance, name, value) {
  var selector = "input[name=" + alliance + name + "]";
  if (value !== undefined) {
    selector += "[value=" + value + "]";
  }
  return $(selector);
};

// Returns the form select element having the given parameters.
var getSelectElement = function(alliance, name) {
  var selector = "select[name=" + alliance + name + "]";
  return $(selector);
};
