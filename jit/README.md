# Monkey JIT Compiler

Monkeyプログラミング言語用のJust-In-Time（JIT）コンパイラーの実装です。

## 概要

このJITコンパイラーは、Monkeyのバイトコード仮想マシンに統合され、頻繁に実行される関数（ホットスポット）を自動的に検出し、ネイティブコードにコンパイルして実行速度を向上させます。

## アーキテクチャ

```
[Monkey Source] → [Bytecode Compiler] → [VM with JIT Integration]
                                              ↓
                          [Profiler] → [Hot Spot Detection] → [JIT Compiler]
                                              ↓                     ↓
                          [Native Code Cache] ← [Memory Manager] ← [Native Code]
```

### 主要コンポーネント

1. **JIT Manager** (`jit.go`): JITシステムの中核制御
2. **Profiler** (`profiler.go`): 関数呼び出しプロファイリング
3. **Memory Manager** (`memory.go`): 実行可能メモリ管理
4. **Compiler** (`compiler.go`): バイトコード→ネイティブコード変換
5. **VM Integration** (`vm_integration.go`): VM統合レイヤー
6. **Extensions** (`extensions.go`): 関数JIT拡張メタデータ

## 基本的な使用方法

### 1. JITシステムの初期化

```go
package main

import "monkey/jit"

func main() {
    // JIT設定
    config := &jit.Config{
        Threshold:         100,   // 100回呼び出しでJITコンパイル
        OptimizationLevel: 1,     // 最適化レベル
        EnableProfiling:   true,  // プロファイリング有効
        MaxCodeCacheSize:  10 * 1024 * 1024, // 10MB キャッシュ
    }

    // JITインスタンス作成
    jitSystem := jit.NewJIT(config)
    defer jitSystem.Cleanup()
}
```

### 2. VM統合の使用

```go
import (
    "monkey/vm"
    "monkey/jit"
)

// 既存のVMをJIT対応にする
originalVM := vm.New(bytecode)
vmWithJIT := jit.NewVMWithJIT(originalVM, jitConfig)
defer vmWithJIT.GetIntegration().Cleanup()

// JIT統計情報の表示
vmWithJIT.GetIntegration().PrintStatistics()
```

### 3. 手動でのJITコンパイル

```go
// 関数の呼び出し記録
args := []object.Object{
    &object.Integer{Value: 42},
    &object.Integer{Value: 24},
}

jitSystem.RecordFunctionCall(function, args)

// JITコンパイル判定
if jitSystem.ShouldCompile(function) {
    nativeCode, err := jitSystem.CompileFunction(function)
    if err != nil {
        // エラーハンドリング
    }
}
```

## 設定オプション

### Config構造体

```go
type Config struct {
    Threshold         int  // JITコンパイル閾値（呼び出し回数）
    OptimizationLevel int  // 最適化レベル (0-3)
    EnableProfiling   bool // プロファイリング有効化
    MaxCodeCacheSize  int  // コードキャッシュ最大サイズ
}
```

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `Threshold` | 関数をJITコンパイルする呼び出し回数閾値 | 100 |
| `OptimizationLevel` | 最適化レベル（0=なし、3=最大） | 1 |
| `EnableProfiling` | プロファイリング機能の有効/無効 | true |
| `MaxCodeCacheSize` | ネイティブコードキャッシュの最大サイズ（bytes） | 10MB |

## 機能詳細

### ホットスポット検出

JITシステムは以下の方法でホットスポットを検出します：

1. **関数呼び出し回数**: 設定された閾値を超えた関数
2. **実行時間**: 累積実行時間が長い関数
3. **型プロファイル**: 引数と戻り値の型情報

### プロファイリング情報

各関数について以下の情報を収集します：

- 呼び出し回数
- 最後の呼び出し時刻
- 累積実行時間
- 引数の型分布
- 戻り値の型分布

### 最適化

現在実装されている最適化：

1. **基本算術演算**: 整数加算、減算、乗算
2. **型特殊化**: 頻繁に使用される型に対する特殊化
3. **定数伝播**: コンパイル時定数の最適化

## API リファレンス

### JIT構造体

#### メソッド

