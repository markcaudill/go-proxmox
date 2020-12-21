package proxmox

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Session stores the API base URL as well as the necessary session identifiers.
type Session struct {
	BaseURL             *url.URL
	CSRFPreventionToken string
	Ticket              *http.Cookie
	restyClient         *resty.Client
}

// JSONResponse is used to unmarshall JSON data
type JSONResponse map[string]interface{}

// QueryParams is used to store HTTP query parameters
type QueryParams map[string]string

// AccessTicketResponseData is used to unmarshal the JSON response from /access/ticket
type AccessTicketResponseData struct {
	Data struct {
		Username            string      `json:"username"`
		Ticket              string      `json:"ticket"`
		CSRFPreventionToken string      `json:"CSRFPreventionToken"`
		Cap                 interface{} `json:"cap"` // We don't care about "cap" for Session purposes
	} `json:"data"`
}

// NewSession creates a new Session struct by authenticating
//
// Example:
//		var credentials = map[string]string{
//			"username": "root@pam",
//			"password": "root",
//		}
//		apiURL, err := url.Parse("https://127.0.0.1:8006/api2/json")
//		if err != nil {
//			log.Fatal("unable to parse URL: ", err)
//		}
//		client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
//		s, err := proxmox.NewSession(client, apiURL, credentials)
//		if err != nil {
//			log.Fatal(err)
//		}
func NewSession(restyClient *resty.Client, apiURL *url.URL, credentials QueryParams) (*Session, error) {
	session := &Session{
		BaseURL:     apiURL,
		restyClient: restyClient,
	}
	var responseData AccessTicketResponseData
	req := session.NewRequest(credentials, &responseData)
	url := apiURL.String() + "/access/ticket"
	res, err := req.Post(url)
	if err != nil {
		return nil, err
	}

	_, ok := res.Result().(*AccessTicketResponseData)
	if !ok {
		return nil, fmt.Errorf("Result not parseable: %#v", responseData)
	}

	session.CSRFPreventionToken = responseData.Data.CSRFPreventionToken
	session.Ticket = &http.Cookie{
		Name:  "PVEAuthCookie",
		Value: responseData.Data.Ticket,
	}

	if res.StatusCode() != http.StatusOK {
		return session, fmt.Errorf(res.Status())
	}
	return session, nil
}

// NewRequest creates a new resty.Request object configured with the current Session variables (e.g. cookies, etc.)
func (s *Session) NewRequest(params QueryParams, result interface{}) *resty.Request {
	req := s.restyClient.R()
	if s.Ticket != nil {
		req = req.SetCookie(s.Ticket)
	}
	req = req.SetQueryParams(params)
	req = req.SetResult(result)
	return req
}

// Do performs an HTTP request. All params get passed as query params.
//
// Example:
//		var credentials = map[string]string{
//			"username": "root@pam",
//			"password": "root",
//		}
//		res, err := session.Do("POST", "/access/ticket", credentials)
func (s *Session) Do(method string, path string, params QueryParams) (*JSONResponse, error) {
	var responseData JSONResponse
	var res *resty.Response
	var err error
	var url = s.BaseURL.String() + path

	method = strings.ToUpper(method)
	req := s.NewRequest(params, &responseData)
	switch method {
	case "GET":
		res, err = req.Get(url)
	case "POST":
		req = req.SetHeader("CSRFPreventionToken", s.CSRFPreventionToken)
		res, err = req.Post(url)
	case "PUT":
		req = req.SetHeader("CSRFPreventionToken", s.CSRFPreventionToken)
		res, err = req.Put(url)
	case "DELETE":
		req = req.SetHeader("CSRFPreventionToken", s.CSRFPreventionToken)
		res, err = req.Delete(url)
	default:
		return nil, fmt.Errorf("Unsupported method: %s", method)
	}
	if err != nil {
		return nil, err
	}

	_, ok := res.Result().(*JSONResponse)
	if !ok {
		return nil, fmt.Errorf("Unable to parse response as JSONResponse: %#v", responseData)
	}

	return &responseData, err
}
