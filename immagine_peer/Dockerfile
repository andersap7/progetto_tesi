FROM golang:buster as peer-base
RUN apt-get install tzdata
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

FROM golang:buster as golang
RUN apt-get install bash \
	gcc \
	git \
	make \
	libc6

ADD . $GOPATH/src/github.com/hyperledger/fabric
WORKDIR $GOPATH/src/github.com/hyperledger/fabric

FROM golang as peer
ARG GO_TAGS=""
RUN make peer GO_TAGS=${GO_TAGS}

FROM peer-base

RUN apt-get update
RUN apt-get install jq wget gcc -y
COPY libtensorflow-cpu-linux-x86_64-2.5.0.tar.gz .
RUN tar -C /usr/local -xzf libtensorflow-cpu-linux-x86_64-2.5.0.tar.gz
RUN ldconfig


ENV FABRIC_CFG_PATH /etc/hyperledger/fabric
VOLUME /etc/hyperledger/fabric
VOLUME /var/hyperledger
COPY --from=peer /go/src/github.com/hyperledger/fabric/build/bin /usr/local/bin
COPY --from=peer /go/src/github.com/hyperledger/fabric/sampleconfig/msp ${FABRIC_CFG_PATH}/msp
COPY --from=peer /go/src/github.com/hyperledger/fabric/sampleconfig/core.yaml ${FABRIC_CFG_PATH}
EXPOSE 7051
CMD ["peer","node","start"]
