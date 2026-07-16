package handlers

import (
	"fmt"
	"net/http"
	"time"

	"vizag-steel-delay-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ShowEntry(c *gin.Context) {
	session := sessions.Default(c)
	empName := session.Get("emp_name")
	role := session.Get("role")

	var shops []string
	db.Model(&models.EqptMaster{}).Distinct("shop_desc").Pluck("shop_desc", &shops)

	var recentDelays []models.DelayData
	db.Order("timestamp desc").Limit(10).Find(&recentDelays)

	c.HTML(http.StatusOK, "entry.html", gin.H{
		"EmpName":      empName,
		"Role":         role,
		"Shops":        shops,
		"RecentDelays": recentDelays,
	})
}

func GetEquipment(c *gin.Context) {
	shopDesc := c.Query("shop_desc")
	var equipment []string
	db.Model(&models.EqptMaster{}).Where("shop_desc = ?", shopDesc).Distinct("eqpt_code").Pluck("eqpt_code", &equipment)

	html := "<option value=''>Select Equipment</option>"
	for _, eq := range equipment {
		html += fmt.Sprintf("<option value='%s'>%s</option>", eq, eq)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func GetSubEquipment(c *gin.Context) {
	shopDesc := c.Query("shop_desc")
	eqptName := c.Query("eqpt_name")
	
	var subEquipment []string
	db.Model(&models.EqptMaster{}).Where("shop_desc = ? AND eqpt_code = ?", shopDesc, eqptName).Distinct("sub_eqpt_code").Pluck("sub_eqpt_code", &subEquipment)

	html := "<option value=''>Select Sub Equipment</option>"
	for _, sub := range subEquipment {
		html += fmt.Sprintf("<option value='%s'>%s</option>", sub, sub)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}


func SubmitDelay(c *gin.Context) {
	session := sessions.Default(c)
	empNo := session.Get("user_id").(string)

	fromDate := c.PostForm("from_date")
	fromTime := c.PostForm("from_time")
	uptoDate := c.PostForm("upto_date")
	uptoTime := c.PostForm("upto_time")
	delayDurationStr := c.PostForm("delay_duration_val")
	shopDesc := c.PostForm("shop_desc")
	eqptName := c.PostForm("eqpt_name")

	if fromDate == "" || fromTime == "" || uptoDate == "" || uptoTime == "" || delayDurationStr == "" || shopDesc == "" || eqptName == "" {
		c.String(http.StatusBadRequest, "<tr class='bg-red-50'><td colspan='5' class='px-6 py-4 text-red-600 font-bold'>Error: Missing required fields. Please fill out all required fields.</td></tr>")
		return
	}

	delayFrom, _ := time.Parse("2006-01-02T15:04", fromDate+"T"+fromTime)
	delayUpto, _ := time.Parse("2006-01-02T15:04", uptoDate+"T"+uptoTime)
	
	var delayDuration int
	fmt.Sscanf(delayDurationStr, "%d", &delayDuration)

	delay := models.DelayData{
		ShopCode:          shopDesc,
		ShopDesc:          shopDesc,
		EqptName:          eqptName,
		SubEqptName:       c.PostForm("sub_eqpt_name"),
		Agency:            c.PostForm("agency"),
		DelayFrom:         delayFrom,
		DelayUpto:         delayUpto,
		DelayDuration:     delayDuration,
		AllocatedDuration: delayDuration,
		Status:            "Work Done",
		DelayDesc:         c.PostForm("delay_desc"),
		UserEntered:       empNo,
		Timestamp:         time.Now(),
	}

	db.Create(&delay)


	c.Redirect(http.StatusSeeOther, "/entry")
}

func StartWork(c *gin.Context) {
	id := c.Param("id")
	var delay models.DelayData
	if err := db.First(&delay, id).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	if delay.Status == "Pending" {
		delay.Status = "In Progress"
		now := time.Now()
		delay.WorkStartedAt = &now
		db.Save(&delay)
	}

	c.String(http.StatusOK, `<span class="px-4 py-2 bg-gray-200 text-gray-700 text-xs font-bold rounded shadow">Working...</span><script>setTimeout(()=>window.location.reload(), 500)</script>`)
}

func MarkDone(c *gin.Context) {
	id := c.Param("id")
	var delay models.DelayData
	if err := db.First(&delay, id).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	if delay.Status == "In Progress" {
		delay.Status = "Work Done"
		delay.DelayUpto = time.Now()
		delay.DelayDuration = int(delay.DelayUpto.Sub(delay.DelayFrom).Minutes())
		if delay.DelayDuration < 0 {
			delay.DelayDuration = 0
		}
		db.Save(&delay)
	}

	c.String(http.StatusOK, `<span class="px-4 py-2 bg-gray-200 text-gray-700 text-xs font-bold rounded shadow">Done!</span><script>setTimeout(()=>window.location.reload(), 500)</script>`)
}
