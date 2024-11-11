package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // 使用 MySQL 驱动

	"github.com/UTC-Six/booming/internal/types"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

/*
	CREATE INDEX idx_landmines_lat_lon ON landmines (latitude, longitude);
	ALTER TABLE landmines ADD location POINT NOT NULL;
     UPDATE landmines SET location = POINT(longitude, latitude);
     CREATE SPATIAL INDEX idx_landmines_location ON landmines(location);

*/

func init() {
	var err error
	dsn := "username:password@tcp(127.0.0.1:3306)/your_database?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
}

var landmines []types.Landmine

func getAllLandmines() []types.Landmine {
	return landmines
}

func CreateLandmine(c *gin.Context) {
	var landmine types.Landmine
	if err := c.ShouldBindJSON(&landmine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	landmine.Radius = 10.0                                                                 // 设置触发半径为10米
	landmine.Location = fmt.Sprintf("POINT(%f %f)", landmine.Longitude, landmine.Latitude) // 设置 location 字段

	// 将地雷信息保存到数据库中
	query := `INSERT INTO landmines (id, latitude, longitude, radius, location) VALUES (?, ?, ?, ?, ST_GeomFromText(?))`
	_, err := db.Exec(query, landmine.ID, landmine.Latitude, landmine.Longitude, landmine.Radius, landmine.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存地雷信息失败"})
		return
	}

	c.JSON(http.StatusOK, landmine)
}

func CheckProximity(c *gin.Context) {
	var userLocation struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&userLocation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取所有地雷（这里需要实现获取地雷的逻辑）
	landmines, err := getNearbyLandmines(userLocation.Latitude, userLocation.Longitude, 10.0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, landmine := range landmines {
		distance := calculateDistance(userLocation.Latitude, userLocation.Longitude, landmine.Latitude, landmine.Longitude)
		if distance <= landmine.Radius {
			// 触发地雷的逻辑
			c.JSON(http.StatusOK, gin.H{"triggered": true, "landmine_id": landmine.ID})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"triggered": false})
}

// 计算两点之间的距离（单位：米）
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // 地球半径，单位米
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func degreesToRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func getNearbyLandmines(lat, lon, radius float64) ([]types.Landmine, error) {
	query := `
		SELECT id, latitude, longitude, radius, ST_AsText(location) as location
		FROM landmines
		WHERE 
			latitude BETWEEN ? AND ?
			AND longitude BETWEEN ? AND ?
			AND ST_Distance_Sphere(location, POINT(?, ?)) <= ?
	`
	// 计算边界框
	latMin := lat - (radius / 111000.0) // 半径转换为度
	latMax := lat + (radius / 111000.0)
	lonMin := lon - (radius / (111000.0 * math.Cos(lat*math.Pi/180)))
	lonMax := lon + (radius / (111000.0 * math.Cos(lat*math.Pi/180)))

	rows, err := db.Query(query, latMin, latMax, lonMin, lonMax, lon, lat, radius)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var landmines []types.Landmine
	for rows.Next() {
		var lm types.Landmine
		if err := rows.Scan(&lm.ID, &lm.Latitude, &lm.Longitude, &lm.Radius, &lm.Location); err != nil {
			return nil, err
		}
		landmines = append(landmines, lm)
	}
	return landmines, nil
}
