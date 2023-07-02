package main

import (
	"errors"
	"fmt"
	"go_test_prac/webApp/pkg/data"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const jwtTokenExpiry = time.Minute * 15 // 15分
const refreshTokenExpiry = time.Hour * 24 // 24時間

type TokenPairs struct {
	Token string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

func (app *application) getTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	// we expect our authorization header to be in the format
	// Authorization: Bearer {token}
	// add a header
	w.Header().Add("Vary", "Authorization")

	// get the authorization header
	authHeader := r.Header.Get("Authorization")

	// check if the authorization header is empty
	if authHeader == "" {
		return "", nil, errors.New("authorization header required")
	}

	// split the authorization header on the space
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("authorization header format must be Bearer {token}")
	}

	// check to see if we have the word "Bearer"
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("authorization header must start with Bearer")
	}

	token := headerParts[1]

	// declare an empty Claims{} struct
	claims := &Claims{}

	// parse the JWT token, passing the expected Claims{} struct into the method
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// jwt secretを返す
		return []byte(app.JWTSecret), nil
	})

	// check if there was an error; note that this cathces expired tokens as well
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("token expired")
		}
		return "", nil, err
	}

	// make sure that we issued the token
	if claims.Issuer != app.Domain {
		return "", nil, errors.New("incorrect token issuer")
	}

	// valid token
	return token, claims, nil

}

func (app *application) generateTokenPair(user *data.User) (TokenPairs, error) {
	// create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprintf("%d", user.ID)
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	if user.IsAdmin == 1 {
		claims["admin"] = true
	} else {
		claims["admin"] = false
	}

	// set the expiry
	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()

	// create the signed token
	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create the refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprintf("%d", user.ID)
	// set the expiry; must be longer than the access token
	refreshTokenClaims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()

	// create the signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	var tokenPair = TokenPairs{
		Token: signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	return tokenPair, nil
}
