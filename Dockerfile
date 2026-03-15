FROM gcr.io/distroless/static-debian12:nonroot

ARG TARGETARCH

COPY dist/osapi_linux_${TARGETARCH}*/osapi /usr/local/bin/osapi

ENTRYPOINT ["osapi"]
