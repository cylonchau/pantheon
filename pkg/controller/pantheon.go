package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cylonchau/pantheon/pkg/api/target"
	"github.com/cylonchau/pantheon/pkg/model"
)

type PantheonClient struct {
	ServerURL   string
	ClusterName string
	AuthToken   string
}

func NewPantheonClient(serverURL, clusterName, authToken string) *PantheonClient {
	url := strings.TrimSuffix(serverURL, "/")
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	return &PantheonClient{
		ServerURL:   url,
		ClusterName: clusterName,
		AuthToken:   authToken,
	}
}

func (c *PantheonClient) FetchMonitorRules() ([]model.MonitorRule, error) {
	url := fmt.Sprintf("%s/ph/v1/monitors", c.ServerURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rules []model.MonitorRule
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (c *PantheonClient) FetchRegisteredTargets() ([]target.TargetList, error) {
	url := fmt.Sprintf("%s/ph/v1/targets/cmd/kubernetes_cluster/%s", c.ServerURL, c.ClusterName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var targets []target.TargetList
	if err := json.Unmarshal(bodyBytes, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

func (c *PantheonClient) RegisterTarget(ruleName string, item target.TargetItem) error {
	payload := target.Target{
		InstanceSelector: map[string]string{
			"kubernetes_cluster": c.ClusterName,
			"monitor_rule":       ruleName,
		},
		Targets: []target.TargetItem{item},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/ph/v1/targets", c.ServerURL)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *PantheonClient) DeleteTarget(id uint) error {
	url := fmt.Sprintf("%s/ph/v1/targets/%d", c.ServerURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
