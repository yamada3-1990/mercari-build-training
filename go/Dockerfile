FROM golang:1.24-alpine

# 作業ディレクトリを設定
WORKDIR /app
RUN addgroup --system mercari && adduser --system --ingroup mercari trainee

# 依存関係をコピー/ダウンロード
COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache gcc musl-dev

# アプリを動かすために必要なフォルダをコピー
COPY app ./app
COPY cmd ./cmd
COPY db ./db
COPY images ./images
RUN CGO_ENABLED=1 go build ./cmd/api

# 以降のCMD命令をtraineeユーザーで実行
RUN chown -R trainee:mercari /app
USER trainee

# /app/apiを実行
CMD ["./api"]