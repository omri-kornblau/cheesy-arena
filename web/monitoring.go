package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (web *Web) deviceErrorPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	if err := r.ParseForm(); err != nil {
		// Log here?
	}

	deviceName := r.FormValue("deviceName")
	errorString := r.FormValue("error")

	if deviceName == "" {
		return
	}

	// Send any other var as an extra data
	extraData := []string{}
	for key, values := range r.PostForm {
		if key == "deviceName" || key == "error" {
			continue
		}

		extraData = append(extraData, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
	}

	web.arena.DevicesMonitor.SetDeviceError(deviceName, errorString, extraData...)
}

func (web *Web) deviceHealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	deviceName := vars["deviceName"]

	if deviceName == "" {
		return
	}

	web.arena.DevicesMonitor.SetDeviceSeen(deviceName)
}
