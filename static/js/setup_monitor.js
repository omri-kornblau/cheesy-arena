// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Device Monitoring page.

var websocket;


const colors = new Map();
colors.set("on", "success");
colors.set("error", "danger");
colors.set("off", "primary");

// Handles a websocket message to update logs.
var handleLogs = function (data) {
  var columns = data.map(createDeviceColumnHTML).join("");
  $("#monitor").html(columns);
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/monitor/websocket", {
    devicesMonitoring: function (event) { handleLogs(event.data); }
  });
});

// Handles an element click and sends the appropriate websocket message.
function handleClick(name) {
  websocket.send("reseterror", name);
};


const createDeviceColumnHTML = (data) => {
  return `<div class="col-xs-4">
  <table class="table">
    <tr>
      <th colspan="2"><b class="btn btn-${colors.get(data.state)} button-${data.state}" onclick="handleClick('${data.name}')">${data.name}</b></th>
    </tr>
    <tr>
      <td>
        <div id="${data.name}" class="terminal">
          ${data.logs.reverse().map((value) => {
                return `<p class="log log-${value.level}">[${moment(value.timestamp).format("YYYY-MM-DD HH:mm:ss")}] ${value.message}</p>`;
          }).join("")}
        </div>
      </td>
    </tr>
  </table>
</div>`;
}