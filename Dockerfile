FROM public.ecr.aws/docker/library/golang:1.20-alpine AS build

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o /app/hello

FROM public.ecr.aws/docker/library/alpine:3.14

WORKDIR /app
COPY --from=build /app/hello .

CMD ["./hello"]