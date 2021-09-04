FROM nginx
#FROM gcr.io/distroless/base-debian10
WORKDIR /app
COPY kfserving-inference-client ./
CMD ["./kfserving-inference-client", "-h"]
