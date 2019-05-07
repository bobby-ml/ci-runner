package dto

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type BenchDefinition struct {
	Image  string   `yaml:"image"`
	Script []string `yaml:"script"`
}

type Input struct {
	Jobs []Jobs `json:"jobs"`
}
type Committer struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}
type Author struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}
type Commit struct {
	Added     []interface{} `json:"added"`
	Removed   []interface{} `json:"removed"`
	ID        string        `json:"id"`
	Distinct  bool          `json:"distinct"`
	Timestamp string        `json:"timestamp"`
	Modified  []string      `json:"modified"`
	Committer Committer     `json:"committer"`
	TreeID    string        `json:"tree_id"`
	Author    Author        `json:"author"`
	URL       string        `json:"url"`
	Message   string        `json:"message"`
}

type Jobs struct {
	Project       string `json:"project"`
	Commit        Commit `json:"commit"`
	OwnerUID      string `json:"owner_uid"`
	Status        string `json:"status"`
	Compare       string `json:"compare"`
	Branch        string `json:"branch"`
	CloneUrl      string `json:"clone_url"`
	JobId         string `json:"job_id"`
	SA            string `json:"sa"`
	GcloudProject string `json:"gcloud_project"`
}

type Logs struct {
	Lvl  string `json:"lvl"`
	Data string `json:"data"`
}

type JobUpdate struct {
	JobId  string `json:"job_id"`
	Logs   []Logs `json:"logs"`
	Status string `json:"status"`
}

func (s Jobs) Log(data []Logs) {
	ju := JobUpdate{
		JobId: s.JobId,
		Logs:  data,
	}
	b, err := json.Marshal(ju)
	req, err := http.NewRequest(http.MethodPut, os.Getenv("BENCHLAB_SERVER")+"/jobV1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
func (s Jobs) UpdateStatus(status string) {
	ju := JobUpdate{
		JobId:  s.JobId,
		Status: status,
	}
	b, err := json.Marshal(ju)
	req, err := http.NewRequest(http.MethodPut, os.Getenv("BENCHLAB_SERVER")+"/jobV1", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
