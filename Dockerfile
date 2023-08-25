FROM golang:1.20 AS build

COPY . /app

WORKDIR /app

RUN go get -v

RUN go build -v -o hf-provisioner-ec2

FROM golang:1.20 AS run

COPY --from=build /app/hf-provisioner-ec2 /hf-provisioner-ec2

WORKDIR /

CMD ["/hf-provisioner-ec2"]