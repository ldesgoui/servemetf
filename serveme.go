package servemetf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 5 * time.Second}

type Reservation struct {
	StartsAt  string                 `json:"starts_at"`
	EndsAt    string                 `json:"ends_at"`
	ServerID  int                    `json:"server_id,omitempty"`
	Password  string                 `json:"password,omitempty"`
	RCON      string                 `json:"rcon,omitempty"`
	FirstMap  string                 `json:"first_map,,omitempty"`
	ID        int                    `json:"id,omitempty"`
	LogSecret uint64                 `json:"logsecret,omitempty"`
	Errors    map[string]interface{} `json:"errors,omitempty"` // errors in response
}

type Server struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

type File struct {
	ID   int    `json:"id"`
	File string `json:"file"`
}

type Response struct {
	Reservation   Reservation `json:"reservation"`
	Servers       []Server    `json:"servers"`
	ServerConfigs []File      `json:"server_configs"`
	Whitelists    []File      `json:"whitelists"`
}

type Context struct {
	APIKey string
}

const (
	TimeFormat = "2006-02-01T15:04:05.999+07:00"
)

func (c Context) URL() string {
	return fmt.Sprintf("http://serveme.tf/api/reservations/new?api_key=" + c.APIKey)
}

func (c Context) newRequest(data interface{}) *http.Request {
	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)
	enc.Encode(data)

	req, _ := http.NewRequest("POST", c.URL(), buffer)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (c Context) FindServers(starts, ends time.Time) (Response, error) {
	reservation := Reservation{
		StartsAt: starts.Format(TimeFormat),
		EndsAt:   ends.Format(TimeFormat),
	}

	req := c.newRequest(struct {
		Reservation Reservation       `json:"reservation"`
		Actions     map[string]string `json:"actions"`
	}{reservation, map[string]string{
		"find_servers": "http://serveme.tf/api/reservations/find_servers",
	}})

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}

	dec := json.NewDecoder(resp.Body)
	var jsonresp Response
	err = dec.Decode(&jsonresp)
	return jsonresp, err
}

func (c Context) Create(reservation Reservation) (Response, error) {
	var response Response
	req := c.newRequest(reservation)

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}
