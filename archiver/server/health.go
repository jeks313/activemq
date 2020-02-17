package server

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/hootsuite/healthchecks"
	"github.com/hootsuite/healthchecks/checks/sqlsc"
	"github.com/rs/zerolog/log"
)

// Health sets up the default health router
func Health(r *mux.Router, route string) {
	mydb, err := sql.Open("mysql", "system:tjmwauki@tcp(127.0.0.1:3306)/test")
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to test database")
		return
	}

	// Define a StatusEndpoint at '/status/db' for a database dependency
	db := healthchecks.StatusEndpoint{
		Name:          "The DB",
		Slug:          "db",
		Type:          "internal",
		IsTraversable: false,
		StatusCheck: sqlsc.SQLDBStatusChecker{
			DB: mydb,
		},
		TraverseCheck: nil,
	}

	// Define the list of StatusEndpoints for your service
	statusEndpoints := []healthchecks.StatusEndpoint{db}

	// Set the path for the about and version files
	aboutFilePath := "conf/about.json"
	versionFilePath := "conf/version.txt"

	// Set up any service injected customData for /status/about response.
	// Values can be any valid JSON conversion and will override values set in about.json.
	customData := make(map[string]interface{})
	// Examples:
	//
	// String value
	customData["a-string"] = "some-value"
	//
	// Number value
	customData["a-number"] = 123
	//
	// Boolean value
	customData["a-bool"] = true
	//
	// Array
	// customData["an-array"] = []string{"val1", "val2"}
	//
	// Custom object
	// customObject := make(map[string]interface{})
	// customObject["key1"] = 1
	// customObject["key2"] = "some-value"
	// customData["an-object"] = customObject

	// Register all the "/status/..." requests to use our health checking framework

	r.PathPrefix(route).Handler(healthchecks.Handler(statusEndpoints, aboutFilePath, versionFilePath, customData))
}
