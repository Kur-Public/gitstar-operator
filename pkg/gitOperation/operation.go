package gitOperation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/client-go/kubernetes/scheme"

	"gitstar-operator/pkg/apis"
	appV1 "gitstar-operator/pkg/apis/app/v1"
	customV1 "gitstar-operator/pkg/apis/app/v1"
	"gitstar-operator/pkg/resource"
)

const (
	GitHubOAuthTokenCMName      = "gitstar-github-token"
	GitHubOAuthTokenCMNameSpace = "default"
	GitHubOAuthTokenCMFileName  = "token"
)

var (
	log              = logf.NewDelegatingLogger(zap.Logger())
	gitHubOAuthToken = ""
	k8sClient        = newK8SClient()
)

func Run(gitStarNameSpace, gitStarName string) {
	log.Info("start")

	err := InitEnv(&gitStarNameSpace, &gitStarName, &gitHubOAuthToken, k8sClient)
	if err != nil {
		log.Error(err, "")
		return
	}

	if k8sClient == nil {
		err := errors.New("k8sClient is nil")
		log.Error(err, "")
		return
	}

	// start
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

	starNumber, err := GetStarOfRepo(gitStar, gitHubOAuthToken)
	if err != nil {
		reqLogger.Error(err, "get star number of repo failed! ")
		if gitStar.Status.UpdatedAt.IsZero() {
			gitStar.Status.UpdatedAt = metav1.NewTime(time.Unix(0, 0))
		}
		gitStar.Status = customV1.GitStarStatus{
			StarNumber:   gitStar.Status.StarNumber,
			UpdatedAt:    gitStar.Status.UpdatedAt,
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
		log.Error(err, "update gitstar obj failed! ")
		return
	}
	reqLogger.Info(fmt.Sprintf("update repo '%s', star number: '%d'", gitStar.Spec.RepoName, gitStar.Status.StarNumber))
	reqLogger.Info("update gitStar success \n")

}

func newK8SClient() client.Client {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "get kubernetes config in cluster was failed! ")
		return nil
	}
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		log.Error(err, "get mapper was failed! ")
		return nil
	}
	scheme := scheme.Scheme
	apis.AddToScheme(scheme)
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		log.Error(err, "new kubernetes client for config was failed ! ")
		return nil
	}
	return c
}

func GetStarOfRepo(gitStar *customV1.GitStar, gitHubOAuthToken string) (int64, error) {
	split := strings.Split(gitStar.Spec.RepoName, "/")
	if len(split) != 2 {
		return -1, errors.New("The repo name is invalid, please check! ")
	}

	var c *github.Client
	if gitHubOAuthToken != "" {
		c = github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gitHubOAuthToken},
		)))
		log.Info("get githubOAuthToken," + gitHubOAuthToken)
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

func InitEnv(gitStarNameSpace, gitStarName, gitHubOAuthToken *string, c client.Client) error {
	// init gitStar target namespace
	if *gitStarNameSpace == "" {
		namespace := os.Getenv(resource.ENVGitStarNameSpace)
		if namespace == "" {
			return errors.New("get gitStarNameSpace was empty form env")
		}
		*gitStarNameSpace = namespace
	}

	// init gitStar target name
	if *gitStarName == "" {
		name := os.Getenv(resource.ENVGitStarName)
		if name == "" {
			return errors.New("get gitStarName was empty form env")
		}
		*gitStarName = name
	}

	// init github oauth token
	oAuthCM := &v1.ConfigMap{}
	err := c.Get(context.TODO(), types.NamespacedName{
		Namespace: GitHubOAuthTokenCMNameSpace,
		Name:      GitHubOAuthTokenCMName,
	}, oAuthCM)
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("not found oauth token cm !")
		return nil
	} else if err != nil {
		log.Error(err, "get oauth token cm failed !")
		return nil
	}

	if data, ok := oAuthCM.Data[GitHubOAuthTokenCMFileName]; ok && len(strings.TrimSpace(data)) == 40 {
		*gitHubOAuthToken = strings.TrimSpace(data)
	} else {
		return nil
	}

	return nil
}
