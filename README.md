# dcv

DCV は docker-compose の状況を確認できるツールです｡

- docker-compose で起動しているアプリの一覧を表示します｡
    - ログを表示することができます.
- docker-compose で起動しているアプリのうち､dind っぽい image 名のものについては､さらにその中のイメージの一覧を見ることができます｡
- dind の中のログも見ることが出来ます｡
- 内部的には docker-compose コマンドを実行します｡

## Views

原則としてショートカットは vim 風に実装すること｡

### process list

`docker compose ps` の結果が一覧表示される｡
enter key を押すと､その container の一覧が見れる｡

d キーを押すと､dind として扱われ dind process list view に移動する｡

### dind process list view

dind process list view では､docker compose 上で稼働している dind のコンテナのうち一つを対象に､さらに docker ps を実行し､その結果を表示する｡
enter key を押すと､dind の中でうごくコンテナのログが表示可能｡

### log view

ログの表示時､`/` による検索や `G` による末尾への移動など､基本的な vim like なコマンドが実行可能｡

## How do I install

```bash
go install github.com/tokuhirom/dcv@latest
```

または､リポジトリをクローンしてビルド:

```bash
git clone https://github.com/tokuhirom/dcv.git
cd dcv
go build -o dcv
./dcv
```

## 内部実装

実装は golang を使用｡
TUI framework として tview を利用｡

## License

```
The MIT License (MIT)

Copyright © 2025 Tokuhiro Matsuno, https://64p.org/ <tokuhirom@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```

