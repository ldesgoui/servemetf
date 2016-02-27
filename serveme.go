package servemetf

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

type Reservation struct {
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	ServerID    int    `json:"server_id,omitempty"`
	RCON        string `json:"rcon,omitempty"`
	Password    string `json:"password,omitempty"`
	FirstMap    string `json:"first_map,omitempty"`
	WhitelistID int    `json:"whitelist_id,omitempty"`
	ID          int    `json:"id,omitempty"`
	LogSecret   uint64 `json:"logsecret,omitempty"`
	Server      struct {
		Name      string `json:"name"`
		IPAndPort string `json:"ip_and_port"`
	} `json:"server"`
	Errors map[string]interface{} `json:"errors,omitempty"` // errors in response
}

type Server struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Flag      string `json:"flag"`
	IPAndPort string `json:"ip_and_port"`
}

type File struct {
	ID   int    `json:"id"`
	File string `json:"file"`
}

type Response struct {
	Reservation   Reservation       `json:"reservation"`
	Servers       []Server          `json:"servers"`
	ServerConfigs []File            `json:"server_configs"`
	Whitelists    []File            `json:"whitelists"`
	Actions       map[string]string `json:"actions"`
}

type Context struct {
	APIKey string
}

const (
	TimeFormat = "2006-01-02T15:04:05.999-07:00"
)

var (
	ErrAlreadyReserved = errors.New("server has already been reserved")
	ErrNotFound        = errors.New("server not found")
)

func (c Context) newRequest(data interface{}, method, url string) *http.Request {
	json, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	return req
}

func (c Context) GetReservationTime() (starts time.Time, ends time.Time, err error) {
	req := c.newRequest(nil, "GET", "http://serveme.tf/api/reservations/new?api_key="+c.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	dec := json.NewDecoder(resp.Body)
	var jsonresp Response
	err = dec.Decode(&jsonresp)
	if err != nil {
		return
	}

	starts, err = time.Parse(TimeFormat, jsonresp.Reservation.StartsAt)
	ends, err = time.Parse(TimeFormat, jsonresp.Reservation.EndsAt)
	return
}

func (c Context) FindServers(starts, ends time.Time) (Response, error) {
	reservation := Reservation{
		StartsAt: starts.Format(TimeFormat),
		EndsAt:   ends.Format(TimeFormat),
	}

	req := c.newRequest(struct {
		Reservation Reservation `json:"reservation"`
	}{reservation}, "POST", "http://serveme.tf/api/reservations/find_servers?api_key="+c.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(resp)
		return Response{}, err
	}

	dec := json.NewDecoder(resp.Body)
	var jsonresp Response
	err = dec.Decode(&jsonresp)
	return jsonresp, err
}

func (c Context) Create(reservation Reservation) (Response, error) {
	var response Response
	req := c.newRequest(reservation, "POST", "http://serveme.tf/api/reservations?api_key="+c.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	if resp.StatusCode == 400 {
		return response, ErrAlreadyReserved
	} else if resp.StatusCode == 400 {
		return response, ErrNotFound
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (c Context) Delete(id int) error {
	str := strconv.Itoa(id)
	req := c.newRequest(nil, "DELETE", "http://serveme.tf/api/reservations/"+str+"?api_key"+c.APIKey)
	_, err := client.Do(req)
	return err
}
