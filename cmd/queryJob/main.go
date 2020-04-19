package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"gitstar-operator/pkg/apis"
	appV1 "gitstar-operator/pkg/apis/app/v1"
	customV1 "gitstar-operator/pkg/apis/app/v1"
	"gitstar-operator/pkg/controller/gitstar"
)

var (
	log              = logf.Log.WithName("controller_gitstar")
	gitStarNameSpace = ""
	gitStarName      = ""
	gitHubOAuthToken = ""
	k8sClient        = newK8SClient()
)

const (
	GitHubOAuthTokenCMName      = "gitstar-github-token"
	GitHubOAuthTokenCMNameSpace = "default"
	GitHubOAuthTokenCMFileName  = "token"
)

func main() {
	err := InitEnv()
	if err != nil {
		log.Error(err, "")
		return
	}
	reqLogger := log.WithValues("Request.Namespace", gitStarNameSpace, "Request.Name", gitStarName)

	gitStar := &appV1.GitStar{}
	err = k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: gitStarNameSpace,
		Name:      gitStarName,
	}, gitStar)
	if err != nil {
		reqLogger.Error(err, "get gitStar apiObject failed! ")
		return
	}

	starNumber, err := GetStarOfRepo(gitStar)
	if err != nil {
		reqLogger.Error(err, "get star number of repo failed! ")
		gitStar.Status = customV1.GitStarStatus{
			StarNumber:   -1,
			UpdatedAt:    metav1.NewTime(time.Now()),
			FailedReason: err.Error(),
		}
	} else {
		gitStar.Status = customV1.GitStarStatus{
			StarNumber:   starNumber,
			UpdatedAt:    metav1.NewTime(time.Now()),
			FailedReason: "",
		}
	}

	err = UpdateGitStarObj(k8sClient, gitStar)
	if err != nil {
		log.Error(err, "get star number of repo failed! ")
		return
	}
	reqLogger.Info(fmt.Sprintf("update repo '%s', star number: '%d'", gitStar.Spec.RepoName, gitStar.Status.StarNumber))
	reqLogger.Info("update gitStar success \n")

}

func GetStarOfRepo(gitStar *customV1.GitStar) (int64, error) {
	split := strings.Split(gitStar.Spec.RepoName, "/")
	if len(split) != 2 {
		return -1, errors.New("The repo name is invalid, please check! ")
	}

	var c *github.Client
	if gitHubOAuthToken != "" {
		c = github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gitHubOAuthToken},
		)))
	} else {
		c = github.NewClient(nil)
	}

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

func InitEnv() error {
	// init gitStar target name
	name := os.Getenv(gitstar.ENVGitStarName)
	if name == "" {
		return errors.New("get gitStarName was empty form env")
	}
	gitStarName = name

	// init gitStar target namespace
	namespace := os.Getenv(gitstar.ENVGitStarNameSpace)
	if name == "" {
		return errors.New("get gitStarNameSpace was empty form env")
	}
	gitStarNameSpace = namespace

	// init github oauth token
	oAuthCM := &v1.ConfigMap{}
	err := k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: GitHubOAuthTokenCMNameSpace,
		Name:      GitHubOAuthTokenCMName,
	}, oAuthCM)
	if err != nil && k8serrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if data, ok := oAuthCM.Data[GitHubOAuthTokenCMFileName]; ok {
		gitHubOAuthToken = data
	} else {
		return nil
	}

	return nil
}
