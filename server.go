package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
)

//------------------------
const (
	userkey = "user"
)

//------------------------
func engine() *gin.Engine {
	r := gin.New()
	//r.Delims("<owl", "owl>")
	//-- Cargar plantillas
	templates := template.Must(template.New("").ParseFS(static, "template/*.html"))
	r.SetHTMLTemplate(templates)

	//-- Cargar archivos estaticos
	r.StaticFS("/static/", http.FS(static))
	r.Use(sessions.Sessions("mysession", cookie.NewStore([]byte("secret"))))
	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(userkey)
		if user == nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"session": user})
		} else {
			http.Redirect(c.Writer, c.Request, "/private/subnets", http.StatusFound)
		}
	})
	r.POST("/login", login)
	r.GET("/logout", logout)

	private := r.Group("/private")
	private.Use(AuthRequired)
	{
		private.GET("/allHost", allHost)
		private.GET("/subnets", pagSubnets)
		private.GET("/subnet/:id", showGroupSubnet)
		//-----DNS
		private.GET("/dns/zones", zonal)
		private.GET("/dns/zone/:id", zonedetails)

		private.POST("/AddSubnet", AddSubnet)
		private.POST("/DelSubnet", DelSubnet)
		private.POST("/EditActiveSubnet", EditActiveSubnet)
		private.POST("/EditActiveHost", EditActiveHost)
		private.POST("/AddGroup", AddGroup)
		private.POST("/EditGroup", EditGroup)
		private.POST("/EditHostGroup", EditHostGroup)
		private.POST("/DelGroup", DelGroup)
		private.POST("/AddHostGroup", AddHostGroup)
		private.POST("/DelHost", DelHost)
		private.POST("/EditSubnet", EditSubnet)
		private.POST("/MoveHost", MoveHost)
		private.POST("/AddRange", AddRange)
		private.POST("/EditRange", EditRange)
		private.POST("/DelRange", DelRange)
		private.POST("/EditActiveRange", EditActiveRange)
		private.POST("/Run", RunConf)

		//	--Post DNS
		private.POST("/dns/AddZone", AddZone)
		private.POST("/dns/AddEntry", AddEntry)
		private.POST("/dns/EditActiveEntry", EditActiveEntry)
		private.POST("/dns/DelEntry", DelEntry)
		private.POST("/dns/DelZone", DelZone)
	}
	return r
}

//-----------

//-----------
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		// Abort the request with the appropriate error code
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

// login is a handler that parses a form and checks for specific data
func login(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate form input
	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters can't be empty"})
		return
	}

	// Check for username and password match, usually from a database
	if username != Config("user") || password != Config("pwd") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Save the username in the session
	session.Set(userkey, username) // In real world usage you'd set this to the users ID
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated user"})
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		http.Redirect(c.Writer, c.Request, "/", http.StatusFound)
		return
	}
	session.Delete(userkey)
	if err := session.Save(); err != nil {
		http.Redirect(c.Writer, c.Request, "/", http.StatusFound)
		return
	}
	http.Redirect(c.Writer, c.Request, "/", http.StatusFound)
}

//----allHost
func allHost(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var hosts []HostGroup
	db.Find(&hosts)

	var subnets []Subnets
	db.Find(&subnets)
	session := sessions.Default(c)
	user := session.Get(userkey)

	c.HTML(http.StatusOK, "allhosts.html", gin.H{"session": user, "hosts": hosts, "subnets": subnets})
}

//----Subnets
func pagSubnets(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var subnets []Subnets
	db.Find(&subnets)
	session := sessions.Default(c)
	user := session.Get(userkey)

	c.HTML(http.StatusOK, "subnets.html", gin.H{"session": user, "subnets": subnets})
}

//-----Show group subnet
func showGroupSubnet(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	session := sessions.Default(c)
	user := session.Get(userkey)

	id := c.Params.ByName("id")
	var subnet Subnets
	idSubnet, _ := strconv.Atoi(id)
	db.Where("ID = ?", idSubnet).Find(&subnet)

	var groups []Groups
	db.Where("Subnet = ?", id).Find(&groups)

	var ranges []Ranges
	db.Where("Subnet = ?", id).Find(&ranges)

	c.HTML(http.StatusOK, "groups_subnet.html", gin.H{"session": user, "subnet": subnet, "ranges": ranges, "groups": groups})

}

