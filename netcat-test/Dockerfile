FROM alpine:3.14
RUN apk add --no-cache netcat-openbsd

COPY ./netcat-test/echo.sh /echo.sh
RUN chmod +x /echo.sh

ENTRYPOINT ["/echo.sh"]
