package main

//-----------------------------------------------------------------
//---------------ESTRUCTURAS DE LA BASE DE DATOS ------------------
//-----------------------------------------------------------------
//---------------ESTRUCTURAS DEL DHCP  ----------------------------
type Subnets struct {
	ID               uint   `json:"id" form:"id"`
	Subnet           string `json:"subnet" form:"subnet" binding:"required"`
	Netmask          string `json:"netmask" form:"netmask" binding:"required"`
	ServerIdentifier string `json:"server_identifier" form:"server_identifier" binding:"required"`
	SubnetMask       string `json:"subnet_mask" form:"subnet_mask" binding:"required"`
	Routers          string `json:"routers" form:"routers" binding:"required"`
	BroadcastAddress string `json:"broadcast_address" form:"broadcast_address" binding:"required"`
	Activa           bool   `json:"activa" form:"activa"`
}

type Ranges struct {
	ID               uint   `json:"id" form:"id"`
	Subnet           string `json:"subnet" form:"subnet" binding:"required"`
	IpaddressInitial string `json:"ipaddress_initial" form:"ipaddress_initial" binding:"required"`
	IpaddressEnd     string `json:"ipaddress_end" form:"ipaddress_end" binding:"required"`
	Activa           bool   `json:"activa" form:"activa"`
}

type Groups struct {
	ID          uint   `json:"id" form:"id"`
	Subnet      string `json:"subnet" form:"subnet" binding:"required"`
	Group       string `json:"group" form:"group" binding:"required"`
	Block       string `json:"block" form:"block" binding:"required"`
	Description string `json:"description" form:"description" binding:"required"`
	Activa      bool   `json:"activa" form:"activa"`
	Assig       int    `json:"assig" form:"assig"`
}

type HostGroup struct {
	ID          uint   `json:"id" form:"id"`
	Group       string `json:"group" form:"group" binding:"required"`
	Host        string `json:"host" form:"host" binding:"required"`
	Macaddress  string `json:"macaddress" form:"macaddress" binding:"required"`
	Ipaddress   string `json:"ipaddress" form:"ipaddress" binding:"required"`
	Description string `json:"description" form:"description"`
	Activa      bool   `json:"activa" form:"activa"`
}

//---------------STRUCTURES DNS   ----------------------------
//-----------------------------------------------------------------
type DnsZones struct {
	ID     uint   `json:"id" form:"id"`
	Name   string `json:"name" form:"name" binding:"required"`
	Active bool   `json:"active" form:"active"`
}

type DnsEntrys struct {
	ID        uint   `json:"id" form:"id"`
	Zone      uint   `json:"zone" form:"zone" binding:"required"`
	Name      string `json:"name" form:"name" binding:"required"`
	Url       string `json:"url" form:"url" binding:"required"`
	Ipaddress string `json:"ipaddress" form:"ipaddress" binding:"required"`
	Active    bool   `json:"active" form:"active"`
}
