docker build ./cmd/queryJob/ -t tmp-kurisux/gitstar-queryjob --no-cache
docker tag tmp-kurisux/gitstar-queryjob kurisux/gitstar-queryjob:latest
docker push kurisux/gitstar-queryjob:latest
docker rmi  tmp-kurisux/gitstar-queryjob kurisux/gitstar-queryjob:latest