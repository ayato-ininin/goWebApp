package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func getSession() *scs.SessionManager {
	session := scs.New() // sessionマネージャー
	session.Lifetime = 24 * time.Hour // 24時間
	session.Cookie.Persist = true // ブラウザを閉じてもセッションを保持
	// Laxモードでは、GETリクエストによるクロスサイトリクエストであればCookieを送信するが、
	// POSTリクエストや他のHTTPメソッドによるクロスサイトリクエストではCookieを送信しない。
	session.Cookie.SameSite = http.SameSiteLaxMode // CSRF攻撃のリスクを軽減
	session.Cookie.Secure = true // HTTPSのみでCookieを送信(http等通信が暗号化されていない場合使用しないため)

	return session
}
