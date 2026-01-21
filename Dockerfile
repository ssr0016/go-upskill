FROM golang:1.25-alpine

# Install dependencies
RUN apk add --no-cache bash git netcat-openbsd

WORKDIR /app

# Copy go modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY . .

# Install Air (updated package path)
RUN go install github.com/air-verse/air@latest

# Add Go bin to PATH
ENV PATH="/go/bin:${PATH}"

# Expose app port
EXPOSE 8000

# Wait for DB and then run Air
CMD ["sh", "-c", "until nc -z $DB_HOST $DB_PORT; do echo 'Waiting for database...'; sleep 2; done && air"]
