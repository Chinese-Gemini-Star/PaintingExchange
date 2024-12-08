FROM gocv/opencv

LABEL maintainer="GeminiStar"

EXPOSE 8880

WORKDIR /app

COPY PaintingExchange .

CMD ["./PaintingExchange"]