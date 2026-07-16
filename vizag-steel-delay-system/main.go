package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"vizag-steel-delay-system/handlers"
	"vizag-steel-delay-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(sqlite.Open("project.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.EqptMaster{}, &models.DelayData{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	handlers.SetDB(db)

	seedData(db)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret_session_key"))
	r.Use(sessions.Sessions("vizag_session", store))

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })
	r.GET("/login", handlers.ShowLogin)
	r.POST("/login", handlers.Login)
	r.GET("/logout", handlers.Logout)

	auth := r.Group("/")
	auth.Use(handlers.AuthRequired())
	{
		auth.GET("/entry", handlers.ShowEntry)
		auth.POST("/entry", handlers.SubmitDelay)
		
		auth.POST("/api/start-work/:id", handlers.StartWork)
		auth.POST("/api/mark-done/:id", handlers.MarkDone)
		auth.GET("/reports", handlers.ShowReports)
		auth.GET("/api/reports/export", handlers.ExportCSV)

		auth.POST("/api/import-csv", handlers.ImportCSV)
		auth.GET("/api/equipment", handlers.GetEquipment)
		auth.GET("/api/subequipment", handlers.GetSubEquipment)
		auth.GET("/api/reports/data", handlers.GetReportData)
		auth.GET("/api/reports/chart", handlers.GetChartData)

		admin := auth.Group("/")
		auth.GET("/users", handlers.RoleRequired("sys_admin"), handlers.ShowUsers)
		auth.POST("/users/add", handlers.RoleRequired("sys_admin"), handlers.CreateUser)
		auth.POST("/users/role", handlers.RoleRequired("sys_admin"), handlers.UpdateUserRole)
		admin.Use(handlers.RoleRequired("sys_admin"))
		{
			admin.POST("/users/toggle", handlers.ToggleUserActive)
		}
	}

	log.Println("Starting server on :8090")
	if err := r.Run(":8090"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("emp_no = ?", "1111").Count(&count)
	if count == 0 {
		db.Create(&models.User{
			EmpNo:       "1111",
			Password:    "test",
			EmpName:     "Master Admin",
			Dept:        "IT",
			Designation: "System Administrator",
			Role:        "sys_admin",
			Active:      true,
		})
		log.Println("Seeded master admin user (1111)")
	}

	db.Model(&models.EqptMaster{}).Count(&count)
	if count == 0 {
		eqpts := []models.EqptMaster{
			{ShopCode: "BF", ShopDesc: "Blast Furnace", EqptCode: "Furnace 1", SubEqptCode: "Tuyeres"},
			{ShopCode: "BF", ShopDesc: "Blast Furnace", EqptCode: "Furnace 1", SubEqptCode: "Taphole"},
			{ShopCode: "BF", ShopDesc: "Blast Furnace", EqptCode: "Cast House", SubEqptCode: "Runner"},
			{ShopCode: "SMS", ShopDesc: "Steel Melting Shop", EqptCode: "Converter A", SubEqptCode: "Vessel"},
			{ShopCode: "SMS", ShopDesc: "Steel Melting Shop", EqptCode: "Converter A", SubEqptCode: "Lance"},
			{ShopCode: "SMS", ShopDesc: "Steel Melting Shop", EqptCode: "Ladle Furnace", SubEqptCode: "Electrodes"},
			{ShopCode: "WRM", ShopDesc: "Wire Rod Mill", EqptCode: "Roughing Mill", SubEqptCode: "Rolls"},
			{ShopCode: "WRM", ShopDesc: "Wire Rod Mill", EqptCode: "Finishing Block", SubEqptCode: "Guides"},
		}
		for _, e := range eqpts {
			db.Create(&e)
		}
		log.Println("Seeded equipment master data")
	}

	db.Model(&models.DelayData{}).Count(&count)
	if count == 0 {
		importLocalCSV(db)
	}
}

