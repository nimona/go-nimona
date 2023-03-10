FROM scratch

COPY nimona-pong /nimona-pong

ENTRYPOINT ["nimona-pong"]
