package api

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DiscoveredService struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}

type ApiServer struct {
	client client.Client
}

func NewApiServer(c client.Client) *ApiServer {
	return &ApiServer{client: c}
}

func (s *ApiServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/services", s.handleListServices)
	return mux
}

func (s *ApiServer) handleListServices(w http.ResponseWriter, r *http.Request) {
	var serviceList corev1.ServiceList

	err := s.client.List(context.Background(), &serviceList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]DiscoveredService, 0, len(serviceList.Items))
	for _, svc := range serviceList.Items {
		var ips []string

		// Extract IPs from status.loadBalancer.ingress
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				ips = append(ips, ingress.IP)
			}
		}

		if len(ips) > 0 {
			response = append(response, DiscoveredService{
				Name:      svc.Name,
				Endpoints: ips,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
