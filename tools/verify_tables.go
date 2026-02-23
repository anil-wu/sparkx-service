package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/anil-wu/spark-x/internal/config"
)

var configFile = "etc/sparkx-api.yaml"

func main() {
	var c config.Config
	conf.MustLoad(configFile, &c)

	db, err := sql.Open("mysql", c.MySQL.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SHOW TABLES LIKE 'workspace_%'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Workspace tables found:")
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		fmt.Printf("  - %s\n", tableName)
	}

	// 验证表结构
	fmt.Println("\nVerifying workspace_canvas structure:")
	descCanvas, err := db.Query("DESCRIBE workspace_canvas")
	if err != nil {
		log.Printf("Error describing workspace_canvas: %v", err)
	} else {
		defer descCanvas.Close()
		for descCanvas.Next() {
			var field, colType, null, key, defaultVal, extra string
			descCanvas.Scan(&field, &colType, &null, &key, &defaultVal, &extra)
			fmt.Printf("  %s: %s (NULL: %s, Key: %s, Default: %s)\n", field, colType, null, key, defaultVal)
		}
	}

	fmt.Println("\nVerifying workspace_layer structure:")
	descLayer, err := db.Query("DESCRIBE workspace_layer")
	if err != nil {
		log.Printf("Error describing workspace_layer: %v", err)
	} else {
		defer descLayer.Close()
		for descLayer.Next() {
			var field, colType, null, key, defaultVal, extra string
			descLayer.Scan(&field, &colType, &null, &key, &defaultVal, &extra)
			fmt.Printf("  %s: %s (NULL: %s, Key: %s, Default: %s)\n", field, colType, null, key, defaultVal)
		}
	}
}
