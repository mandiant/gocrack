package web

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
	"github.com/fireeye/gocrack/shared/ginlog"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/tankbusta/gzip"
)

const currentAPIVer = "/api/v2"

func init() {
	gin.DefaultWriter = log.Logger
}

type apiError struct {
	Error string `json:"error"`
}

// APIValidationErrors describes errors that occurred in the validation of a request
type APIValidationErrors struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"validation_errors"`
}

type WebAPIError struct {
	StatusCode            int
	UserError             string
	CanErrorBeShownToUser bool
	Err                   error
}

// MarshalJSON builds a JSON version of WebAPIError and modifies the error message based on the struct
func (e *WebAPIError) MarshalJSON() ([]byte, error) {
	// If Gin is in debug mode, go ahead and return the error in the response
	if !e.CanErrorBeShownToUser && gin.IsDebugging() {
		e.CanErrorBeShownToUser = true
	}

	if e.CanErrorBeShownToUser {
		if e.Err == nil {
			goto GenericError
		}
		return json.Marshal(&apiError{
			Error: e.Err.Error(),
		})
	}

GenericError:
	if e.UserError == "" {
		e.UserError = "An unknown error occurred while handling your request. Please try again later"
	}

	return json.Marshal(&apiError{
		Error: e.UserError,
	})
}

func (e WebAPIError) Error() string {
	return e.Err.Error()
}

// WebAPI defines the type of most gocrack API's.
type WebAPI func(c *gin.Context) *WebAPIError

// WrapAPIForError is a wrapper around Gin handlers to reduce error handling code
func WrapAPIForError(f WebAPI) gin.HandlerFunc {
	return func(c *gin.Context) {
		if werr := f(c); werr != nil {
			evt := log.Error().Int("status_code", werr.StatusCode)

			if werr.Err != nil {
				evt.Err(werr.Err)
			}

			evt.Msg("An error occurred while handling an API request")
			c.JSON(werr.StatusCode, werr)
		}
	}
}

func getClaimInformation(c *gin.Context) *authentication.AuthClaim {
	ac, ok := c.Get("claim")
	if !ok {
		return nil
	}

	if claim, ok := ac.(*authentication.AuthClaim); ok {
		return claim
	}
	return nil
}

func newHTTPServer(cfg Config, s *Server) *http.Server {
	var isCSRFEnabled = true

	engine := gin.New()
	engine.Use(gin.Recovery(), setSecureHeaders(), ginlog.LogRequests(), gzip.Gzip(gzip.DefaultCompression))

	engine.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "HEAD", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Xsrf-Token"},
		AllowCredentials: true,
		MaxAge:           cfg.CORS.MaxPreflightAge.Duration,
		AllowOrigins:     cfg.CORS.AllowedOrigins,
	}))

	if cfg.UserInterface.CSRFEnabled != nil && !*cfg.UserInterface.CSRFEnabled {
		isCSRFEnabled = false
	}

	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	rootAPIG := engine.Group(currentAPIVer)
	rootAPIG.Use(shared.RecordAPIMetrics(requestDuration, requestCounter))
	{
		rootAPIG.POST("/login", setXSRFTokenIfNecessary(isCSRFEnabled), WrapAPIForError(s.webSubmitLogin))
		rootAPIG.POST("/users/register", setXSRFTokenIfNecessary(isCSRFEnabled), WrapAPIForError(s.webRegisterNewUser))
	}

	rootAPIG.Use(s.requestHasValidAuth(), setXSRFTokenIfNecessary(isCSRFEnabled), shared.RecordAPIMetrics(requestDuration, requestCounter))
	{
		rootAPIG.GET("/workers/", WrapAPIForError(s.webGetActiveWorkers))
		rootAPIG.GET("/version/", WrapAPIForError(s.webGetVersion))

		rootAPIG.POST("/task/", WrapAPIForError(s.webCreateTask))
		rootAPIG.GET("/task/", WrapAPIForError(s.getAvailableTasks))

		granularTaskV2 := rootAPIG.Group("/task/:taskid").Use(checkParamValidUUID("taskid"), s.checkIfUserIsEntitled("taskid", storage.EntitlementTask))
		{
			granularTaskV2.GET("", s.logAction(storage.ActivityViewTask, "taskid"), WrapAPIForError(s.webGetTaskInfo))
			granularTaskV2.PATCH("", s.logAction(storage.ActivityModifiedTask, "taskid"), WrapAPIForError(s.webModifyTask))
			granularTaskV2.DELETE("", checkIfUserIsAdmin(), WrapAPIForError(s.webDeleteTask))
			granularTaskV2.GET("/passwords", s.logAction(storage.ActivityViewPasswords, "taskid"), WrapAPIForError(s.webGetTaskPasswords))
			granularTaskV2.GET("/entitlements", WrapAPIForError(s.webGetTaskEntitlements))
			granularTaskV2.PATCH("/status", s.logAction(storage.ActivityModifiedTask, "taskid"), WrapAPIForError(s.webChangeTaskStatus))
		}

		rootAPIG.GET("/files/task/", WrapAPIForError(s.webListAvailableTaskFiles))
		rootAPIG.DELETE("/files/task/:fileid", checkParamValidUUID("fileid"), WrapAPIForError(s.webDeleteFile(deleteTaskFileAPI)))
		rootAPIG.GET("/files/task/:fileid/download", checkParamValidUUID("fileid"), s.checkIfUserIsEntitled("fileid", storage.EntitlementTaskFile), WrapAPIForError(s.webDownloadTaskFile))
		rootAPIG.PUT("/files/task/:filename", WrapAPIForError(s.webUploadTaskFile))

		rootAPIG.GET("/files/engine/", WrapAPIForError(s.webGetEngineFiles))
		rootAPIG.DELETE("/files/engine/:fileid", checkParamValidUUID("fileid"), WrapAPIForError(s.webDeleteFile(deleteEngineFileAPI)))
		rootAPIG.GET("/files/engine/:fileid/download", checkParamValidUUID("fileid"), WrapAPIForError(s.webDownloadEngineFile))
		rootAPIG.PUT("/files/engine/:filename", WrapAPIForError(s.webUploadEngineFile))

		rootAPIG.GET("/engine/hashcat/hash_modes", s.apiHashcatGetTaskModes)

		rootAPIG.GET("/users/", WrapAPIForError(s.webGetUsers))
		rootAPIG.GET("/users/:user_uuid", checkParamValidUUID("user_uuid"), WrapAPIForError(s.webGetUser))
		rootAPIG.PATCH("/users/:user_uuid", checkParamValidUUID("user_uuid"), WrapAPIForError(s.webEditUser))

		rootAPIG.GET("/audit/:entityid", checkParamValidUUID("entityid"), checkIfUserIsAdmin(), WrapAPIForError(s.webGetAuditLog))
	}

	// SSE Endpoint
	rootAPIG.GET("/realtime/", s.requestHasValidAuth(), s.rt.ServeStream)

	// Catch all for Vue SPA
	engine.NoRoute(setXSRFTokenIfNecessary(isCSRFEnabled), func(c *gin.Context) {
		// Handle NoRoute's differently if the path starts with our API endpoint
		if strings.HasPrefix(c.Request.URL.Path, currentAPIVer) {
			c.JSON(http.StatusNotFound, &apiError{Error: "Invalid API route"})
			return
		}
		c.File(filepath.Join(cfg.UserInterface.StaticPath, "index.html"))
	})

	engine.StaticFS("/js/", http.Dir(filepath.Join(cfg.UserInterface.StaticPath, "js")))
	engine.StaticFS("/css/", http.Dir(filepath.Join(cfg.UserInterface.StaticPath, "css")))
	engine.StaticFS("/img/", http.Dir(filepath.Join(cfg.UserInterface.StaticPath, "img")))
	engine.StaticFile("/favicon.ico", filepath.Join(cfg.UserInterface.StaticPath, "favicon.ico"))
	engine.GET("/gocrack-config.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"base_endpoint":        currentAPIVer,
			"server":               "",
			"registration_enabled": s.auth.CanUsersRegister(),
		})
	})

	// END UI ENDPOINTS

	var h http.Handler
	if isCSRFEnabled {
		h = csrf.Protect(
			[]byte(cfg.UserInterface.CSRFKey),
			csrf.RequestHeader("X-Xsrf-Token"),
			csrf.Secure(cfg.Listener.UseSSL),
			csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Error().
					Str("client", r.RemoteAddr).
					Err(csrf.FailureReason(r)).
					Msg("A client failed CSRF protection")

				json.NewEncoder(w).Encode(&apiError{Error: "CSRF Validation Failed"})
			})),
		)(engine)
	} else {
		log.Warn().Msg("CSRF Protection is disabled!")
		h = engine
	}

	return &http.Server{
		Handler: h,
	}
}
