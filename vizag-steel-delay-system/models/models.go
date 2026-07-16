package models

import (
	"fmt"
	"html/template"
	"time"
)

type User struct {
	EmpNo       string `gorm:"uniqueIndex"`
	Password    string
	EmpName     string
	Dept        string
	Designation string
	Role        string
	Active      bool
}

type EqptMaster struct {
	ShopCode    string
	ShopDesc    string
	EqptCode    string
	SubEqptCode string
}

type DelayData struct {
	ID                uint      `gorm:"primaryKey"`
	ShopCode          string
	ShopDesc          string
	EqptName          string
	SubEqptName       string
	Agency            string
	DelayFrom         time.Time
	DelayUpto         time.Time // Set when marked done
	DelayDuration     int       // Actual total duration calculated when done
	AllocatedDuration int       // The expected duration
	Status            string    // "Pending", "In Progress", "Work Done"
	WorkStartedAt     *time.Time
	DelayDesc         string
	UserEntered       string
	Timestamp         time.Time
	
	DelDate       string
	Material      string
	Rake          string
	CumDelay      float64
	Remarks       string
	AgencyCode    string
	DelayFreq     string
	Continuity    string
	ExpectedComp  string
	EffDuration   float64
	UserModified  string
	TmstpModified string
	LegacyDelayID string
}

func (d *DelayData) DurationFormatted() template.HTML {
	if d.Status == "In Progress" {
		return template.HTML(`<span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-bold bg-yellow-100 text-yellow-800 animate-pulse"><svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>In Progress</span>`)
	} else if d.Status == "Pending" {
		return template.HTML(`<span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-bold bg-gray-100 text-gray-800"><svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>Pending</span>`)
	}
	hours := d.DelayDuration / 60
	mins := d.DelayDuration % 60
	if hours > 0 {
		return template.HTML(fmt.Sprintf("%d<span class=\"text-xs text-gray-500 font-normal mx-1\">h</span>%d<span class=\"text-xs text-gray-500 font-normal ml-1\">m</span>", hours, mins))
	}
	return template.HTML(fmt.Sprintf("%d<span class=\"text-xs text-gray-500 font-normal ml-1\">m</span>", mins))
}