- `NewJIT(config *Config) *JIT`: 新しいJITインスタンスを作成
- `ShouldCompile(fn *object.CompiledFunction) bool`: JITコンパイル判定
- `CompileFunction(fn *object.CompiledFunction) (*NativeCode, error)`: 関数のJITコンパイル
- `RecordFunctionCall(fn *object.CompiledFunction, args []object.Object)`: 関数呼び出し記録
- `GetStatistics() Statistics`: 統計情報取得
- `PrintStatistics()`: 統計情報表示
- `Cleanup() error`: リソースクリーンアップ

### Profiler構造体

#### メソッド

- `RecordCall(fn *object.CompiledFunction, args []object.Object)`: 呼び出し記録
- `RecordReturn(fn *object.CompiledFunction, result object.Object)`: 戻り値記録
- `ShouldCompile(fn *object.CompiledFunction, threshold int) bool`: コンパイル判定
- `GetHotFunctions(limit int) []*FunctionProfile`: ホット関数リスト取得

### MemoryManager構造体

#### メソッド

- `AllocateExecutable(size int) (unsafe.Pointer, error)`: 実行可能メモリ確保
- `WriteCode(ptr unsafe.Pointer, code []byte) error`: コード書き込み
- `MakeExecutable(ptr unsafe.Pointer, size int) error`: 実行権限設定
- `Free(ptr unsafe.Pointer) error`: メモリ解放

## パフォーマンス

### ベンチマーク結果

```
BenchmarkJITCompilation-10       	  927855	      1367 ns/op	    1320 B/op	      17 allocs/op
BenchmarkProfilerRecording-10    	14205441	        81.14 ns/op	       0 B/op	       0 allocs/op
```

- JITコンパイル: 約1.4マイクロ秒/関数
- プロファイリング: 約81ナノ秒/呼び出し（ゼロアロケーション）

### 期待されるスピードアップ

- 算術集約型関数: 2-5x
- ループ処理: 3-10x
- 再帰関数: 1.5-3x

## 制限事項

### 現在の制限

1. **プラットフォーム**: x86-64 Linux/macOS のみ対応
2. **命令セット**: 基本算術演算のみ実装
3. **最適化**: 限定的な最適化パス
4. **デバッグ**: JITコードのデバッグサポートなし

### 既知の問題

1. メモリリークの可能性（実行可能メモリ）
2. 複雑な制御フローの未対応
3. 例外処理の未実装

## 開発者向け情報

### テスト実行

```bash
# 全テスト実行
go test ./jit

# ベンチマーク実行
go test ./jit -bench=. -benchmem

# 詳細テスト実行
go test ./jit -v
```

### デモ実行

```bash
# JITデモンストレーション
go run ./jit/demo.go
```

### 新しい最適化の追加

1. `compiler.go`の`generateNativeCode`関数を拡張
2. 新しいバイトコード命令のハンドラーを追加
3. テストケースを追加

### メモリ管理の改良

実際の本番環境では、以下の改良が必要です：

1. プラットフォーム固有のメモリ確保実装
2. セキュリティ考慮（W^X）
3. メモリリーク検出とガベージコレクション

## 今後の予定

### Phase 2 (次期実装)

1. **高度な最適化**
   - インライン化
   - ループ最適化
   - デッドコード除去

2. **型推論とタイプガード**
   - 動的型ガード挿入
   - 脱最適化メカニズム

3. **並列JITコンパイル**
   - バックグラウンドコンパイル
   - 段階的最適化

### Phase 3 (将来実装)

1. **投機的最適化**
   - プロファイルガイド最適化
   - 適応的最適化

2. **クロスプラットフォーム対応**
   - ARM64サポート
   - Windows対応

3. **デバッグサポート**
   - JITコードデバッグ情報
   - プロファイリングツール統合

## ライセンス

このJITコンパイラーは、Monkeyプロジェクトと同じライセンスの下で提供されます。

## 貢献

バグレポートや機能要求は、GitHubのIssueで受け付けています。プルリクエストも歓迎します。

### 貢献ガイドライン

1. 新機能はテストケースと共に提出
2. パフォーマンス関連の変更はベンチマークを含める
3. ドキュメントの更新を忘れずに
4. 既存のテストが全て通ることを確認

## サポート

技術的な質問や支援が必要な場合は、以下までお問い合わせください：

- GitHub Issues: バグレポートや機能要求
- Discord/Slack: リアルタイムサポート（コミュニティ）