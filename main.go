package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Group struct {
	Name     string `json:"name"`
	GroupId  int    `json:"groupid"`
	Path     string `json:"path"`
	PGroupId int    `json:"pgroupid"`
	PName    string `json:"pname"`
	PPath    string `json:"ppath"`
}

type Groupdata struct {
	Status string  `json:"status"`
	Data   []Group `json:"data"`
	Err    string  `json:"err"`
	Msg    string  `json:"msg"`
}

type deviceResponse struct {
	Status string        `json:"status"`
	Data   []VehicleData `json:"data"`
	Err    string        `json:"err"`
	Msg    string        `json:"msg"`
}

type VehicleData struct {
	VehicleID       int             `json:"vehicleid"`
	VehicleNo       string          `json:"vehicleno"`
	Devices         []Device        `json:"devices"`
	GroupInfo       [][]interface{} `json:"groupInfo"`
	VehiclePrefData *string         `json:"vehicleprefdata"`
}

type Device struct {
	DeviceNo   string `json:"deviceno"`
	DeviceID   int    `json:"deviceid"`
	DeviceType string `json:"devicetype"`
	SimNo      string `json:"simno"`
	SimID      int    `json:"simid"`
	BindTag    string `json:"bindtag"`
}
type Modelinfo struct {
	ModelId      int    `json:"modelid"`
	Vehicletype  string `json:"vehicletype"`
	Oem          string `json:"oem"`
	Model        string `json:"model"`
	Variant      string `json:"variant"`
	Year         int    `json:"year"`
	Fueltype     string `json:"fueltype"`
	Transmission string `json:"transmission"`
}
type DeviceInfo struct {
	DeviceNo   string
	GroupId    int
	GroupNames []string
	Model      string
	ModelId    int
}
type Clientmodel struct {
	DeviceNo   string
	GroupId    int
	ModelId    int
	GroupNames string
	Model      string
}

// var packagemap map[string]*Packages
var logger *log.Logger

