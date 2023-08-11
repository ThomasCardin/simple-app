FROM golang

WORKDIR /app

COPY . /app/

RUN go mod download && go mod tidy

RUN go build -o main main.go

ENV MONGO_USERNAME=$MONGODB_USERNAME
ENV MONGO_PASSWORD=$MONGODB_PASSWORD
ENV MONGO_HOST=$MONGODB_HOST

EXPOSE 8080

CMD [ "./main" ]