//--add subnet
func AddSubnet(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	subnet := c.PostForm("subnet")
	netmask := c.PostForm("netmask")
	serverIdentifier := c.PostForm("server_identifier")
	subnetMask := c.PostForm("subnet_mask")
	routers := c.PostForm("routers")
	broadcastAddress := c.PostForm("broadcast_address")

	var subnets Subnets
	db.Where("Subnet = ? ", subnet).Find(&subnets)

	if subnets.ID == 0 {
		subnets.Subnet = subnet
		subnets.Netmask = netmask
		subnets.ServerIdentifier = serverIdentifier
		subnets.SubnetMask = subnetMask
		subnets.Routers = routers
		subnets.BroadcastAddress = broadcastAddress
		subnets.Activa = true
		db.Save(&subnets)
	}

	c.JSON(200, gin.H{"msg": "add subnet"})
	return

}

//--Edit subnet
func EditSubnet(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	subnet := c.PostForm("subnet")
	netmask := c.PostForm("netmask")
	serverIdentifier := c.PostForm("server_identifier")
	subnetMask := c.PostForm("subnet_mask")
	routers := c.PostForm("routers")
	broadcastAddress := c.PostForm("broadcast_address")

	var subnets Subnets
	db.Where("ID = ? ", id).Find(&subnets)

	subnets.Subnet = subnet
	subnets.Netmask = netmask
	subnets.ServerIdentifier = serverIdentifier
	subnets.SubnetMask = subnetMask
	subnets.Routers = routers
	subnets.BroadcastAddress = broadcastAddress

	db.Save(&subnets)

	c.JSON(200, gin.H{"msg": "add subnet"})
	return

}

//--Move Host
func MoveHost(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	g := c.PostForm("group")
	id := c.PostForm("id")

	var host HostGroup
	db.Where("ID = ?", id).Find(&host)

	var group Groups
	db.Where("ID = ? ", g).Find(&group)

	host.Host = group.Group + "-" + group.IPLIBRE()
	host.Group = g
	host.Ipaddress = group.Block + "." + group.IPLIBRE()

	db.Save(&host)

	s1, _ := strconv.ParseInt(group.IPLIBRE(), 10, 64)
	if s1 == int64(group.Assig) {
		group.Assig = group.Assig + 1
		db.Save(&group)
	}

	c.JSON(200, gin.H{"msg": "Move host "})
	return

}

//--Add range to the subnet
func AddRange(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	subnet := c.PostForm("subnet")
	ipInitial := c.PostForm("ip_initial")
	ipEnd := c.PostForm("ip_end")

	var ranges Ranges

	ranges.Subnet = subnet
	ranges.IpaddressInitial = ipInitial
	ranges.IpaddressEnd = ipEnd
	ranges.Activa = true
	db.Save(&ranges)

	c.JSON(200, gin.H{"msg": "Add range to the subnet"})
	return

} //--Add range to the subnet

func EditRange(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	ipInitial := c.PostForm("ip_initial")
	ipEnd := c.PostForm("ip_end")

	var ranges Ranges
	db.Where("ID = ?", id).Find(&ranges)
	ranges.IpaddressInitial = ipInitial
	ranges.IpaddressEnd = ipEnd
	db.Save(&ranges)

	c.JSON(200, gin.H{"msg": "Edit range to the subnet"})
	return

}

func DelRange(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	var ranges Ranges
	db.Where("ID = ?", id).Delete(&ranges)
	c.JSON(200, gin.H{"msg": "Delete range"})
	return

}

func EditActiveRange(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var us Ranges
	id := c.PostForm("id")
	db.Where("ID = ? ", id).Find(&us)
	val := c.PostForm("locked")
	if val == "true" {
		us.Activa = true
	} else {
		us.Activa = false
	}
	db.Save(&us)
	c.JSON(200, gin.H{"msg": "Updated active range"})
	return

}

//--Edit active subnet
func EditActiveSubnet(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var us Subnets
	id := c.PostForm("id")
	db.Where("ID = ? ", id).Find(&us)
	val := c.PostForm("locked")
	if val == "true" {
		us.Activa = true
	} else {
		us.Activa = false
	}
	db.Save(&us)
	c.JSON(200, gin.H{"msg": "Updated active subnet"})
	return

}

//--Edit active host
func EditActiveHost(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var us HostGroup
	id := c.PostForm("id")
	db.Where("ID = ? ", id).Find(&us)
	val := c.PostForm("locked")
	if val == "true" {
		us.Activa = true
	} else {
		us.Activa = false
	}
	db.Save(&us)
	c.JSON(200, gin.H{"msg": "Updated active host"})
	return

}

