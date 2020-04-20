operator-sdk build tmp-kurisux/gitstar-operator --image-build-args --no-cache
docker tag tmp-kurisux/gitstar-operator kurisux/gitstar-operator:latest
docker push kurisux/gitstar-operator:latest
docker rmi tmp-kurisux/gitstar-operator kurisux/gitstar-operator:latest