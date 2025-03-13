FROM node:22.12.0-alpine
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install --frozen-lockfile

RUN npm ci

RUN addgroup -S mercari && adduser -S trainee -G mercari
# USER trainee

# アプリを動かすために必要なフォルダをコピー
COPY . .

# コンテナのポートを公開
EXPOSE 3000

CMD ["npm", "start", "--", "--host", "0.0.0.0"]
