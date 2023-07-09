package main

import (
	"go_test_prac/webApp/pkg/repository/dbrepo"
	"os"
	"testing"
)

var app application
var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXVkIjoiZXhhbXBsZS5jb20iLCJleHAiOjE2ODg1MDQzNDIsImlzcyI6ImV4YW1wbGUuY29tIiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMSJ9.slr7l3MuTHaPzZgvoDLwjJ0DIbUhJAhblE4mGGEp1gs"

// TestMain is the entry point for the test suite
func TestMain(m *testing.M) {
	app.DB = &dbrepo.TestDBRepo{}
	app.Domain = "example.com"
	app.JWTSecret = "2dce505d96a53c5768052ee90fsdf2055657518ad489160df9913f66042e160"
	os.Exit(m.Run())
}
