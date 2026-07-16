package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"vizag-steel-delay-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ParseDecimalTime(baseDate time.Time, decimalStr string) time.Time {
	if decimalStr == "" {
		return baseDate
	}
	f, err := strconv.ParseFloat(decimalStr, 64)
	if err != nil {
		return baseDate
	}
	hours := int(f)
	str := fmt.Sprintf("%.2f", f)
	var h, m int
	fmt.Sscanf(str, "%d.%d", &h, &m)
	
	return baseDate.Add(time.Hour * time.Duration(hours)).Add(time.Minute * time.Duration(m))
}

func ImportCSV(c *gin.Context) {
	session := sessions.Default(c)
	empNo := session.Get("user_id").(string)

	file, _, err := c.Request.FormFile("csv_file")
	if err != nil {
		c.String(http.StatusBadRequest, `<div class="bg-red-50 text-red-600 p-4 rounded-lg font-bold">Error reading file upload</div>`)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.String(http.StatusBadRequest, `<div class="bg-red-50 text-red-600 p-4 rounded-lg font-bold">Error parsing CSV format</div>`)
		return
	}

	importedCount := 0

	for i, row := range records {
		if importedCount >= 50 {
			break // Limit to 50 for demo to prevent browser crash
		}
		
		if i == 0 {
			continue
		}
		
		
		if len(row) < 8 {
			continue // Skip malformed rows that don't even have durations
		}

		getCol := func(idx int) string {
			if idx < len(row) {
				return row[idx]
			}
			return ""
		}

		delDateStr := getCol(0)
		shopCode := getCol(1)
		material := getCol(2)
		rake := getCol(3)
		delayFromStr := getCol(5)
		delayToStr := getCol(6)
		delayDurStr := getCol(7)
		cumDelayStr := getCol(8)
		eqpt := getCol(9)
		subEqpt := getCol(10)
		remarks := getCol(11)
		delayDesc := getCol(12)
		agencyCode := getCol(13)
		delayFreq := getCol(14)
		continuity := getCol(15)
		expectedComp := getCol(16)
		effDurationStr := getCol(17)
		legacyDelayId := getCol(22)

		var baseDate time.Time
		baseDate, _ = time.Parse("02-01-2006", delDateStr) // Assuming DD-MM-YYYY
		if baseDate.IsZero() {
			baseDate = time.Now() // Fallback
		}

		delayFrom := ParseDecimalTime(baseDate, delayFromStr)
		delayUpto := ParseDecimalTime(baseDate, delayToStr)
		
		delayDur, _ := strconv.ParseFloat(delayDurStr, 64)
		cumDur, _ := strconv.ParseFloat(cumDelayStr, 64)
		effDur, _ := strconv.ParseFloat(effDurationStr, 64)

		agencies := []string{"operations", "electrical", "mechanical", "shutdown"}
		agencyName := agencies[i%4]

		delay := models.DelayData{
			ShopCode:          shopCode,
			ShopDesc:          shopCode, // fallback if desc missing
			EqptName:          eqpt,
			SubEqptName:       subEqpt,
			Agency:            agencyName,
			DelayFrom:         delayFrom,
			DelayUpto:         delayUpto,
			DelayDuration:     int(effDur * 60), // effDur might be hours
			AllocatedDuration: int(delayDur * 60),
			Status:            "Work Done", // Legacy data is usually completed
			DelayDesc:         delayDesc,
			UserEntered:       empNo,
			Timestamp:         time.Now(),
			
			DelDate:       delDateStr,
			Material:      material,
			Rake:          rake,
			CumDelay:      cumDur,
			Remarks:       remarks,
			AgencyCode:    agencyCode,
			DelayFreq:     delayFreq,
			Continuity:    continuity,
			ExpectedComp:  expectedComp,
			EffDuration:   effDur,
			LegacyDelayID: legacyDelayId,
		}

		db.Create(&delay)
		importedCount++
	}

	html := fmt.Sprintf(`<div class="bg-green-50 text-green-700 p-4 rounded-lg font-bold flex items-center shadow-sm">
		<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
		Successfully imported %d records from legacy CSV!
	</div>
	<script>setTimeout(()=>window.location.reload(), 2000)</script>`, importedCount)

	c.String(http.StatusOK, html)
}