//--Add group subnet
func AddGroup(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	subnet := c.PostForm("subnet")
	group := c.PostForm("group")
	block := c.PostForm("block")
	description := c.PostForm("description")

	var groups Groups

	db.Where("Group = ?", group).Find(&groups)
	if groups.Group != group {
		groups.Subnet = subnet
		groups.Group = group
		groups.Block = block
		groups.Description = description
		groups.Activa = true
		groups.Assig = 1
		db.Save(&groups)
	}
	c.JSON(200, gin.H{"msg": "add group"})
	return
}

//--Edit group subnet
func EditGroup(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	group := c.PostForm("group")
	block := c.PostForm("block")
	description := c.PostForm("description")

	var groups Groups

	db.Where("ID = ?", id).Find(&groups)
	if groups.ID != 0 {
		groups.Group = group
		groups.Block = block
		groups.Description = description
		groups.Activa = true
		db.Save(&groups)
	}

	for _, host := range groups.HostsLIST() {
		if host.Group == id {
			blockHost := host.Ipaddress[:strings.LastIndex(host.Ipaddress, ".")]
			ip := host.Ipaddress[strings.LastIndex(host.Ipaddress, ".")+1:]
			if block != blockHost {
				var h HostGroup
				db.Where("ID = ?", host.ID).Find(&h)
				h.Ipaddress = block + "." + ip
				db.Save(h)
			}
		}
	}

	c.JSON(200, gin.H{"msg": "Edit group"})
	return
}

//--Del group subnet
func DelSubnet(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")

	var subnets Subnets
	var groups []Groups

	u64, _ := strconv.ParseUint(id, 10, 32)
	db.Where("Subnet = ?", uint(u64)).Find(&groups)
	for _, group := range groups {
		//-------------
		var hosts []HostGroup
		db.Find(&hosts)
		for _, host := range hosts {
			u64, _ := strconv.ParseUint(host.Group, 10, 32)
			if uint(u64) == group.ID {
				db.Where("ID = ?", host.ID).Delete(&hosts)
			}
		}
		db.Where("ID = ?", group.ID).Delete(&groups)
		//-------------
	}
	db.Where("ID = ?", id).Delete(&subnets)
	c.JSON(200, gin.H{"msg": "Del group"})
	return
}

//--Del group subner
func DelGroup(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	var groups Groups
	u64, _ := strconv.ParseUint(id, 10, 32)

	var hosts []HostGroup
	db.Find(&hosts)

	for _, host := range hosts {
		if host.Group == id {
			db.Where("ID = ?", host.ID).Delete(&hosts)
		}
	}
	db.Where("ID = ?", uint(u64)).Find(&groups).Delete(&groups)

	c.JSON(200, gin.H{"msg": "Del group"})
	return
}

//--Add group subner
func AddHostGroup(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	host := c.PostForm("host")
	group := c.PostForm("group")
	ipacaddress := c.PostForm("ipacaddress")
	macaddress := c.PostForm("macaddress")
	description := c.PostForm("description")

	res := strings.Replace(macaddress, "o", "0", 12)
	res1 := strings.Replace(res, "-", ":", 5)

	var hostgroup HostGroup
	var msg string
	db.Where("Host = ? or Ipaddress = ? ", host, ipacaddress).Find(&hostgroup)
	if hostgroup.ID == 0 {
		var hostgroup1 HostGroup
		db.Where("Macaddress = ?", res1).Find(&hostgroup1)
		if hostgroup1.ID == 0 || hostgroup1.Group != group {
			hostgroup.Host = host
			hostgroup.Group = group
			hostgroup.Macaddress = res1
			hostgroup.Ipaddress = ipacaddress
			hostgroup.Description = description
			hostgroup.Activa = true
			db.Save(&hostgroup)
			msg = "Add host"

			var groups Groups
			db.Where("ID = ?", group).Find(&groups)
			cortar := strings.Split(ipacaddress, ".")
			s1, _ := strconv.ParseInt(cortar[3], 10, 64)
			if s1 == int64(groups.Assig) {
				groups.Assig = groups.Assig + 1
				db.Save(&groups)
			}
		}
	} else {
		msg = hostgroup.Macaddress
	}

	c.JSON(200, gin.H{"msg": msg})
	return
}

