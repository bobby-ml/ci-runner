package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bench-runner/dto"
	"github.com/bench-runner/git"
	docker "github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var pathTemp = "/home/nicolas/go/src/github.com/bench-runner/temp"

func executeJob(job *dto.Jobs) {
	job.UpdateStatus("running")
	err := git.Clone(pathTemp, job.CloneUrl, job.Commit.ID)
	if err != nil {
		//TODO
		print(err)
		return
	}
	if _, err := os.Stat(pathTemp + "/.benchlab.yml"); os.IsNotExist(err) {
		//TODO
		print("FILE DO NOT EXIST")
		return
	}
	b, err := ioutil.ReadFile(pathTemp + "/.benchlab.yml")
	if err != nil {
		fmt.Print(err)
	}
	var bd dto.BenchDefinition
	err = yaml.Unmarshal(b, &bd)
	if err != nil {
		//TODO
		print("FILE DO NOT EXIST")
		return
	}

	cl, _ := docker.NewClientFromEnv()
	l, _ := cl.ListImages(docker.ListImagesOptions{Filter: bd.Image})
	if len(l) == 0 {
		cl.PullImage(docker.PullImageOptions{Repository: bd.Image}, docker.AuthConfiguration{})
	}
	c, err := cl.CreateContainer(docker.CreateContainerOptions{
		Name: job.Commit.ID,
		Config: &docker.Config{
			Image:      bd.Image,
			Cmd:        []string{"/bin/sh", "-c", "while true; do sleep 100000; done"},
			WorkingDir: "/tmp/test",
		},
		HostConfig: &docker.HostConfig{
			Mounts: []docker.HostMount{
				{
					Target: "/var/run/docker.sock",
					Source: "/var/run/docker.sock",
					Type:   "bind",
				},
				{
					Target: "/tmp/test",
					Source: "/home/nicolas/go/src/github.com/bench-runner/temp",
					Type:   "bind",
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	cl.StartContainer(c.ID, nil)
	script := ""
	for _, e := range bd.Script {
		script = script + `echo "########## ` + e + ` ########";` + e + `;`
	}
	de := docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          false,
		Cmd:          []string{"/bin/sh", "-c", script},
		Container:    c.ID,
	}
	var dExec *docker.Exec
	if dExec, err = cl.CreateExec(de); err != nil {
		fmt.Println(err)
		return
	}
	var stdout, stderr bytes.Buffer
	execId := dExec.ID

	var reader = strings.NewReader("echo hello world")
	opts := docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
		InputStream:  reader,
		RawTerminal:  false,
	}
	go func() {
		if err = cl.StartExec(execId, opts); err != nil {
			fmt.Println(err)
			return
		}
	}()
	ticker1 := time.NewTicker(1 * time.Second)
	ticker5 := time.NewTicker(5 * time.Second)
	withError := false
	for {
		quit := false
		select {
		case <-ticker5.C:
			r, err := cl.InspectExec(execId)
			if err != nil {
				panic(err)
			}
			if !r.Running {
				if r.ExitCode != 0 {
					job.UpdateStatus("fail")
					withError = true
				}
				quit = true
			}
		case <-ticker1.C:
			if stdout.Len() > 0 {
				log := stdout.String()
				stdout.Reset()
				logsRow := strings.Split(log, "%%##%%")
				logs := make([]dto.Logs, len(logsRow))
				for i, e := range logsRow {
					logs[i] = dto.Logs{
						Lvl:  "info",
						Data: e,
					}
				}
				job.Log(logs)

			}
			if stderr.Len() > 0 {
				log := stderr.String()
				stderr.Reset()
				logsRow := strings.Split(log, "%%##%%")
				logs := make([]dto.Logs, len(logsRow))
				for i, e := range logsRow {
					logs[i] = dto.Logs{
						Lvl:  "error",
						Data: e,
					}
				}
				job.Log(logs)
			}
		}
		if quit {
			break
		}
	}
	if withError {
		return
	}

	cl.TagImage("result:latest", docker.TagImageOptions{
		Repo: "us.gcr.io/" + job.GcloudProject + "/" + job.Project,
		Tag:  job.JobId,
	})
	PushImage("us.gcr.io/"+job.GcloudProject+"/"+job.Project, job.JobId, job.SA)
	cl.KillContainer(docker.KillContainerOptions{ID: c.ID})
	cl.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID})
	job.UpdateStatus("success")

}

func PushImage(image string, tag string, SA string) (err error) {

	cl, err := docker.NewClientFromEnv()

	err = cl.PushImage(docker.PushImageOptions{
		Name:     image,
		Tag:      tag,
		Registry: "us.gcr.io",
	}, docker.AuthConfiguration{
		ServerAddress: "https://us.gcr.io",
		Username:      "_json_key",
		Password:      SA,
	})
	return
}

func main() {

	ticker1 := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker1.C:
			url := os.Getenv("BENCHLAB_SERVER") + "/jobV1"
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			var input dto.Input
			err = json.Unmarshal(body, &input)
			if err != nil {
				panic(err)
			}
			for _, e := range input.Jobs {
				executeJob(&e)
			}

		}
	}

}
