# step 0
FROM kurisux/golang-base-compile-container:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN echo "export GOPROXY=https://goproxy.cn" >> /etc/profile

# install
RUN  source /etc/profile \
    && git clone git@github.com:Kurisu-public/gitstar-operator.git \
    && cd cmd/queryJob \
    && go build -o target \
    && mkdir -p /root/go/bin \
    && cp target /root/go/bin/target

# compression && packing
RUN strip --strip-unneeded target \
    && upx target \
    && chmod a+x ./target \
    && cp target /usr/local/bin

# step 1

FROM alpine:latest

COPY --from=0 /usr/local/bin/supergroup.mixin.one /usr/local/bin/