// Package health is responsible for all telling kubernetes the health
// of the apis. If you had db, auth or other systems you would use this
// package to reach out to them and ensure they were working
package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/app"
	"go.uber.org/zap"
	"os"
)

type CheckHandler struct {
	build string
	log   *zap.SugaredLogger
	kube  app.KubeInfo
}

func NewCheckHandler(d *app.Dependencies) *CheckHandler {
	return &CheckHandler{
		build: d.Build,
		log:   d.Logger,
		kube:  d.Kubernetes,
	}
}

// Readiness check the status of the system, but since this is an
// example app we really don't have any systems to check
func (h *CheckHandler) Readiness(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

type SystemStatus struct {
	Status    string `json:"status,omitempty"`
	Build     string `json:"build,omitempty"`
	Host      string `json:"host,omitempty"`
	Pod       string `json:"pod,omitempty"`
	PodIP     string `json:"podIP,omitempty"`
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Liveness report's back stats of the kube cluster
func (h CheckHandler) Liveness(c *fiber.Ctx) error {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	status := SystemStatus{
		Status:    "up",
		Build:     h.build,
		Host:      host,
		Pod:       h.kube.Pod,
		PodIP:     h.kube.PodIP,
		Node:      h.kube.Node,
		Namespace: h.kube.Namespace,
	}

	return c.Status(fiber.StatusOK).JSON(status)
}
