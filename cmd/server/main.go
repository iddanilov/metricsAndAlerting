// Package server/main running server application
package main

import "github.com/iddanilov/metricsAndAlerting/cmd/server/app"

// @title Metric and Alerting
// @version 0.0.2
// @description API Server for Metric and Alerting Application

// @host localhost:8000
// @BasePath /
func main() {
	app.Run()
}
