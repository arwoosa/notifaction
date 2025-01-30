package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/identity/dao"
	"github.com/spf13/viper"
)

type classificationLang struct {
	keys     []string
	data     map[string][]*service.Info
	From     *service.Info
	FromLang string
}

func (c *classificationLang) isEqual(cc *classificationLang) bool {
	if len(c.keys) != len(cc.keys) {
		return false
	}
	for _, k := range c.keys {
		if _, ok := cc.data[k]; !ok {
			return false
		}
	}
	if c.FromLang != cc.FromLang {
		return false
	}
	if !reflect.DeepEqual(c.From, cc.From) {
		return false
	}
	if len(c.data) != len(cc.data) {
		return false
	}

	for k, v := range c.data {
		match := cc.data[k]
		if len(match) != len(v) {
			return false
		}
		for i, vv := range v {
			if !reflect.DeepEqual(vv, match[i]) {
				return false
			}
		}
	}
	return true
}

func newClassficationLang() *classificationLang {
	return &classificationLang{
		keys: []string{},
		data: map[string][]*service.Info{},
	}
}

func (c *classificationLang) GetLangs() []string {
	if c == nil {
		return nil
	}
	return c.keys
}

func (c *classificationLang) GetInfos(lang string) []*service.Info {
	if c == nil {
		return nil
	}
	return c.data[lang]
}

func (c *classificationLang) add(lang string, info *service.Info) {
	if infos, ok := c.data[lang]; ok {
		c.data[lang] = append(infos, info)
	} else {
		c.keys = append(c.keys, lang)
		c.data[lang] = []*service.Info{info}
	}
}

type Identity interface {
	// return notify info and classification by lang
	SubToInfo(from string, to []string) (*classificationLang, error)
	service.Health
}

const (
	identityPath = "/admin/identities"
	healthPath   = "/admin/health/ready"
)

type option func(*identityApi)

func WithHttpClient(httpClient myHttpClient) option {
	return func(i *identityApi) {
		i.httpClient = httpClient
	}
}

func NewIdentity(opts ...option) (Identity, error) {
	url := viper.GetString("identity.url")
	if url == "" {
		return nil, errors.New("identity.url is empty")
	}
	api := &identityApi{
		identityUri: url + identityPath,
		heathUri:    url + healthPath,
	}

	for _, opt := range opts {
		opt(api)
	}

	if api.httpClient == nil {
		api.httpClient = &http.Client{
			Timeout: time.Second * 5,
		}
	}

	return api, nil
}

type identityApi struct {
	httpClient  myHttpClient
	identityUri string
	heathUri    string
}

func (i *identityApi) SubToInfo(from string, to []string) (*classificationLang, error) {
	if len(to) == 0 {
		return nil, nil
	}
	resp, err := i.fetchIdentityData(append(to, from))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	cl := newClassficationLang()

	for _, r := range resp {
		if r.Id == from {
			cl.From = &service.Info{
				Sub:    r.Id,
				Name:   r.Traits.Name,
				Email:  r.Traits.Email,
				Enable: r.State == "active",
			}
			cl.FromLang = r.Traits.Language
			continue
		}
		cl.add(r.Traits.Language, &service.Info{
			Sub:    r.Id,
			Name:   r.Traits.Name,
			Email:  r.Traits.Email,
			Enable: r.State == "active",
		})
	}
	return cl, nil
}

func (i *identityApi) fetchIdentityData(sub []string) (dao.FetchIdentityResponse, error) {

	response := dao.FetchIdentityResponse{}
	// http request to identiy service i.url with params ids = sub and page_size = 100
	params := url.Values{
		"ids":       sub,
		"page_size": {"100"},
	}
	req, err := http.NewRequest("GET", i.identityUri, nil)
	if err != nil {
		return response, fmt.Errorf("failed to create request: %w", err)
	}
	req.URL.RawQuery = params.Encode()

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return response, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// bind response to FetchIdentityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

func (i *identityApi) IsReady() (bool, error) {
	req, err := http.NewRequest("GET", i.heathUri, nil)
	if err != nil {
		return false, err
	}
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
