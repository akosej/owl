package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"os"
)

var (
	mysqlCredentials = Config("user_db") + ":" + Config("pass_db") + "@tcp(" + Config("server_db") + ":" + Config("port_db") + ")/" + Config("name_db") + "?charset=utf8&parseTime=True&loc=Local"
	path, _          = os.Getwd()
	port             = Config("server_port")
	agentDHCP        = Config("serverDHCP")
)

//go:embed css/* fonts/* img/* images/* logo/* js/* template/*
var static embed.FS

func main() {
	db, err := gorm.Open("mysql", mysqlCredentials)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	//--Migrate DHCP estruct
	db.AutoMigrate(Subnets{}, Ranges{}, Groups{}, HostGroup{})
	//--Migrate DNS estruct
	db.AutoMigrate(DnsZones{}, DnsEntrys{})
	gin.SetMode(gin.ReleaseMode)
	r := engine()
	r.Use(gin.Logger())
	if err := engine().Run(":" + port); err != nil {
		log.Fatal("Unable to start:", err)
	}
}
