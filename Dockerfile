FROM gcr.io/distroless/static:nonroot

WORKDIR /app/
COPY gameserver /app/

ENTRYPOINT ["/app/gameserver"]