//--Add group subner
func EditHostGroup(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	host := c.PostForm("host")
	ipacaddress := c.PostForm("ipacaddress")
	macaddress := c.PostForm("macaddress")
	description := c.PostForm("description")

	res := strings.Replace(macaddress, "o", "0", 12)
	res1 := strings.Replace(res, "-", ":", 5)

	var hostgroup HostGroup
	var msg string
	db.Where("ID = ?", id).Find(&hostgroup)

	hostgroup.Host = host
	hostgroup.Macaddress = res1
	hostgroup.Ipaddress = ipacaddress
	hostgroup.Description = description
	db.Save(&hostgroup)
	msg = "Edit host"

	c.JSON(200, gin.H{"msg": msg})
	return
}

//--Add group subner
func DelHost(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")

	var hostgroup HostGroup

	db.Where("ID = ?", id).Find(&hostgroup).Delete(&hostgroup)

	c.JSON(200, gin.H{"msg": "Dell host"})
	return
}

//-----Run
func RunConf(c *gin.Context) {
	key := c.PostForm("key")
	result := ""
	if key == Config("keyRun") {
		result = ConnDhcpAgent()
	}
	c.JSON(200, gin.H{"msg": result})
	return
}
func ConnDhcpAgent() string {
	connection, _ := net.Dial("tcp", agentDHCP+":27000")
	defer connection.Close()
	//--- orden + name_client
	order := fillString("rewrite_configuration", 64)
	_, _ = connection.Write([]byte(order))
	//---
	hearResult := make([]byte, 64)
	_, _ = connection.Read(hearResult)
	Result := strings.Trim(string(hearResult), ":")
	return Result
}
func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

//--------------------------------------------
//--------DNS
func zonal(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var zones []DnsZones
	db.Find(&zones)
	session := sessions.Default(c)
	user := session.Get(userkey)

	c.HTML(http.StatusOK, "zones.html", gin.H{"session": user, "zones": zones})
}

func zonedetails(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.Params.ByName("id")

	var zone DnsZones
	db.Where("ID = ?", id).Find(&zone)

	var entrys []DnsEntrys
	db.Where("Zone = ?", id).Find(&entrys)
	session := sessions.Default(c)
	user := session.Get(userkey)

	c.HTML(http.StatusOK, "entrys.html", gin.H{"session": user, "zone": zone, "entrys": entrys})
}

func AddZone(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	name := c.PostForm("name")

	var zone DnsZones
	db.Where("Name = ?", name).Find(&zone)

	if zone.ID == 0 {
		zone.Name = name
		zone.Active = true
		db.Save(&zone)
	}

	c.JSON(200, gin.H{"msg": "Add zone"})
	return

}

func AddEntry(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	name := c.PostForm("name")
	ipaddress := c.PostForm("ipaddress")
	zone := c.PostForm("zone")
	zoneName := c.PostForm("zone_name")
	u64, _ := strconv.ParseUint(zone, 10, 32)

	var entry DnsEntrys
	db.Where("Name = ?", name).Find(&entry)

	if entry.ID == 0 {
		entry.Zone = uint(u64)
		entry.Name = name
		entry.Url = name + "." + zoneName + "."
		entry.Ipaddress = ipaddress
		entry.Active = true
		db.Save(&entry)
	}
	c.JSON(200, gin.H{"msg": "Add entry "})
	return
}

func EditActiveEntry(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var entry DnsEntrys
	id := c.PostForm("id")
	db.Where("ID = ? ", id).Find(&entry)
	val := c.PostForm("locked")
	if val == "true" {
		entry.Active = true
	} else {
		entry.Active = false
	}
	db.Save(&entry)
	c.JSON(200, gin.H{"msg": "Updated active entry"})
	return

}

func DelEntry(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	var entry DnsEntrys
	id := c.PostForm("id")
	db.Where("ID = ? ", id).Delete(&entry)
	c.JSON(200, gin.H{"msg": "Delete entry"})
	return

}

func DelZone(c *gin.Context) {
	db, _ := gorm.Open("mysql", mysqlCredentials)
	defer db.Close()
	id := c.PostForm("id")
	var zone DnsZones
	db.Where("ID = ? ", id).Delete(&zone)
	var entry []DnsEntrys
	db.Where("Zone = ? ", id).Delete(&entry)
	c.JSON(200, gin.H{"msg": "Delete entry and zone"})
	return
}
