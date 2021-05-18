FROM golang:1.16-alpine as build
ARG LINK
RUN apk add unzip
RUN adduser -D -g '' user

WORKDIR /var/lib/glove
RUN wget http://nlp.stanford.edu/data/glove.$LINK.zip
RUN unzip glove.$LINK.zip

ENV CGO_ENABLED=0
COPY . /go/src/github.com/mewil/gloves
WORKDIR /go/src/github.com/mewil/gloves
RUN go mod download
RUN go install .

FROM scratch AS gloves
ARG TOKENS
ARG DIMENSIONS
LABEL Author="Michael Wilson"
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/gloves /bin/gloves
COPY --from=build /var/lib/glove/glove.$TOKENS.$DIMENSIONS.txt /var/lib/glove/glove.$TOKENS.$DIMENSIONS.txt
ENV MODEL_FILE /var/lib/glove/glove.$TOKENS.$DIMENSIONS.txt
USER user
ENTRYPOINT ["/bin/gloves"]
EXPOSE 9090