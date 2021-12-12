package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"
	"github.com/skillz/opvic/agent/api/v1alpha1"
	"github.com/skillz/opvic/utils"
	"gopkg.in/yaml.v2"
)

const indexPath string = "index.yaml"

type ChartVersion struct {
	Version string `yaml:"version"`
}
type Index struct {
	APIVersion string `yaml:"apiVersion"`
	Entries    map[string][]*ChartVersion
}

type Provider struct {
	client *http.Client
	cache  *cache.Cache
	log    logr.Logger
}

func NewProvider(cache *cache.Cache, logger logr.Logger) *Provider {
	return &Provider{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  cache,
		log:    logger,
	}
}

func (p *Provider) GetCacheValue(key string) (interface{}, bool) {
	return p.cache.Get(key)
}

func (p *Provider) SetCacheValue(key string, value interface{}) {
	p.cache.Set(key, value, cache.DefaultExpiration)
}

func ReleasesCacheKey(repo string) string {
	// drop the https://
	return fmt.Sprintf("helm/%s", repo[8:])
}

func AppendIndex(repo string) string {
	return fmt.Sprintf("%s/%s", repo, indexPath)
}

func (p *Provider) GetIndex(repo string) (*Index, error) {
	log := p.log.WithValues("repo", repo)
	indexCache, ok := p.GetCacheValue(ReleasesCacheKey(repo))
	if ok {
		log.Info("found index in cache")
		return indexCache.(*Index), nil
	}
	log.Info("getting index from remote")
	url := AppendIndex(repo)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	index, err := LoadIndex(data)
	if err != nil {
		return nil, err
	}
	p.SetCacheValue(ReleasesCacheKey(repo), index)
	return index, nil
}

func LoadIndex(data []byte) (*Index, error) {
	i := &Index{}
	if len(data) == 0 {
		return i, fmt.Errorf("%s is empty", indexPath)
	}
	if err := yaml.Unmarshal(data, i); err != nil {
		return i, err
	}
	return i, nil
}

func (p *Provider) GetVersions(conf v1alpha1.RemoteVersion) ([]string, error) {
	var matchedVersions []string
	var versions []string
	index, err := p.GetIndex(conf.Repo)
	if err != nil {
		return versions, err
	}
	chartVersions := index.Entries[conf.Chart]
	if len(chartVersions) == 0 {
		return versions, nil
	}
	for _, chartVersion := range chartVersions {
		matched, v := utils.MatchPattern(conf.Extraction.Regex.Pattern, conf.Extraction.Regex.Result, chartVersion.Version)
		if matched {
			matchedVersions = append(matchedVersions, v)
		}
	}
	if conf.Constraint == "" {
		return matchedVersions, nil
	} else {
		for _, version := range matchedVersions {
			meet, err := utils.MeetConstraint(conf.Constraint, version)
			if err != nil {
				return nil, err
			}
			if meet {
				versions = append(versions, version)
			}
		}
	}
	return versions, nil
}
