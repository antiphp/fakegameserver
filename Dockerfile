FROM gcr.io/distroless/static:nonroot

WORKDIR /app/
COPY fakegs /app/

ENTRYPOINT ["/app/fakegs"]
