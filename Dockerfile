# use official golang image
FROM golang:1.21.3-alpine3.18

# set working directory
WORKDIR /app

# Copy the source code 
COPY . .

# Download and install dependencies
RUN go get -d -v ./...

#Build the Go app
RUN go build -o catfact .

#EXPOSE THE PORT
EXPOSE 3000

# Run the excutable 
CMD [ "./catfact" ]