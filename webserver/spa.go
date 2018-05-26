package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ghetzel/canibus/api"
	"github.com/ghetzel/canibus/candevice"
	"github.com/ghetzel/canibus/core"
	"github.com/ghetzel/canibus/hacksession"
	"github.com/ghetzel/canibus/logger"
	"github.com/gorilla/mux"
)

type CanDeviceJSON struct {
	Id          int
	DeviceType  string
	DeviceDesc  string
	HackSession string
	Year        string
	Make        string
	Model       string
}

type ConfigJSSON struct {
	Id         int
	DeviceType string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("index-spa.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", p.Body)
}

func partialLobbyHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log("Lobby checking auth...")
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	p, err := loadPage("partials/lobby.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", p.Body)
}

func haxTransmitHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log("Transmit Packet")
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	hax := dev.GetHackSession()
	if hax == nil {
		http.Error(w, "Session not configured", http.StatusNotFound)
		return
	}
	if !hax.IsActiveUser(user) {
		http.Error(w, "You are not a part of this hacksession", http.StatusNotFound)
		return
	}
	jsonTx := r.FormValue("tx")
	var TxPkts []api.TransmitPacket
	jerr := json.Unmarshal([]byte(jsonTx), &TxPkts)
	if jerr != nil {
		logger.Log("Transmit unmarshal error on: " + jsonTx)
		http.Error(w, jerr.Error(), http.StatusNotFound)
		return
	}
	for i := range TxPkts {
		inject_err := hax.InjectPacket(user, TxPkts[i])
		if inject_err != nil {
			logger.Log("Transmit packet error: " + inject_err.Error())
			http.Error(w, inject_err.Error(), http.StatusNotFound)
			return
		}
	}
	fmt.Fprintf(w, "%s", "OK")
}

func haxStopHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log("Stop Sniffer")
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	hax := dev.GetHackSession()
	if hax == nil {
		http.Error(w, "Session not configured", http.StatusNotFound)
		return
	}
	if !hax.IsActiveUser(user) {
		http.Error(w, "You are not a part of this hacksession", http.StatusNotFound)
		return
	}
	dev.StopSniffing()
	fmt.Fprintf(w, "%s", "OK")
}

func haxStartHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log("Start Sniffer")
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	hax := dev.GetHackSession()
	if hax == nil {
		http.Error(w, "Session not configured", http.StatusNotFound)
		return
	}
	if !hax.IsActiveUser(user) {
		/* No reason for this, just add them
		http.Error(w, "You are not a part of this hacksession", http.StatusNotFound)
		return
		*/
		hax.AddUser(user)
	}
	if hax.GetStateValue() != hacksession.STATE_SNIFF {
		hax.SetState(hacksession.STATE_SNIFF)
		dev.StartSniffing()
	}
	fmt.Fprintf(w, "%s", "OK")
}

func haxPacketsHandler(w http.ResponseWriter, r *http.Request) {
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	hax := dev.GetHackSession()
	if hax == nil {
		http.Error(w, "Session not configured", http.StatusNotFound)
		return
	}
	if !hax.IsActiveUser(user) {
		http.Error(w, "You are not a part of this hacksession", http.StatusNotFound)
		return
	}
	packets := hax.GetPackets(user)

	j, err := json.Marshal(packets)
	if err != nil {
		logger.Log("Could not convert can packets to json")
		return
	}
	fmt.Fprintf(w, "%s", j)
}

func configCanHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log("Config CAN Device, checking auth...")
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	if dev.GetHackSession() == nil {
		hacks := hacksession.HackSession{}
		hacks.SetState(hacksession.STATE_CONFIG)
		hacks.SetDevice(dev)
		user.SetDeviceId(dev.GetId())
		dev.SetHackSession(&hacks)
		hacks.AddUser(user)
	}

	p, err := loadPage("partials/config.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", p.Body)
}

