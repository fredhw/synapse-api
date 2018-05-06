package handlers

import "net/http"

/* TODO: implement a CORS middleware handler, as described
in https://drstearns.github.io/tutorials/cors/ that responds
with the following headers to all requests:

  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, PUT, POST, PATCH, DELETE
  Access-Control-Allow-Headers: Content-Type, Authorization
  Access-Control-Expose-Headers: Authorization
  Access-Control-Max-Age: 600
*/

const (
	headerAllowOrigin   = "Access-Control-Allow-Origin"
	headerAllowMethods  = "Access-Control-Allow-Methods"
	headerAllowHeaders  = "Access-Control-Allow-Headers"
	headerExposeHeaders = "Access-Control-Expose-Headers"
	headerMaxAge        = "Access-Control-Max-Age"

	headerOrigin         = "Origin"
	headerAuthorization  = "Authorization"
	headerRequestMethod  = "Access-Control-Request-Method"
	headerRequestHeaders = "Access-Control-Request-Headers"
)

//CORSHandler is a middleware that handles requests
type CORSHandler struct {
	Handler http.Handler
}

func (ch *CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//use the `Origin` and `Access-Control-Request-*` headers to
	//determine if this request should be allowed
	//for example...
	// if r.Header.Get("Origin") != "https://fredhw.me" {
	// 	http.Error(w, "error origin", http.StatusUnauthorized)
	// 	return
	// }

	//set the various CORS response headers depending on
	//what you want your server to allow
	w.Header().Add(headerAllowOrigin, "*")
	w.Header().Add(headerAllowMethods, "GET")
	w.Header().Add(headerAllowMethods, "PUT")
	w.Header().Add(headerAllowMethods, "POST")
	w.Header().Add(headerAllowMethods, "PATCH")
	w.Header().Add(headerAllowMethods, "DELETE")
	w.Header().Add(headerAllowHeaders, headerContentType)
	w.Header().Add(headerAllowHeaders, headerAuthorization)
	w.Header().Add(headerAllowHeaders, "filename")
	w.Header().Add(headerExposeHeaders, headerAuthorization)
	w.Header().Add(headerMaxAge, "600")

	//if this is preflight request, the method will
	//be OPTIONS, so call the real handler only if
	//the method is something else
	if r.Method != "OPTIONS" {
		ch.Handler.ServeHTTP(w, r)
	}
}

//NewCORSHandler adds CORS support to all handler functions in a mux
func NewCORSHandler(handlerToWrap http.Handler) *CORSHandler {
	return &CORSHandler{handlerToWrap}
}
