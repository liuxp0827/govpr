package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/liuxp0827/govpr/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var appid string = "1474614455854954882"
var appkey string = "VjZ5WQDpWQZ3WjF3WQw3WQF6VzCoDF=="
var host string = "http://207.226.247.222:6060"
var userid string = "test123"
var waveFile, content string
var step int
var ops string
var help bool

func init() {
	flag.StringVar(&userid, "u", "test123", "userid")
	flag.IntVar(&step, "step", -1, "train step 1~5, effective in 'addsample' operation")
	flag.StringVar(&ops, "op", "", "operation [ registeruser | deleteuser | detectquery | detectregister | addsample | trainmodel | verifymodel | deletemodel ]")
	flag.StringVar(&waveFile, "wav", "", "wave file")
	flag.StringVar(&content, "ct", "", "content, effective in 'addsample' and 'verifymodel' operation")
	flag.BoolVar(&help, "h", false, "help bool default false")
}

func usage() {
	flag.PrintDefaults()
	os.Exit(0)
}

func token() string {
	str := fmt.Sprintf("%s&%s", appid, appkey)
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if help {
		usage()
	}

	if userid == "" {
		log.Fatal("userid can not be \"\"")
	}

	client := &http.Client{}
	var err error
	var req *http.Request

	log.Infof("token: %s", token())

	switch ops {
	case "registeruser":
		req, err = registerUser(userid, token())

	case "deleteuser":
		req, err = deteleUser(userid, token())

	case "detectquery":
		req, err = detectquery(userid, token())

	case "detectregister":
		req, err = detectregister(userid, token())

	case "addsample":
		if waveFile == "" || !strings.HasSuffix(waveFile, ".wav") {
			log.Fatalf("wave file %s invalid", waveFile)
		}
		if len(content) != 8 {
			log.Fatalf("content %s invalid", content)
		}
		if step < 1 || step > 5 {
			log.Fatalf("step %d invalid", step)
		}

		req, err = addsample(userid, token(), waveFile, content, strconv.Itoa(step))

	case "trainmodel":
		req, err = trainmodel(userid, token())

	case "deletemodel":
		req, err = deletemodel(userid, token())

	case "verifymodel":

		if waveFile == "" || !strings.HasSuffix(waveFile, ".wav") {
			log.Fatalf("wave file %s invalid", waveFile)
		}
		if len(content) != 8 {
			log.Fatalf("content %s invalid", content)
		}

		req, err = verifymodel(userid, token(), waveFile, content)
		//for i := 0; i < 100; i++ {
		//	go func() {
		//		benchverifymodel(userid, token(), waveFile, content)
		//	}()
		//}

		//select {}

	default:
		log.Fatalf("ops %s invalid", ops)
	}

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(string(data))
}

func registerUser(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/registeruser", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func deteleUser(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/deleteuser", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func detectquery(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/detectquery", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func detectregister(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/detectregister", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func trainmodel(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/trainmodel", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func deletemodel(userid, token string) (*http.Request, error) {
	req, err := http.NewRequest("POST", host+"/deletemodel", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)

	req.URL.RawQuery = query.Encode()

	return req, nil
}

func benchverifymodel(userid, token, path, content string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", path)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		log.Error(err)
		return err
	}
	req, err := http.NewRequest("POST", host+"/verifymodel", body)

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)
	query.Add("content", content)

	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info(string(data))

	return err
}

func verifymodel(userid, token, path, content string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", host+"/verifymodel", body)

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)
	query.Add("content", content)

	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, err
}

func addsample(userid, token, path, content, step string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", host+"/addsample", body)

	query := req.URL.Query()
	query.Add("userid", userid)
	query.Add("token", token)
	query.Add("content", content)
	query.Add("step", step)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
