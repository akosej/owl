package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

//--------------------------------------------------------------------------------
//EXTRAER VARIABLES DEL FICHERO DE CONFIGURACION
func Config(data string) string {
	var value string
	lines, err := File2lines(path + "/config.owl")
	if err != nil {
		fmt.Println("No se pudo encontrar el fichero de configuracion")
	} else {
		for _, line := range lines {
			if strings.Contains(line, "ad."+data) {
				corte := strings.Split(line, "=")
				value = corte[1]
			}
		}
	}
	return value
}

//--------------------------------------------------------------------------------
//-------Filters
func (g Groups) HostsLIST() []HostGroup {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var host []HostGroup
	db.Find(&host)

	return host
}

//------
func (g Groups) IPLIBRE() string {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var i int
	for i = 1; i < g.Assig; i++ {
		s1 := strconv.FormatInt(int64(i), 10)
		var hosts []HostGroup
		db.Where("Ipaddress = ?", g.Block+"."+s1).Find(&hosts)
		if len(hosts) == 0 {
			return s1
		}
	}
	s2 := strconv.FormatInt(int64(g.Assig), 10)
	return s2
}

//convertir en uin
func (h HostGroup) Uint() uint {
	//--------------
	u64, _ := strconv.ParseUint(h.Group, 10, 32)
	return uint(u64)
}

//--Group subnet
func (s Subnets) ListGroup() []Groups {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var groups []Groups
	db.Where("Subnet = ?", s.ID).Find(&groups)
	return groups
}
