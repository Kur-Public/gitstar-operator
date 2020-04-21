# GitStar Operator

This's a simple dome of Kubernete Operator, you can submit some repo name to this operator , the Operator will be check the star number of this repo name .

[![asciicast](https://asciinema.org/a/322231.svg)](https://asciinema.org/a/322231)

## Deploy

### deploy crd && role && operator

```shell
$ git clone git@github.com:Kurisu-public/gitstar-operator.git
# deploy crd
$ kubectl apply -f deploy/crds/
# deploy role && role binding
$ kubectl apply -f deploy/role.yaml 
$ kubectl apply -f deploy/role_binding.yaml
# deploy SA
$ kubectl apply -f deploy/service_account.yaml
# deploy operator
$ kubectl apply -f deploy/operator.yaml
```

### (Optional) Configure OAuth Token Of GitHub

```shell
$ vim deploy/github_token.yaml 
# apiVersion: v1
# kind: ConfigMap
# metadata:
#   name: gitstar-github-token
# data:
#   token: |
#     <input your 'GitHub Personal access tokens' in here>      <----- modify here
#
$ kubectl apply -f deploy/github_token.yaml 
```

## Use

```shell
$ kubectl apply -f deploy/examples
$ watch kubectl get gstar -A
# enjoy :)
```


## LICENSE

Apache-2.0

First, let ’s check the current situation



begin