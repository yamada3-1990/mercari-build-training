# syntax=docker/dockerfile:1
# Python 3.7 イメージを使って当イメージを構築する。
FROM python:3.7-alpine

# 作業用ディレクトリを /code に指定する。
WORKDIR /code

# flaskコマンドにより用いられる環境変数を設定する。
ENV FLASK_APP=app.py
ENV FLASK_RUN_HOST=0.0.0.0

# gcc や依存パッケージをインストールする。
RUN apk add --no-cache gcc musl-dev linux-headers

# requirements.txtをコピーして Python 依存パッケージをインストールする。
COPY requirements.txt requirements.txt
RUN pip install -r requirements.txt

# イメージにメタデータを追加して、コンテナーがポート 5000 をリッスンするように記述する。
EXPOSE 5000

# このプロジェクトのカレントディレクトリ.を、イメージ内のワークディレクトリ.にコピーする。
COPY . .

# コンテナーに対するデフォルトのコマンドを flask run にする。
CMD ["flask", "run"]



