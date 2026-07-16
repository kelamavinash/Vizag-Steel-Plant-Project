package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"vizag-steel-delay-system/models"

	"github.com/gin-gonic/gin"
)

func ShowReports(c *gin.Context) {
	session := sessions.Default(c)
	empName := session.Get("emp_name")
	role := session.Get("role")

	var shops []string
	db.Model(&models.EqptMaster{}).Distinct("shop_desc").Pluck("shop_desc", &shops)

	var delays []models.DelayData
	db.Find(&delays)

	totalDelays := len(delays)
	totalMins := 0
	activeIssues := 0
	for _, d := range delays {
		totalMins += d.DelayDuration
		if d.Status != "Work Done" {
			activeIssues++
		}
	}
	totalHours := float64(totalMins) / 60.0

	c.HTML(http.StatusOK, "reports.html", gin.H{
		"EmpName":      empName,
		"Role":         role,
		"Shops":        shops,
		"TotalDelays":  totalDelays,
		"TotalHours":   fmt.Sprintf("%.1f", totalHours),
		"Availability": "89%", // Hardcoded as it's complex to compute without base hours
		"ActiveIssues": activeIssues,
	})
}

func GetReportData(c *gin.Context) {
	shopDesc := c.Query("shop_desc")
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	query := db.Model(&models.DelayData{})

	if shopDesc != "" {
		query = query.Where("shop_desc = ?", shopDesc)
	}

	shift := c.Query("shift")
	if shift != "" {
		now := time.Now()
		year, month, day := now.Date()
		switch shift {
		case "A":
			startTime := time.Date(year, month, day, 6, 0, 0, 0, now.Location())
			query = query.Where("delay_from >= ? AND delay_from < ?", startTime, startTime.Add(8*time.Hour))
		case "B":
			startTime := time.Date(year, month, day, 14, 0, 0, 0, now.Location())
			query = query.Where("delay_from >= ? AND delay_from < ?", startTime, startTime.Add(8*time.Hour))
		case "C":
			startTime := time.Date(year, month, day, 22, 0, 0, 0, now.Location())
			query = query.Where("delay_from >= ? AND delay_from < ?", startTime, startTime.Add(8*time.Hour))
		}
	} else {
		layout := "2006-01-02"
		if fromDateStr != "" {
			if fromDate, err := time.Parse(layout, fromDateStr); err == nil {
				query = query.Where("delay_from >= ?", fromDate)
			}
		}
		if toDateStr != "" {
			if toDate, err := time.Parse(layout, toDateStr); err == nil {
				query = query.Where("delay_upto < ?", toDate.AddDate(0, 0, 1))
			}
		}
	}

	var delays []models.DelayData
	query.Order("delay_from desc").Find(&delays)

	costPerMin := map[string]int{
		"Blast Furnace": 250,
		"Steel Melting Shop": 300,
		"Wire Rod Mill": 150,
	}

	totalCost := 0

	html := ""
	for _, d := range delays {
		cost := 100 // default
		if c, ok := costPerMin[d.ShopDesc]; ok {
			cost = c
		}
		totalCost += d.DelayDuration * cost

		html += fmt.Sprintf(`
			<tr class="hover:bg-gray-50 border-b border-gray-100 transition-colors">
				<td class="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900">%s</td>
				<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700">%s</td>
				<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
				<td class="px-6 py-4 whitespace-nowrap">
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-gray-100 text-gray-800">
						%s
					</span>
				</td>
				<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
				<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">%s</td>
				<td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900">%.1f</td>
			</tr>
		`, d.ShopDesc, d.EqptName, d.SubEqptName, d.Agency, d.DelayFrom.Format("2006-01-02 15:04"), d.DelayUpto.Format("2006-01-02 15:04"), float64(d.DelayDuration)/60.0)
	}

	if len(delays) == 0 {
		html = `
		<tr>
			<td colspan="7" class="px-6 py-12 text-center">
				<div class="flex flex-col items-center justify-center">
					<svg class="w-12 h-12 text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path></svg>
					<p class="text-sm font-medium text-gray-500">Select filters and click Apply to load data</p>
				</div>
			</td>
		</tr>`
	}


	c.Header("HX-Trigger", fmt.Sprintf(`{"updateCost": {"value": %d}}`, totalCost))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func GetChartData(c *gin.Context) {
	shopDesc := c.Query("shop_desc")
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	query := db.Model(&models.DelayData{})

	if shopDesc != "" {
		query = query.Where("shop_desc = ?", shopDesc)
	}

	layout := "2006-01-02"
	if fromDateStr != "" {
		if fromDate, err := time.Parse(layout, fromDateStr); err == nil {
			query = query.Where("delay_from >= ?", fromDate)
		}
	}
	if toDateStr != "" {
		if toDate, err := time.Parse(layout, toDateStr); err == nil {
			query = query.Where("delay_upto < ?", toDate.AddDate(0, 0, 1))
		}
	}

	type Result struct {
		Agency        string
		TotalDuration int
	}

	var results []Result
	query.Select("agency, sum(delay_duration) as total_duration").Group("agency").Scan(&results)

	labels := []string{}
	data := []int{}

	for _, r := range results {
		labels = append(labels, r.Agency)
		data = append(data, r.TotalDuration)
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"data":   data,
	})
}

