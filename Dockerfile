FROM golang

WORKDIR /app

COPY . /app/

RUN go mod download && go mod tidy

RUN go build -o main main.go

ENV MONGODB_URI=mongodb://mongodb:27017

ENV PORT=8081

EXPOSE 8081

CMD [ "./main" ]