FROM alpine:latest

ENV USER=docker
ENV GROUPNAME=$USER
ENV UID=1000
ENV GID=1000

COPY --chmod=755 entrypoint.sh /entrypoint.sh

# INSTALL PYTHON WITH TOOLS
RUN apk add --no-cache py3-pip python3 py3-numpy py3-pandas py3-scipy py3-sympy py3-pillow

# INSTALL GO
RUN apk add --no-cache make musl-dev go

# CONFIGURE GO
ENV GOPATH=/go
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin
RUN chown -R $UID:$GID ${GOPATH}

WORKDIR /
RUN addgroup --gid "$GID" "$GROUPNAME" &&  adduser --disabled-password --gecos "" --home "$(pwd)" --ingroup "$GROUPNAME" --no-create-home --uid "$UID" $USER
USER $USER

# NEED TO REDEFINE ENV BECAUSE WE SWITCHED THE USER
ENV GOROOT=/usr/lib/go
ENV GOPATH=/go
ENV PATH /go/bin:$PATH

ENTRYPOINT ["/entrypoint.sh"]