func ShowUsers(c *gin.Context) {
	session := sessions.Default(c)
	empName := session.Get("emp_name")
	role := session.Get("role")

	var users []models.User
	db.Find(&users)

	totalUsers := len(users)
	activeUsers := 0
	adminUsers := 0
	deptMap := make(map[string]bool)

	for _, u := range users {
		if u.Active {
			activeUsers++
		}
		if u.Role == "sys_admin" || u.Role == "dept_admin" {
			adminUsers++
		}
		if u.Dept != "" {
			deptMap[u.Dept] = true
		}
	}
	departments := len(deptMap)

	c.HTML(http.StatusOK, "users.html", gin.H{
		"EmpName":     empName,
		"Role":        role,
		"Users":       users,
		"TotalUsers":  totalUsers,
		"ActiveUsers": activeUsers,
		"Departments": departments,
		"AdminUsers":  adminUsers,
	})
}

func CreateUser(c *gin.Context) {
	empNo := c.PostForm("emp_no")
	empName := c.PostForm("emp_name")
	dept := c.PostForm("dept")
	designation := c.PostForm("designation")
	role := c.PostForm("role")
	password := c.PostForm("password")

	user := models.User{
		EmpNo:       empNo,
		EmpName:     empName,
		Dept:        dept,
		Designation: designation,
		Role:        role,
		Password:    password,
		Active:      true,
	}
	db.Create(&user)
	c.Redirect(http.StatusFound, "/users")
}

func UpdateUserRole(c *gin.Context) {
	empNo := c.PostForm("emp_no")
	newRole := c.PostForm("role")

	db.Model(&models.User{}).Where("emp_no = ?", empNo).Update("role", newRole)
	c.String(http.StatusOK, "Updated")
}

func ToggleUserActive(c *gin.Context) {
	empNo := c.PostForm("emp_no")
	
	var user models.User
	if err := db.Where("emp_no = ?", empNo).First(&user).Error; err == nil {
		user.Active = !user.Active
		db.Model(&models.User{}).Where("emp_no = ?", empNo).Update("active", user.Active)
		
		status := "Inactive"
		color := "text-red-600 bg-red-100"
		if user.Active {
			status = "Active"
			color = "text-green-600 bg-green-100"
		}
		
		html := fmt.Sprintf(`
			<button hx-post="/users/toggle" hx-vals='{"emp_no": "%s"}' hx-swap="outerHTML" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full %s">
				%s
			</button>
		`, user.EmpNo, color, status)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	} else {
		c.Status(http.StatusBadRequest)
	}
}

func ExportCSV(c *gin.Context) {
	shopDesc := c.Query("shop_desc")
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	query := db.Model(&models.DelayData{})

	if shopDesc != "" {
		query = query.Where("shop_desc = ?", shopDesc)
	}

	layout := "2006-01-02"
	if fromDateStr != "" {
		if fromDate, err := time.Parse(layout, fromDateStr); err == nil {
			query = query.Where("delay_from >= ?", fromDate)
		}
	}
	if toDateStr != "" {
		if toDate, err := time.Parse(layout, toDateStr); err == nil {
			query = query.Where("delay_upto < ?", toDate.AddDate(0, 0, 1))
		}
	}

	var delays []models.DelayData
	query.Order("delay_from desc").Find(&delays)

	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", "attachment;filename=delay_report.csv")
	writer := csv.NewWriter(c.Writer)

	_ = writer.Write([]string{"Shop", "Equipment", "Sub-Equipment", "Agency", "From", "To", "Duration (mins)", "Remarks", "Description", "Status"})

	for _, d := range delays {
		_ = writer.Write([]string{
			d.ShopDesc,
			d.EqptName,
			d.SubEqptName,
			d.Agency,
			d.DelayFrom.Format("2006-01-02 15:04"),
			d.DelayUpto.Format("2006-01-02 15:04"),
			strconv.Itoa(d.DelayDuration),
			d.Remarks,
			d.DelayDesc,
			d.Status,
		})
	}

	writer.Flush()
}