func importLocalCSV(db *gorm.DB) {

	file, err := os.Open("sample_delays_data.csv")
	if err == nil {
		defer file.Close()
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err == nil {
			var delays []models.DelayData
			for i, row := range records {
				if len(delays) >= 50 {
					break
				}
				if i == 0 || len(row) < 8 {
					continue
				}
				getCol := func(idx int) string {
					if idx < len(row) {
						return row[idx]
					}
					return ""
				}

				delDateStr := getCol(0)
				delayFromStr := getCol(5)
				delayToStr := getCol(6)
				delayDurStr := getCol(7)
				cumDur, _ := strconv.ParseFloat(getCol(8), 64)
				effDur, _ := strconv.ParseFloat(getCol(17), 64)
				allocDur, _ := strconv.ParseFloat(delayDurStr, 64)

				agencies := []string{"operations", "electrical", "mechanical", "shutdown"}
				agencyName := agencies[i%4]

				var baseDate time.Time
				baseDate, _ = time.Parse("02-01-2006", delDateStr)
				if baseDate.IsZero() {
					baseDate = time.Now()
				}

				delay := models.DelayData{
					ShopCode:          getCol(1),
					ShopDesc:          getCol(1),
					EqptName:          getCol(9),
					SubEqptName:       getCol(10),
					Agency:            agencyName,
					DelayFrom:         handlers.ParseDecimalTime(baseDate, delayFromStr),
					DelayUpto:         handlers.ParseDecimalTime(baseDate, delayToStr),
					DelayDuration:     int(effDur * 60),
					AllocatedDuration: int(allocDur * 60),
					Status:            "Work Done",
					DelayDesc:         getCol(12),
					UserEntered:       "1111",
					Timestamp:         time.Now(),
					DelDate:           delDateStr,
					Material:          getCol(2),
					Rake:              getCol(3),
					CumDelay:          cumDur,
					Remarks:           getCol(11),
					AgencyCode:        getCol(13),
					DelayFreq:         getCol(14),
					Continuity:        getCol(15),
					ExpectedComp:      getCol(16),
					EffDuration:       effDur,
					LegacyDelayID:     getCol(22),
				}
				delays = append(delays, delay)
			}
			
			db.Exec("PRAGMA synchronous = OFF")
			db.Exec("PRAGMA journal_mode = MEMORY")
			
			db.Transaction(func(tx *gorm.DB) error {
				return tx.CreateInBatches(delays, 5000).Error
			})
			
			db.Exec("PRAGMA synchronous = NORMAL")
			db.Exec("PRAGMA journal_mode = WAL")
			
			log.Println("Seeded data directly from sample_delays_data.csv")
			return
		}
	}

	now := time.Now()
	delays := []models.DelayData{
		{
			ShopCode:          "BF",
			ShopDesc:          "Blast Furnace",
			EqptName:          "Furnace 1",
			SubEqptName:       "Taphole",
			Agency:            "operations",
			DelayFrom:         now.Add(-24 * time.Hour),
			DelayUpto:         now.Add(-20 * time.Hour),
			DelayDuration:     240,
			AllocatedDuration: 240,
			Status:            "Work Done",
			DelayDesc:         "Taphole clay hardened, required extended manual drilling.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-24 * time.Hour),
		},
		{
			ShopCode:          "WRM",
			ShopDesc:          "Wire Rod Mill",
			EqptName:          "Roughing Mill",
			SubEqptName:       "Rolls",
			Agency:            "electrical",
			DelayFrom:         now.Add(-10 * time.Hour),
			DelayUpto:         now.Add(-9 * time.Hour),
			DelayDuration:     60,
			AllocatedDuration: 60,
			Status:            "Work Done",
			DelayDesc:         "Motor trip due to overload. Relay reset and tested.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-10 * time.Hour),
		},
		{
			ShopCode:          "SMS",
			ShopDesc:          "Steel Melting Shop",
			EqptName:          "Converter A",
			SubEqptName:       "Lance",
			Agency:            "mechanical",
			DelayFrom:         now.Add(-8 * time.Hour),
			DelayUpto:         now.Add(-5 * time.Hour),
			DelayDuration:     180,
			AllocatedDuration: 180,
			Status:            "Work Done",
			DelayDesc:         "Lance tip replacement required due to heavy wear and tear.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-8 * time.Hour),
		},
		{
			ShopCode:          "BF",
			ShopDesc:          "Blast Furnace",
			EqptName:          "Cast House",
			SubEqptName:       "Runner",
			Agency:            "shutdown",
			DelayFrom:         now.Add(-4 * time.Hour),
			DelayUpto:         now.Add(-1 * time.Hour),
			DelayDuration:     180,
			AllocatedDuration: 180,
			Status:            "Work Done",
			DelayDesc:         "Scheduled minor shutdown for runner refractory repair.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-4 * time.Hour),
		},
		{
			ShopCode:          "SMS",
			ShopDesc:          "Steel Melting Shop",
			EqptName:          "Ladle Furnace",
			SubEqptName:       "Electrodes",
			Agency:            "electrical",
			DelayFrom:         now.Add(-45 * time.Minute),
			AllocatedDuration: 90,
			Status:            "In Progress",
			DelayDesc:         "Electrode slipping mechanism jammed. Inspection ongoing.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-45 * time.Minute),
		},
		{
			ShopCode:          "WRM",
			ShopDesc:          "Wire Rod Mill",
			EqptName:          "Finishing Block",
			SubEqptName:       "Guides",
			Agency:            "mechanical",
			DelayFrom:         now.Add(-10 * time.Minute),
			AllocatedDuration: 45,
			Status:            "Pending",
			DelayDesc:         "Roller guide alignment issue causing cobbles.",
			UserEntered:       "1111",
			Timestamp:         now.Add(-10 * time.Minute),
		},
	}
	for _, d := range delays {
		db.Create(&d)
	}
	log.Println("Seeded fallback dummy data")
}
