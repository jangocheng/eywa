package handlers

import (
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"io/ioutil"
	"net/http"
	"time"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, connections.Counts())
}

func ConnectionCount(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	Render.JSON(w, http.StatusOK, map[string]int{c.URLParams["channel_id"]: cm.Count()})
}

func ConnectionStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	devId := c.URLParams["device_id"]
	history := r.URL.Query().Get("history")

	status, err := models.FindConnectionStatus(ch, devId, history == "true")
	if err != nil {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	} else {
		Render.JSON(w, http.StatusOK, status)
	}
}

func SendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sender, ok := conn.(connections.Sender)
	if !ok {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": errors.New("connection is not allowed to send").Error()})
		return
	}

	err = sender.Send(bodyBytes)
	if err != nil {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
}

func RequestToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	timeout := Config().Connections.Websocket.Timeouts.Response.Duration
	var err error
	timeoutStr := r.URL.Query().Get("timeout")
	if len(timeoutStr) > 0 {
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	requester, ok := conn.(connections.Requester)
	if !ok {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": errors.New("connection is not allowed to request").Error()})
		return
	}

	msg, err := requester.Request(bodyBytes, timeout)
	if err != nil {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	w.Write(msg)
}