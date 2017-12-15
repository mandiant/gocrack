package web

import (
	"net/http"

	"github.com/fireeye/gocrack/shared"

	"github.com/gin-gonic/gin"
)

type WorkerVersionInfo struct {
	Hostname      string               `json:"hostname"`
	WorkerVersion string               `json:"version"`
	Engines       shared.EngineVersion `json:"engines"`
}

type ServerVersionInfo struct {
	Workers           []WorkerVersionInfo `json:"workers"`
	ServerVersion     string              `json:"version"`
	ServerCompileTime string              `json:"compile_time"`
}

func (s *Server) webGetVersion(c *gin.Context) *WebAPIError {
	workers := s.wmgr.GetCurrentWorkers()

	resp := ServerVersionInfo{
		Workers:           make([]WorkerVersionInfo, len(workers)),
		ServerCompileTime: serverCompileTime,
		ServerVersion:     serverVersion,
	}

	i := 0
	for hostname, worker := range workers {
		resp.Workers[i] = WorkerVersionInfo{
			Hostname:      hostname,
			Engines:       worker.LastBeacon.Engines,
			WorkerVersion: worker.LastBeacon.WorkerVersion,
		}
		i++
	}

	c.JSON(http.StatusOK, &resp)
	return nil
}
