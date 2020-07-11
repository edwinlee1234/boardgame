package middleware

import "net/http"

// Before 全部request前都會先經過
func Before(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalHeader := "aid,cid,origin,x-requested-with,content-type,accept,authentication,set-cookie,auth_token,token,trace_time"
		origin_host := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin_host)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", normalHeader)
		w.Header().Set("Access-Control-Expose-Headers", normalHeader)
		w.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Add("Access-Control-Max-Age", "1728000")

		if r.Method != "OPTIONS" {
			next.ServeHTTP(w, r)
		}
	})
}
