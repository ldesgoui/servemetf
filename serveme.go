package servemetf

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

type Reservation struct {
	Status      string `json:"status"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	ServerID    int    `json:"server_id,omitempty"`
	RCON        string `json:"rcon,omitempty"`
	Password    string `json:"password,omitempty"`
	FirstMap    string `json:"first_map,omitempty"`
	WhitelistID int    `json:"whitelist_id,omitempty"`
	ID          int    `json:"id,omitempty"`
	LogSecret   string `json:"logsecret,omitempty"`
	Ended       bool   `json:"ended"`
	ZipFileURL  string `json:"zipfile_url"`
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
	Host   string
	APIKey string
}

const (
	TimeFormat = "2006-01-02T15:04:05.999-07:00"
)

var (
	ErrAlreadyReserved = errors.New("serveme: you have already reserved a server")
	ErrNotFound        = errors.New("serveme: server not found")
)

func (c Context) newRequest(data interface{}, method, url string) *http.Request {
	json, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	return req
}

// ID is reservation id
func (c Context) Status(id int, steamID string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/" + strconv.FormatUint(uint64(id), 10),
	}

	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	req := c.newRequest(nil, "GET", u.String())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	var jsonresp Response
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	if err != nil {
		return "", err
	}

	return jsonresp.Reservation.Status, nil
}

func (c Context) GetReservationTime(steamID string) (starts time.Time, ends time.Time, err error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/new",
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	req := c.newRequest(nil, "GET", u.String())
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	var jsonresp Response
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	if err != nil {
		return
	}

	starts, err = time.Parse(TimeFormat, jsonresp.Reservation.StartsAt)
	ends, err = time.Parse(TimeFormat, jsonresp.Reservation.EndsAt)
	return
}

func (c Context) FindServers(starts, ends time.Time, steamID string) (Response, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/find_servers",
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	reservation := Reservation{
		StartsAt: starts.Format(TimeFormat),
		EndsAt:   ends.Format(TimeFormat),
	}

	req := c.newRequest(struct {
		Reservation Reservation `json:"reservation"`
	}{reservation}, "POST", u.String())

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}

	var jsonresp Response

	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	return jsonresp, err
}

func (c Context) Create(reservation Reservation, steamID string) (Response, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations",
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	var response Response
	req := c.newRequest(reservation, "POST", u.String())

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	if resp.StatusCode == 400 {
		return response, ErrAlreadyReserved
	} else if resp.StatusCode == 404 {
		return response, ErrNotFound
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (c Context) Delete(id int, steamID string) error {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/" + strconv.Itoa(id),
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	req := c.newRequest(nil, "DELETE", u.String())
	_, err := client.Do(req)
	return err
}

func (c Context) Ended(id int, steamID string) (bool, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/" + strconv.Itoa(id),
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	req := c.newRequest(nil, "GET", u.String())
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	var jsonresp Response
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)
	if err != nil {
		return false, err
	}

	return jsonresp.Reservation.Ended || jsonresp.Reservation.Status == "ended", nil
}

func (c Context) GetZipFileURL(id int, steamID string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "api/reservations/" + strconv.Itoa(id),
	}
	values := u.Query()
	values.Set("api_key", c.APIKey)
	values.Set("steam_uid", steamID)
	u.RawQuery = values.Encode()

	req := c.newRequest(nil, "GET", u.String())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var jsonresp Response
	err = json.NewDecoder(resp.Body).Decode(&jsonresp)

	return jsonresp.Reservation.ZipFileURL, err
}
