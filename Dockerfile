FROM bash
WORKDIR /app
COPY kfserving-inference-client ./
CMD ["./kfserving-inference-client", "-h"]
