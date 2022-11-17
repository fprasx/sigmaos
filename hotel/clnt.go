package hotel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func webRequest(url string, vals url.Values) ([]byte, error) {
	resp, err := http.PostForm(url, vals)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v %s", resp.StatusCode, body)
	}
	return body, nil
}

func WebLogin(u, p string) (string, error) {
	vals := url.Values{}
	vals.Set("username", u)
	vals.Set("password", p)
	body, err := webRequest("http://localhost:8090/user", vals)
	if err != nil {
		return "", err
	}
	repl := make(map[string]interface{})
	err = json.Unmarshal(body, &repl)
	if err != nil {
		return "", err
	}
	return repl["message"].(string), nil
}

func WebSearch(inDate, outDate string, lat, lon float64) error {
	vals := url.Values{}
	vals.Set("inDate", inDate)
	vals.Set("outDate", outDate)
	vals.Add("lat", fmt.Sprintf("%f", lat))
	vals.Add("lon", fmt.Sprintf("%f", lon))
	body, err := webRequest("http://localhost:8090/hotels", vals)
	if err != nil {
		return err
	}
	log.Printf("%v", string(body))
	return nil
}

func WebRecs(require string, lat, lon float64) error {
	vals := url.Values{}
	vals.Set("require", require)
	vals.Add("lat", fmt.Sprintf("%f", lat))
	vals.Add("lon", fmt.Sprintf("%f", lon))
	body, err := webRequest("http://localhost:8090/recommendations", vals)
	if err != nil {
		return err
	}
	log.Printf("%v", string(body))
	return nil
}

func WebReserve(inDate, outDate string, lat, lon float64, hotelid, name, u, p string, n int) (string, error) {
	vals := url.Values{}
	vals.Set("inDate", inDate)
	vals.Set("outDate", outDate)
	vals.Set("lat", fmt.Sprintf("%f", lat))
	vals.Set("lon", fmt.Sprintf("%f", lon))
	vals.Set("hotelId", hotelid)
	vals.Set("customername", name)
	vals.Set("username", u)
	vals.Set("password", p)
	vals.Set("number", fmt.Sprintf("%d", n))
	body, err := webRequest("http://localhost:8090/reservation", vals)
	if err != nil {
		return "", err
	}
	repl := make(map[string]interface{})
	err = json.Unmarshal(body, &repl)
	if err != nil {
		return "", err
	}
	return repl["message"].(string), nil
}