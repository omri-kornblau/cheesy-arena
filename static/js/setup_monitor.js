// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;
let raspberryPiLogs = new Array();
let pc1Logs = [];
let pc2Logs = [];

const terminals = new Map();
terminals.set("raspberypi", raspberryPiLogs);
terminals.set("pc1", pc1Logs);
terminals.set("pc2", pc2Logs);

const colors = new Map();
colors.set("on", "green");
colors.set("error", "orange");
colors.set("off", "red");

// Handles a websocket message to update logs.
var handleLogs = function(data) {
  var name = data.deviceName;
  var state = data.state;
  var color = colors.get(state);
  var logs = data.logs;

  logs.forEach(log => {
    terminals.get(name).push(log + "\n");
  });
  $("#" + name + "-button").css("background-color", color);
  $("#" + name).text(terminals.get(name));
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/monitor/websocket", {
    logs: function(event) { handleLogs(event.data); }
  });
});

// Handles an element click and sends the appropriate websocket message.
function handleClick(name) {
  // websocket.send(name + "-reset");
  $("#" + name + "-button").css("background-color", "green");
  $("#" + name).text("recovered");
  clear(name);
};

function clear(name){
  if(name === "raspberrypi"){
    raspberryPiLogs = [];
    return;
  }
  if(name === "pc1"){
    pc1Logs = [];
    return;
  }
  if(name === "pc2"){
    pc2Logs = [];
    return;
  }
}

function test() {
  var data = {
    "deviceName": "pc1",
    "state": "error",
    "logs": ["1","2","3"],
  }
  handleLogs(data);
}