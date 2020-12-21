package proxmox

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
)

func mockJSONResponder(status int, body string) httpmock.Responder {
	response := httpmock.NewStringResponse(status, body)
	response.Header.Set("Content-Type", "application/json")
	return httpmock.ResponderFromResponse(response)
}

func TestNewSession(t *testing.T) {
	goodURL, err := url.Parse("https://localhost:8006/api2/json")
	if err != nil {
		t.Error(err)
		return
	}
	goodCredentials := map[string]string{
		"username": "root@pam",
		"password": "root",
	}
	badCredentials := map[string]string{
		"username": "fail@pam",
		"password": "fail",
	}
	goodResponder := mockJSONResponder(200, `{"data":{"ticket":"PVE:root@pam:5FDFEFAA::VOz9e34ULtm1h+Sv3ZrP+Cp1vn99bQvLFJT61JxKf9tSsZcPli76YICIK1xKnQLT2d/FnhY8XBgU4owTXFuCN22BxzqGAkEz2V0Q0eJt18hPLHy1MkLYj4IxCL6pkPmfzDNZzIe4cn1ShjwWGzpYS3JjNdSHiVh3n8tiswOrCPDZMZRV4LrRHe7LamAxAS7sLEqjprxMG1InbRU/RrG3I5gEGrtGrmm3rRrwCbI4Ve7/hkWlDL8x37G2FwtD9D7YaMl3BiMZZDPIp5JI9734CX13OD45iGJsx6br0DuZ54B23/4bsYCa0H0M2bQy29VVNSVgfdS/H55WngEWorDtOQ==","username":"root@pam","cap":{"vms":{"VM.PowerMgmt":1,"VM.Migrate":1,"VM.Allocate":1,"VM.Clone":1,"VM.Config.CDROM":1,"Permissions.Modify":1,"VM.Config.Network":1,"VM.Config.Cloudinit":1,"VM.Monitor":1,"VM.Snapshot.Rollback":1,"VM.Console":1,"VM.Config.HWType":1,"VM.Config.Options":1,"VM.Backup":1,"VM.Snapshot":1,"VM.Config.CPU":1,"VM.Audit":1,"VM.Config.Disk":1,"VM.Config.Memory":1},"access":{"Group.Allocate":1,"User.Modify":1,"Permissions.Modify":1},"storage":{"Datastore.Allocate":1,"Permissions.Modify":1,"Datastore.AllocateTemplate":1,"Datastore.AllocateSpace":1,"Datastore.Audit":1},"sdn":{"SDN.Audit":1,"Permissions.Modify":1,"SDN.Allocate":1},"dc":{"SDN.Audit":1,"Sys.Audit":1,"SDN.Allocate":1},"nodes":{"Sys.Modify":1,"Sys.Console":1,"Sys.Audit":1,"Permissions.Modify":1,"Sys.Syslog":1,"Sys.PowerMgmt":1}},"CSRFPreventionToken":"5FDFEFAA:F4MJATbjo6mYMtlPx3vko043K+kijnVpU7tTbGX2Bm8"}}`)
	type args struct {
		restyClient *resty.Client
		apiURL      *url.URL
		credentials QueryParams
		method      string
		url         string
		responder   httpmock.Responder
	}
	tests := []struct {
		name    string
		args    args
		want    *Session
		wantErr bool
	}{
		{
			name: "Good Credentials",
			args: args{
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
				apiURL:      goodURL,
				credentials: goodCredentials,
				method:      "POST",
				url:         goodURL.String() + "/access/ticket",
				responder:   goodResponder,
			},
			want: &Session{
				BaseURL:             goodURL,
				CSRFPreventionToken: `5FDFEFAA:F4MJATbjo6mYMtlPx3vko043K+kijnVpU7tTbGX2Bm8`,
				Ticket: &http.Cookie{
					Name:  "PVEAuthCookie",
					Value: `PVE:root@pam:5FDFEFAA::VOz9e34ULtm1h+Sv3ZrP+Cp1vn99bQvLFJT61JxKf9tSsZcPli76YICIK1xKnQLT2d/FnhY8XBgU4owTXFuCN22BxzqGAkEz2V0Q0eJt18hPLHy1MkLYj4IxCL6pkPmfzDNZzIe4cn1ShjwWGzpYS3JjNdSHiVh3n8tiswOrCPDZMZRV4LrRHe7LamAxAS7sLEqjprxMG1InbRU/RrG3I5gEGrtGrmm3rRrwCbI4Ve7/hkWlDL8x37G2FwtD9D7YaMl3BiMZZDPIp5JI9734CX13OD45iGJsx6br0DuZ54B23/4bsYCa0H0M2bQy29VVNSVgfdS/H55WngEWorDtOQ==`,
				},
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
			},
			wantErr: false,
		},
		{
			name: "Bad Credentials",
			args: args{
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
				apiURL:      goodURL,
				credentials: badCredentials,
				method:      "POST",
				url:         goodURL.String() + "/access/ticket",
				responder:   mockJSONResponder(401, `{"data":null}`),
			},
			want: &Session{
				BaseURL:             goodURL,
				CSRFPreventionToken: "",
				Ticket: &http.Cookie{
					Name:  "PVEAuthCookie",
					Value: "",
				},
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
			},
			wantErr: true,
		},
		{
			name: "Connection Failure",
			args: args{
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
				apiURL:      goodURL,
				credentials: goodCredentials,
				method:      "POST",
				url:         goodURL.String() + "/errorerrorerror",
				responder:   goodResponder,
			},
			want: &Session{
				BaseURL:             goodURL,
				CSRFPreventionToken: "",
				Ticket: &http.Cookie{
					Name:  "PVEAuthCookie",
					Value: "",
				},
				restyClient: resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.ActivateNonDefault(tt.args.restyClient.GetClient())
			defer httpmock.Deactivate()
			httpmock.RegisterResponder(tt.args.method, tt.args.url, tt.args.responder)

			got, err := NewSession(tt.args.restyClient, tt.args.apiURL, tt.args.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSession() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if got.CSRFPreventionToken != tt.want.CSRFPreventionToken {
				t.Errorf("got.CSRFPreventionToken != tt.want.CSRFPreventionToken\n%#v != %#v", got.CSRFPreventionToken, tt.want.CSRFPreventionToken)
			}
			if got.Ticket.Value != tt.want.Ticket.Value {
				t.Errorf("got.Ticket.Value != tt.want.Ticket.Value\n%#v != %#v", got.Ticket.Value, tt.want.Ticket.Value)
			}
		})
	}
}
