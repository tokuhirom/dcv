# dcv - Docker Container Viewer

DCV は Docker および Docker Compose で管理されているコンテナの状況を確認できる TUI (Terminal User Interface) ツールです｡

## 主な機能

- Docker で動作している全てのコンテナの一覧を表示
- Docker Compose で管理されているコンテナの一覧を表示
- 複数の Docker Compose プロジェクトの切り替え
- コンテナのログをリアルタイムで確認（最新1000行を初期表示､その後リアルタイム追従）
- Docker-in-Docker (dind) コンテナの中で動作するコンテナの管理
- vim 風のキーバインディング
- 実行コマンドの表示（デバッグに便利）

## Views

### Docker Container List View

`docker ps` の結果を見やすくテーブル形式で表示します。Docker Compose に限らず、全ての Docker コンテナを表示できます。

**キーバインド:**
- `↑`/`k`: 上へ移動
- `↓`/`j`: 下へ移動
- `Enter`: 選択したコンテナのログを表示
- `a`: 停止中のコンテナも含めて表示
- `r`: リストを更新
- `K`: コンテナを強制終了
- `S`: コンテナを停止
- `U`: コンテナを起動
- `R`: コンテナを再起動
- `D`: 停止中のコンテナを削除
- `q`/`Esc`: Docker Compose プロセスリストへ戻る

### Docker Compose Process List View

`docker compose ps` の結果を見やすくテーブル形式で表示します。

**キーバインド:**
- `↑`/`k`: 上へ移動
- `↓`/`j`: 下へ移動  
- `Enter`: 選択したコンテナのログを表示
- `d`: dind コンテナの中身を表示（dind コンテナ選択時のみ）
- `p`: 全ての Docker コンテナ一覧へ切り替え
- `P`: プロジェクト一覧へ切り替え
- `a`: 停止中のコンテナも含めて表示
- `r`: リストを更新
- `q`: 終了

### Log View

コンテナのログを表示します。初期表示で最新1000行を取得し、その後新しいログをリアルタイムで追加表示します。

**キーバインド:**
- `↑`/`k`: 上スクロール
- `↓`/`j`: 下スクロール
- `G`: 最下部へジャンプ
- `g`: 最上部へジャンプ
- `/`: 検索モード（未実装）
- `Esc`/`q`: Process List View へ戻る

### Docker-in-Docker Process List View

dind コンテナ内で動作しているコンテナの一覧を表示します。

**キーバインド:**
- `↑`/`k`: 上へ移動
- `↓`/`j`: 下へ移動
- `Enter`: 選択したコンテナのログを表示
- `r`: リストを更新
- `Esc`: Process List View へ戻る
- `q`: Process List View へ戻る

## 使い方

### オプション

```bash
dcv [-p <project>] [-f <compose-file>] [--projects]
```

- `-p <project>`: 指定した Docker Compose プロジェクトを表示
- `-f <compose-file>`: 指定した compose ファイルのプロジェクトを表示
- `--projects`: 起動時にプロジェクト一覧を表示

### 例

```bash
# 現在のディレクトリの Docker Compose プロジェクトを表示
dcv

# 特定のプロジェクトを表示
dcv -p myproject

# プロジェクト一覧から選択して起動
dcv --projects

# 全ての Docker コンテナを表示するには、起動後に 'p' キーを押す
```

## インストール

### Go install を使う場合

```bash
go install github.com/tokuhirom/dcv@latest
```

### ソースからビルドする場合

```bash
git clone https://github.com/tokuhirom/dcv.git
cd dcv
go build -o dcv
./dcv
```

## 要件

- Go 1.24.3 以上
- Docker および Docker Compose がインストールされていること
- ターミナルが TUI をサポートしていること

## 内部実装

- 言語: Go
- TUI フレームワーク: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- スタイリング: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- テスト: testify

### 特徴

- Model-View-Update (MVU) アーキテクチャを採用
- 非同期でログをストリーミング
- エラー時に実行したコマンドを表示してデバッグを容易に
- 包括的なユニットテスト

## デバッグ機能

- 実行されたコマンドがフッターに常時表示される
- エラー発生時は詳細なエラーメッセージとコマンドが表示される
- stderr 出力は `[STDERR]` プレフィックス付きで表示

## 開発

### テストの実行

```bash
go test ./...
```

### ビルド

```bash
go build -o dcv
```

### サンプル環境の起動

リポジトリには Docker Compose のサンプル環境が含まれています：

```bash
# サンプル環境を起動
docker compose up -d

# dcv でモニタリング
./dcv

# サンプル環境を停止
docker compose down
```

サンプル環境には以下が含まれます：
- `web`: Nginx サーバー
- `db`: PostgreSQL データベース
- `redis`: Redis キャッシュ
- `dind`: Docker-in-Docker（内部でテストコンテナが自動起動）

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

