package main

import "net/http"

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS（Cross-Origin Resource Sharing）を有効化
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8090")
		// プリフライトリクエストに対するレスポンス
		//(HTTPヘッダにAccept, Accept-Language, Content-Language, Content-Type以外のフィールド→Authorization)があるので、全てプリフライトリクエストがくる
		if r.Method == http.MethodOptions {
			//Access-Control-Allow-Credentialsは、bearerの認証情報を使用しているため必要
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (app *application) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, err := app.getTokenFromHeaderAndVerify(w,r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
