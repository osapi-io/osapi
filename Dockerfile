FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

ARG TARGETARCH

COPY dist/osapi_linux_${TARGETARCH}*/osapi /usr/local/bin/osapi

ENTRYPOINT ["osapi"]
