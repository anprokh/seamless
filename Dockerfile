FROM golang:1.19.1-bullseye
WORKDIR /
COPY . ./app-seamless
WORKDIR /app-seamless
RUN go build -o seamless
CMD [ "./seamless" ]