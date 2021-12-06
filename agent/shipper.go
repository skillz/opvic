package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	controlplane "github.com/skillz/opvic/controlplane/api/v1alpha1"
)

type ShipperConfig struct {
	URL       string
	Token     string
	TLSVerify bool
	Timeout   time.Duration
}

type Shipper struct {
	Client    *http.Client
	BaseURL   string
	AuthToken string
}

func NewShipper(config *ShipperConfig) *Shipper {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !config.TLSVerify,
		},
	}
	if config.Timeout > 0 {
		tr.DialContext = (&net.Dialer{
			Timeout: config.Timeout,
		}).DialContext
	}
	return &Shipper{
		Client: &http.Client{
			Transport: tr,
		},
		BaseURL:   config.URL,
		AuthToken: config.Token,
	}
}

func (s *Shipper) Post(payload controlplane.AgentPayload) error {
	agentsEndpoint := fmt.Sprintf("%s%s", s.BaseURL, controlplane.AgentsAPIEndpoint)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", agentsEndpoint, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AuthToken))
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusAlreadyReported {
		return fmt.Errorf("unexpected status code: %d status: %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func (r *VersionTrackerReconciler) ShipToControlPlane(ver SubjectVersion) error {
	log := r.Log
	log.Info("Sending AppVersion", "identifier", ver.ID)
	shipperConf := &ShipperConfig{
		URL:       r.Config.ControlPlaneUrl,
		Token:     r.Config.ControlPlaneAuthToken,
		Timeout:   time.Second * 30,
		TLSVerify: false,
	}
	shipper := NewShipper(shipperConf)
	err := shipper.Post(r.PrepareThePayload(ver))
	if err != nil {
		return err
	}
	return nil
}

func (r *VersionTrackerReconciler) PrepareThePayload(sv SubjectVersion) controlplane.AgentPayload {
	payload := controlplane.AgentPayload{}
	payload.AgentID = r.Config.ID
	payload.AgentTags = r.Config.Tags
	vers := []controlplane.Version{}
	for _, v := range sv.Versions {
		vers = append(vers, controlplane.Version{
			RunningVersion: v.Version,
			ResourceCount:  v.ResourceCount,
			ResourceKind:   v.ResourceKind,
			ExtractedFrom:  v.ExtractedFrom,
		})
	}
	payload.Version = controlplane.SubjectVersion{
		ID:              sv.ID,
		NameSpace:       sv.NameSpace,
		RunningVersions: sv.UniqVersions,
		ResourceCount:   sv.TotalResourceCount,
		Versions:        vers,
		RemoteVersion:   sv.RemoteVersion,
	}
	return payload
}
