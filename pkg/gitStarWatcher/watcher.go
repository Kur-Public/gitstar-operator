package gitStarWatcher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"gitstar-operator/pkg/apis"
	customV1 "gitstar-operator/pkg/apis/app/v1"
)

var log = logf.Log.WithName("controller_gitstar")

func Run() {
	c := newK8SClient()
	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				gitStarList, err := ListGitStarObj(c)
				if err != nil {
					log.Error(err, "list gitstar apiObject failed! ")
					break
				}
				for _, gitStar := range gitStarList.Items {
					starNumber, err := GetStarOfRepo(gitStar)
					if err != nil {
						log.Error(err, "get star number of repo failed! ")
						gitStar.Status = customV1.GitStarStatus{
							StarNumber:   -1,
							UpdatedAt:    time.Now(),
							FailedReason: err.Error(),
						}
					} else {
						gitStar.Status = customV1.GitStarStatus{
							StarNumber:   starNumber,
							UpdatedAt:    time.Now(),
							FailedReason: "",
						}
					}

					err = UpdateGitStarObj(c, &gitStar)
					if err != nil {
						log.Error(err, "get star number of repo failed! ")
						break
					}
					log.Info(fmt.Sprintf("update repo '%s', star number: '%d'", gitStar.Spec.RepoName, gitStar.Status.StarNumber))
				}
			}
		}
		log.Info("update gitStar success \n")
	}()
}

func ListGitStarObj(c client.Client) (*customV1.GitStarList, error) {
	list := &customV1.GitStarList{}
	err := c.List(context.TODO(), list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func GetStarOfRepo(gitStar customV1.GitStar) (int64, error) {
	split := strings.Split(gitStar.Spec.RepoName, "/")
	if len(split) != 2 {
		return -1, errors.New("The repo name is invalid, please check! ")
	}

	c := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "8073e8e8dfa9dd998b35840ddfd8f12e9bf60a7b"},
	)))

	get, _, err := c.Repositories.Get(context.TODO(), split[0], split[1])
	if err != nil {
		return -1, err
	} else if get == nil {
		return -1, errors.New(fmt.Sprintf("repo : '%s' not found , please check! ", gitStar.Spec.RepoName))
	} else if get.StargazersCount == nil {
		return -1, errors.New(fmt.Sprintf("repo : '%s' star number is nil , please check! ", gitStar.Spec.RepoName))
	}

	return int64(*get.StargazersCount), nil
}

func UpdateGitStarObj(c client.Client, gitStar *customV1.GitStar) error {
	return c.Status().Update(context.TODO(), gitStar)
}

func newK8SClient() client.Client {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/home/kurisu/.kube/config")
	// cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "get kubernetes config in cluster was failed! ")
	}
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	scheme := scheme.Scheme
	apis.AddToScheme(scheme)
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		log.Error(err, "new kubernetes client for config was failed ! ")
	}
	return c
}
