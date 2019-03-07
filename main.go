package main

import (
	"encoding/hex"
	"git.icinga.com/icingadb/icingadb-connection"
	"git.icinga.com/icingadb/icingadb-ha"
	"git.icinga.com/icingadb/icingadb-json-decoder"
	"git.icinga.com/icingadb/icingadb-main/configobject"
	"git.icinga.com/icingadb/icingadb-main/supervisor"
	"log"
)

func main() {
	chErr := make (chan error)
	redisConn, err := icingadb_connection.NewRDBWrapper("127.0.0.1:6379")
	if err != nil {
		return
	}

	mysqlConn, err := icingadb_connection.NewDBWrapper("module-dev:icinga0815!@tcp(127.0.0.1:3306)/icingadb"	)
	if err != nil {
		return
	}

	super := supervisor.Supervisor{
		ChEnv: make(chan *icingadb_ha.Environment),
		ChDecode: make(chan *icingadb_json_decoder.JsonDecodePackage),
		Rdbw: redisConn,
		Dbw: mysqlConn,
	}

	ha := icingadb_ha.HA{}
	go ha.Run(super.Rdbw, super.Dbw, super.ChEnv, chErr)
	go func() {
		chErr <- icingadb_ha.IcingaEventsBroker(redisConn, super.ChEnv)
	}()

	go icingadb_json_decoder.DecodePool(super.ChDecode, chErr, 16)

	go func() {
		chErr <- configobject.HostOperator(&super)
	}()

	for {
		select {
		case err := <- chErr:
			if err != nil {
				log.Fatal(err)
				return
			}
		case env := <- super.ChEnv: {
			if env != nil {
				log.Print("Got env: " + hex.EncodeToString(env.ID))
			}
		}
		}
	}

	//go create object type supervisors



}