func main() {
	pid := os.Getpid()
	fmt.Printf("Process ID: %d\n", pid)
	logFile, err := os.OpenFile("./application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)

	logger.Println("Application started pid:", pid)

	// fmt.Println(token)
	start := time.Now()

	p := getPackages()

	elapsed := time.Since(start)
	fmt.Println("tolal packages to insert:", len(p), "time:", elapsed)
	start = time.Now()

	newp := InsertDb(p)
	// testmap := make(map[string][]*Clientmodel)
	// for _, val := range p {
	// 	key := val.GroupNames + val.Model
	// 	testmap[key] = append(testmap[key], val)
	// 	// fmt.Println(key, ",", val.GroupNames, ",", val.Model)
	// }
	// for _, val := range testmap {
	// 	fmt.Println("::::::::::::::", val[0].GroupNames, ",", val[0].Model)
	// }

	elapsed = time.Since(start)
	fmt.Println("Sucefully Updated :)........... time:", elapsed)
	fmt.Println("Calling webhooks.........")
	SendNotification(newp)
	fmt.Println("Finised !!!!!!!!!!!!!!!!!!!")
}

func getPackages() map[string]*Clientmodel {
	packagemap := make(map[string]*Clientmodel)
	deviceMap := make(map[string]*DeviceInfo)
	var mu sync.Mutex
	token := Gettoken()
	// fmt.Println(token)
	// token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaW5mbyI6eyJ1c2VyaWQiOjU1MzAsInVzZXJuYW1lIjoiZmxlZXQuYWRtaW4iLCJteXVzZXJpZCI6MTAzMDAsIm1haW51c2VybmFtZSI6InZhcmFzaGVia2FudGhpQGludGVsbGljYXIuaW4ifSwiaWF0IjoxNzM1ODEyNTQ3LCJleHAiOjE3Mzk0MTI1NDd9.PBfKpJ_J0gA4aIStgItCzJplJVSXvt2yOyOIfvrRQx8"
	groups, err := Getmygroups(token)
	if err != nil {
		log.Fatalln("Error fatal:", err.Error())
	}

	numWorkers := 63
	taskQueue := make(chan Group, len(groups))
	var wg sync.WaitGroup
	worker := func() {
		for group := range taskQueue {
			vdsdata, err := GetmyDevice(token, group.GroupId)
			if err != nil {
				log.Println("Error getting device data:", err.Error())
				wg.Done()
				continue
			}
			for _, val := range vdsdata {
				if val.VehicleNo != "" && len(val.Devices) > 0 && val.Devices[0].DeviceType == "laf" {
					// modeldata := getmodel(*val.VehiclePrefData)
					modeldata, modelid := getmodel(*val.VehiclePrefData)

					deviceNo := val.Devices[0].DeviceNo
					mu.Lock()
					if existingDevice, exists := deviceMap[deviceNo]; exists {
						if !contains(existingDevice.GroupNames, group.Name) {
							existingDevice.GroupNames = append(existingDevice.GroupNames, group.Name)
						}
					} else {
						deviceMap[deviceNo] = &DeviceInfo{
							DeviceNo:   val.Devices[0].DeviceNo,
							GroupId:    group.GroupId,
							GroupNames: []string{group.Name},
							Model:      modeldata,
							ModelId:    modelid,
						}
					}
					mu.Unlock()
				}
			}
			wg.Done()
		}
	}
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	for _, group := range groups {
		if group.PName == "fleet" && group.Name != "la5.ic" && group.Name != "LA5.IC.HARDWARE" && group.Name != "Demo_batt" {
			wg.Add(1)
			taskQueue <- group
		}
	}
	close(taskQueue)
	wg.Wait()
	fmt.Println("Getting Mqtt data...........")
	// odata := GetallmergeData()

	if err != nil {
		logger.Fatal("Error reading Optix data:", err)
	}

	fmt.Println("entering data to map.....")
	for deviceNo, device := range deviceMap {
		// if optixRecord, found := odata[deviceNo]; found {
		groupNames := strings.Join(device.GroupNames, " / ")
		packagemap[deviceNo] = &Clientmodel{
			DeviceNo:   device.DeviceNo,
			GroupId:    device.GroupId,
			ModelId:    device.ModelId,
			GroupNames: groupNames,
			Model:      device.Model,
			// SIM:               optixRecord.Sim,
			// HWversion:         optixRecord.HWVersion,
			// LAFFirmware:       optixRecord.LAFFirmware,
			// CANFirmware:       optixRecord.CANFirmware,
			// PLSign:            optixRecord.PLSign,
			// IOTSettingsSigned: optixRecord.IOTSettingsSigned,
			// }
		}
	}
	fmt.Println("completed Data Fetching Starting Insertion........")
	// for key, val := range packagemap {
	// 	fmt.Println(key, ",", val.GroupNames, ",", val.Model, ",", val.SIM, ",", val.HWversion, ",", val.LAFFirmware, ",", val.CANFirmware)
	// }
	return packagemap
}
func getmodel(data string) (string, int) {
	var vehiclePrefData Modelinfo

	err := json.Unmarshal([]byte(data), &vehiclePrefData)
	if err != nil {
		logger.Println("Error unmarshalling JSON:", err)
	}
	str := []string{vehiclePrefData.Vehicletype, vehiclePrefData.Oem, vehiclePrefData.Model, vehiclePrefData.Variant, strconv.Itoa(vehiclePrefData.Year), vehiclePrefData.Fueltype, vehiclePrefData.Transmission}
	return strings.Join(str, "_"), vehiclePrefData.ModelId
}

func Gettoken() string {
	postBody, _ := json.Marshal(map[string]map[string]string{
		"user": {
			"type":     "localuser",
			"username": "debug.admin",
			"password": "xyz321",
		}})

	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://apiplatform.intellicar.in/gettoken", "application/json", responseBody)

	if err != nil {
		logger.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalln(err)
	}
	tokenResp := struct {
		Status string `json:"status"`
		Data   struct {
			Token string `json:"token"`
		} `json:"data"`
		Userinfo struct {
			Userid   int    `json:"userid"`
			Typeid   int    `json:"typeid"`
			Username string `json:"username"`
		} `json:"userinfo"`
		Err string `json:"err"`
		Msg string `json:"msg"`
	}{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		logger.Fatal("Token parsing error")
	}

	Token := tokenResp.Data.Token

	// fmt.Println(Token)
	return Token
}

func Getmygroups(Token string) ([]Group, error) {
	postBody, _ := json.Marshal(map[string]string{
		"token": Token,
	})

	requestBody := bytes.NewBuffer(postBody)

	resp, err := http.Post("https://apiplatform.intellicar.in/api/user/getmygroups", "application/json", requestBody)
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var responseData Groupdata
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	if responseData.Status != "SUCCESS" {
		return nil, fmt.Errorf("API error: %v", responseData.Err)
	}

	return responseData.Data, nil
}

func GetmyDevice(Token string, Groupid int) ([]VehicleData, error) {

	PostBody, err := json.Marshal(map[string]string{
		"token":   Token,
		"groupid": fmt.Sprint(Groupid),
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	ResponseBody := bytes.NewBuffer(PostBody)

	Resp, err := http.Post("http://apiplatform.intellicar.in/api/vehicle/getmyvdsnew", "application/json", ResponseBody)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while making the POST request: %v", err)
	}
	defer Resp.Body.Close()

	if Resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", Resp.Status)
	}

	body, err := ioutil.ReadAll(Resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	var response deviceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Println("error1", err)
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	if response.Status != "SUCCESS" {
		logger.Println("error2")
		return nil, fmt.Errorf("API error: %v", response.Err)
	}
	return response.Data, nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
