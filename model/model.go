package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var DB gorm.DB

type Server struct {
	Id         int64
	Ip         string
	InstanceId string
	Kitchen    Kitchen
	XiaoLong   XiaoLong
	IsRouter   bool
}

type Kitchen struct {
	Id       int64
	ServerId int64
}

type XiaoLong struct {
	Id       int64
	ServerId int64
	Dockers  []Docker
}

type Docker struct {
	Id string
}

func init() {

	var err error
	DB, err = gorm.Open("postgres", "dbname=chushitest sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.DropTableIfExists(&Server{})
}
