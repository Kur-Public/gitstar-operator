# step 0
FROM kurisux/golang-base-compile-container:latest

RUN echo "export GOPROXY=https://goproxy.cn" >> /etc/profile

# install
RUN source /etc/profile \
    && git clone https://github.com/Kurisu-public/gitstar-operator.git \
    && cd ./gitstar-operator/cmd/queryJob \
    && go build -o target \
    && mkdir -p /usr/local/bin \
    && cp target /usr/local/bin/target

# compression && packing
RUN strip --strip-unneeded /usr/local/bin/target \
    && upx /usr/local/bin/target \
    && chmod a+x /usr/local/bin/target

# step 1
FROM alpine:latest

COPY --from=0 /usr/local/bin/target /usr/local/bin/

CMD ["target"]