func joinHaxHandler(w http.ResponseWriter, r *http.Request) {
	auth_err := checkAuth(w, r)
	if auth_err != nil {
		return
	}
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	session, _ := store.Get(r, "canibus")
	userName := session.Values["user"].(string)
	user, _ := core.GetUserByName(userName)

	hax := dev.GetHackSession()
	hax.AddUser(user)
	var p *Page
	var err error
	if hax.GetStateValue() == hacksession.STATE_SNIFF {
		p, err = loadPage("partials/sniff.html")
	} else {
		p, err = loadPage("partials/config.html")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", p.Body)
}

func addSimHandler(w http.ResponseWriter, r *http.Request) {
	config := core.GetConfig()
	dev := &candevice.Simulator{}
	newId := config.AppendDriver(dev)
	data := CanDeviceJSON{}
	data.Id = newId
	data.DeviceType = dev.DeviceType()
	data.HackSession = "Idle"
	j, err := json.Marshal(data)
	if err != nil {
		logger.Log("Could not convert candevices to json")
		return
	}
	fmt.Fprintf(w, "%s", j)
}

func candeviceInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	canId, canId_err := strconv.Atoi(vars["id"])
	if canId_err != nil {
		http.Error(w, canId_err.Error(), http.StatusNotFound)
		return
	}
	dev, dev_err := core.GetDeviceById(canId)
	if dev_err != nil {
		http.Error(w, dev_err.Error(), http.StatusNotFound)
		return
	}
	data := CanDeviceJSON{}
	data.Id = dev.GetId()
	data.DeviceType = dev.DeviceType()
	data.DeviceDesc = dev.DeviceDesc()
	j, err := json.Marshal(data)
	if err != nil {
		logger.Log("Could not convert candevices to json")
		return
	}
	fmt.Fprintf(w, "%s", j)
}

func candevicesHandler(w http.ResponseWriter, r *http.Request) {
	config := core.GetConfig()
	drivers := config.GetDrivers()
	var data []CanDeviceJSON
	for i := range drivers {
		dev := CanDeviceJSON{}
		dev.Id = drivers[i].GetId()
		dev.DeviceType = drivers[i].DeviceType()
		dev.DeviceDesc = drivers[i].DeviceDesc()
		hax := drivers[i].GetHackSession()
		if hax == nil {
			dev.HackSession = "Idle"
		} else {
			dev.HackSession = hax.GetState()
		}
		dev.Year = drivers[i].GetYear()
		dev.Make = drivers[i].GetMake()
		dev.Model = drivers[i].GetModel()
		data = append(data, dev)
	}
	j, err := json.Marshal(data)
	if err != nil {
		logger.Log("Could not convert candevices to json")
		return
	}
	fmt.Fprintf(w, "%s", j)
}

func StartSPAWebListener(root string, ip string, port string) error {
	web_root = root
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/logout", logoutHandler)
	r.HandleFunc("/partials/lobby.html", partialLobbyHandler)
	r.HandleFunc("/candevice/{id}/config", configCanHandler)
	r.HandleFunc("/candevice/{id}/join", joinHaxHandler)
	r.HandleFunc("/candevice/{id}/info", candeviceInfoHandler)
	r.HandleFunc("/hax/{id}/packets", haxPacketsHandler)
	r.HandleFunc("/hax/{id}/start", haxStartHandler)
	r.HandleFunc("/hax/{id}/stop", haxStopHandler)
	r.HandleFunc("/hax/{id}/transmit", haxTransmitHandler)
	r.HandleFunc("/candevices", candevicesHandler)
	r.HandleFunc("/lobby/AddSimulator", addSimHandler)

	http.Handle("/partials/", http.FileServer(FS(false)))
	http.Handle("/js/", http.FileServer(FS(false)))
	http.Handle("/css/", http.FileServer(FS(false)))
	http.Handle("/fonts/", http.FileServer(FS(false)))
	http.Handle("/images/", http.FileServer(FS(false)))
	http.Handle("/bootstrap/", http.FileServer(FS(false)))
	http.Handle("/", r)
	remote := ip + ":" + port
	logger.Log("Starting CANiBUS Web server on " + remote)
	err := http.ListenAndServe(remote, nil)
	if err != nil {
		return logger.Err("Could not bind web to port: " + err.Error())
	}
	return nil
}
