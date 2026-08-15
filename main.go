package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type LoginData struct {
	Success     bool
	Error       string
	Username    string
	IsSubmitted bool
}

type SystemStats struct {
	CPUUsage   int64
	TotalMem   int64
	UsedMem    int64
	AvailMem   int64
	RAMCap     int64
	Timestamp  string
	OSName     string
	OSVersion  string
	OSPlatform string
}

func fetchStats() SystemStats {

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		fmt.Println("Error fetching CPU stats:", err)
	}

	var cpuLoad int64 = 0

	if err == nil && len(cpuPercent) > 0 {
		cpuLoad = int64(cpuPercent[0])
	}

	vMem, err := mem.VirtualMemory()

	var totalMB, usedMB, availMB int64 = 0, 0, 0

	if err != nil {
		fmt.Println("Error fetching memory stats:", err)
	} else {
		totalMB = int64(vMem.Total / 1024 / 1024)
		usedMB = int64(vMem.Used / 1024 / 1024)
		availMB = totalMB - usedMB
	}

	hInfo, err := host.Info()
	osName := "Unknown OS"
	osVersion := "Unknown Version"
	osPlatform := "Unknown Platform"

	if err == nil {
		osName = hInfo.OS
		osVersion = hInfo.PlatformVersion
		osPlatform = hInfo.Platform
	}

	return SystemStats{
		CPUUsage:   cpuLoad,
		TotalMem:   totalMB,
		UsedMem:    usedMB,
		AvailMem:   availMB,
		RAMCap:     totalMB,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		OSName:     osName,
		OSVersion:  osVersion,
		OSPlatform: osPlatform,
	}
}

func main() {

	fmt.Println("Starting server on port 6767...")

	signInTmpl := template.Must(template.ParseFiles("signin.html"))
	indexTmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	})

	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {

		loginData := LoginData{}

		if r.Method == http.MethodPost {
			userInput := r.FormValue("username")
			passwordInput := r.FormValue("password")

			if userInput == "admin" && passwordInput == "pwd123456" {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			} else {
				loginData.IsSubmitted = true
				loginData.Error = "Invalid username or password"
				loginData.Success = false
			}
		}

		signInTmpl.Execute(w, loginData)

	})

	renderPage := func(w http.ResponseWriter, r *http.Request, blockName string, data interface{}) {
		if r.Header.Get("HX-Request") == "true" {
			indexTmpl.ExecuteTemplate(w, blockName, data)
		} else {
			indexTmpl.ExecuteTemplate(w, "base", map[string]interface{}{
				"ActiveBlock": blockName,
				"Data":        data,
			})
		}
	}

	//====================================================================================

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {

		renderPage(w, r, "dashboard", nil)

	})

	http.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
		stats := fetchStats()
		renderPage(w, r, "memory-details", stats)
	})

	http.HandleFunc("/api/live-stats", func(w http.ResponseWriter, r *http.Request) {
		stats := fetchStats()
		renderPage(w, r, "live-counters", stats)
	})

	http.HandleFunc("/system", func(w http.ResponseWriter, r *http.Request) {
		stats := fetchStats()
		renderPage(w, r, "system-info", stats)
	})

	if err := http.ListenAndServe(":6767", nil); err != nil {
		fmt.Println("Error starting server : ", err)
	}

}
