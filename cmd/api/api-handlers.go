package main

import "net/http"

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read a json payload

	// look up the user in the database based on the email address

	// check if the password matches

	// generate a JWT token

	// send the token back to the client

}

func (app *application) refresh(w http.ResponseWriter, r *http.Request) {

}

func (app *application) allUsers(w http.ResponseWriter, r *http.Request) {

}


func (app *application) getUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {

}
